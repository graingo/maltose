// Package observability initializes Maltose trace and metric exporters as one lifecycle resource.
package observability

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/graingo/maltose/contrib/metric/otlpmetric"
	"github.com/graingo/maltose/contrib/trace/otlptrace"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mcfg"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	defaultConfigPattern   = "observability"
	defaultServiceName     = "maltose-service"
	defaultEnvironment     = "production"
	defaultExportTimeout   = 10 * time.Second
	defaultExportInterval  = 10 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

// Config defines shared resource information and signal-specific OTLP settings.
type Config struct {
	Enabled         bool              `mconv:"enabled"`
	ServiceName     string            `mconv:"service_name"`
	ServiceVersion  string            `mconv:"service_version"`
	Environment     string            `mconv:"environment"`
	Insecure        bool              `mconv:"insecure"`
	Attributes      map[string]string `mconv:"attributes"`
	ShutdownTimeout time.Duration     `mconv:"shutdown_timeout"`
	Trace           TraceConfig       `mconv:"trace"`
	Metric          MetricConfig      `mconv:"metric"`
}

// TraceConfig defines OTLP trace exporter settings.
type TraceConfig struct {
	Enabled     bool          `mconv:"enabled"`
	Endpoint    string        `mconv:"endpoint"`
	Protocol    string        `mconv:"protocol"`
	Timeout     time.Duration `mconv:"timeout"`
	URLPath     string        `mconv:"url_path"`
	SampleRatio float64       `mconv:"sample_ratio"`
}

// MetricConfig defines OTLP metric exporter settings.
type MetricConfig struct {
	Enabled        bool          `mconv:"enabled"`
	Endpoint       string        `mconv:"endpoint"`
	Protocol       string        `mconv:"protocol"`
	Timeout        time.Duration `mconv:"timeout"`
	URLPath        string        `mconv:"url_path"`
	ExportInterval time.Duration `mconv:"export_interval"`
}

// Provider owns initialized telemetry providers and shuts them down together.
type Provider struct {
	traceShutdown   func(context.Context)
	metricShutdown  func(context.Context) error
	shutdownTimeout time.Duration
	shutdownOnce    sync.Once
	shutdownErr     error
}

// FromConfig initializes observability from an mcfg node.
// It reads the "observability" node unless a custom pattern is supplied.
func FromConfig(ctx context.Context, config *mcfg.Config, pattern ...string) (*Provider, error) {
	if config == nil {
		return nil, merror.New("observability config source is required")
	}
	configPattern := defaultConfigPattern
	if len(pattern) > 0 && pattern[0] != "" {
		configPattern = pattern[0]
	}

	settings := Config{}
	if err := config.Struct(ctx, &settings, configPattern); err != nil {
		return nil, merror.Wrapf(err, "failed to load observability config from %q", configPattern)
	}
	return New(ctx, settings)
}

// New initializes the configured OTLP exporters.
// A disabled configuration returns a no-op Provider that remains safe to close.
func New(ctx context.Context, config Config) (*Provider, error) {
	config = normalizeConfig(config)
	provider := &Provider{shutdownTimeout: config.ShutdownTimeout}
	if !config.Enabled {
		return provider, nil
	}
	if !config.Trace.Enabled && !config.Metric.Enabled {
		return nil, merror.New("observability requires at least one enabled signal")
	}

	if config.Trace.Enabled {
		shutdown, err := initTrace(config)
		if err != nil {
			return nil, err
		}
		provider.traceShutdown = shutdown
	}
	if config.Metric.Enabled {
		shutdown, err := initMetric(config)
		if err != nil {
			provider.shutdownTrace(ctx)
			return nil, err
		}
		provider.metricShutdown = shutdown
	}
	return provider, nil
}

// Shutdown flushes metrics and traces exactly once.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}
	p.shutdownOnce.Do(func() {
		if p.metricShutdown != nil {
			p.shutdownErr = errors.Join(p.shutdownErr, p.metricShutdown(ctx))
		}
		p.shutdownTrace(ctx)
	})
	return p.shutdownErr
}

// Close implements io.Closer for use with m.WithCloser.
func (p *Provider) Close() error {
	if p == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.shutdownTimeout)
	defer cancel()
	return p.Shutdown(ctx)
}

func (p *Provider) shutdownTrace(ctx context.Context) {
	if p.traceShutdown != nil {
		p.traceShutdown(ctx)
	}
}

func normalizeConfig(config Config) Config {
	if config.ServiceName == "" {
		config.ServiceName = defaultServiceName
	}
	if config.Environment == "" {
		config.Environment = defaultEnvironment
	}
	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = defaultShutdownTimeout
	}
	if config.Trace.Protocol == "" {
		config.Trace.Protocol = string(otlptrace.ProtocolGRPC)
	}
	if config.Trace.Timeout <= 0 {
		config.Trace.Timeout = defaultExportTimeout
	}
	if config.Trace.SampleRatio <= 0 || config.Trace.SampleRatio > 1 {
		config.Trace.SampleRatio = 1
	}
	if config.Metric.Protocol == "" {
		config.Metric.Protocol = string(otlpmetric.ProtocolGRPC)
	}
	if config.Metric.Timeout <= 0 {
		config.Metric.Timeout = defaultExportTimeout
	}
	if config.Metric.ExportInterval <= 0 {
		config.Metric.ExportInterval = defaultExportInterval
	}
	return config
}

func initTrace(config Config) (func(context.Context), error) {
	protocol, err := traceProtocol(config.Trace.Protocol)
	if err != nil {
		return nil, err
	}
	if config.Trace.Endpoint == "" {
		return nil, merror.New("observability.trace.endpoint is required")
	}

	options := []otlptrace.Option{
		otlptrace.WithServiceName(config.ServiceName),
		otlptrace.WithServiceVersion(config.ServiceVersion),
		otlptrace.WithEnvironment(config.Environment),
		otlptrace.WithProtocol(protocol),
		otlptrace.WithTimeout(config.Trace.Timeout),
		otlptrace.WithInsecure(config.Insecure),
		otlptrace.WithURLPath(config.Trace.URLPath),
		otlptrace.WithSampler(sdktrace.TraceIDRatioBased(config.Trace.SampleRatio)),
	}
	for key, value := range config.Attributes {
		options = append(options, otlptrace.WithResourceAttribute(key, value))
	}
	shutdown, err := otlptrace.Init(config.Trace.Endpoint, options...)
	if err != nil {
		return nil, merror.Wrap(err, "failed to initialize OTLP trace exporter")
	}
	return shutdown, nil
}

func initMetric(config Config) (func(context.Context) error, error) {
	protocol, err := metricProtocol(config.Metric.Protocol)
	if err != nil {
		return nil, err
	}
	if config.Metric.Endpoint == "" {
		return nil, merror.New("observability.metric.endpoint is required")
	}

	options := []otlpmetric.Option{
		otlpmetric.WithServiceName(config.ServiceName),
		otlpmetric.WithServiceVersion(config.ServiceVersion),
		otlpmetric.WithEnvironment(config.Environment),
		otlpmetric.WithProtocol(protocol),
		otlpmetric.WithTimeout(config.Metric.Timeout),
		otlpmetric.WithInsecure(config.Insecure),
		otlpmetric.WithURLPath(config.Metric.URLPath),
		otlpmetric.WithExportInterval(config.Metric.ExportInterval),
	}
	for key, value := range config.Attributes {
		options = append(options, otlpmetric.WithResourceAttribute(key, value))
	}
	shutdown, err := otlpmetric.Init(config.Metric.Endpoint, options...)
	if err != nil {
		return nil, merror.Wrap(err, "failed to initialize OTLP metric exporter")
	}
	return shutdown, nil
}

func traceProtocol(protocol string) (otlptrace.Protocol, error) {
	switch protocol {
	case string(otlptrace.ProtocolGRPC):
		return otlptrace.ProtocolGRPC, nil
	case string(otlptrace.ProtocolHTTP):
		return otlptrace.ProtocolHTTP, nil
	default:
		return "", merror.Newf("unsupported observability.trace.protocol %q", protocol)
	}
}

func metricProtocol(protocol string) (otlpmetric.Protocol, error) {
	switch protocol {
	case string(otlpmetric.ProtocolGRPC):
		return otlpmetric.ProtocolGRPC, nil
	case string(otlpmetric.ProtocolHTTP):
		return otlpmetric.ProtocolHTTP, nil
	default:
		return "", merror.Newf("unsupported observability.metric.protocol %q", protocol)
	}
}
