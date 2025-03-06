package otlpgrpc

import (
	"context"
	"time"

	"github.com/savorelle/maltose/net/mipv4"
	"github.com/savorelle/maltose/os/mmetric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	metricHostnameTagKey  = "hostname"
	defaultEndpoint       = "localhost:4317"
	defaultExportInterval = 10 * time.Second
	defaultTimeout        = 10 * time.Second
)

// Option 配置选项函数
type Option func(*config)

// 内部配置结构
type config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Insecure       bool
	ExportInterval time.Duration
	Timeout        time.Duration
}

// WithServiceVersion 设置服务版本
func WithServiceVersion(version string) Option {
	return func(c *config) {
		c.ServiceVersion = version
	}
}

// WithInsecure 设置不安全连接
func WithInsecure(insecure bool) Option {
	return func(c *config) {
		c.Insecure = insecure
	}
}

// WithExportInterval 设置导出间隔
func WithExportInterval(interval time.Duration) Option {
	return func(c *config) {
		c.ExportInterval = interval
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.Timeout = timeout
	}
}

// Init 初始化并注册 OTLP gRPC 指标提供者
// serviceName: 服务名称
// endpoint: 收集器地址，默认 localhost:4317
// env: 环境名称
// options: 可选配置项
func Init(serviceName, endpoint, env string, options ...Option) (func(ctx context.Context), error) {
	// 初始化默认配置
	cfg := &config{
		ServiceName:    serviceName,
		ServiceVersion: "v1.0.0",
		Environment:    env,
		Endpoint:       endpoint,
		Insecure:       true,
		ExportInterval: defaultExportInterval,
		Timeout:        defaultTimeout,
	}

	// 应用自定义选项
	for _, opt := range options {
		opt(cfg)
	}

	// 设置默认值
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}

	// 获取主机IP
	var (
		intranetIPArray, err = mipv4.GetIntranetIpArray()
		hostIP               = "NoHostIpFound"
	)
	if err != nil {
		return nil, err
	}
	if len(intranetIPArray) == 0 {
		if intranetIPArray, err = mipv4.GetIpArray(); err != nil {
			return nil, err
		}
	}
	if len(intranetIPArray) > 0 {
		hostIP = intranetIPArray[0]
	}

	ctx := context.Background()

	// 创建导出器
	exporterOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithTimeout(cfg.Timeout),
	}
	if cfg.Insecure {
		exporterOpts = append(exporterOpts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, err
	}

	// 创建资源
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
			semconv.HostNameKey.String(hostIP),
			attribute.String(metricHostnameTagKey, hostIP),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建指标提供者
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(cfg.ExportInterval),
			),
		),
	)

	// 设置全局提供者
	otel.SetMeterProvider(mp)
	mmetric.SetGlobalProvider(&provider{mp: mp})
	mmetric.SetEnabled(true)

	// 返回关闭函数
	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := mp.Shutdown(ctx); err != nil {
			// 这里可以添加日志记录
		}
	}, nil
}

// provider 实现 mmetric.Provider 接口
type provider struct {
	mp *sdkmetric.MeterProvider
}

// Meter 创建计量器
func (p *provider) Meter(option mmetric.MeterOption) mmetric.Meter {
	return newMeter(p.mp.Meter(
		option.Instrument,
		metric.WithInstrumentationVersion(option.InstrumentVersion),
	))
}

// Shutdown 关闭提供者
func (p *provider) Shutdown(ctx context.Context) error {
	return p.mp.Shutdown(ctx)
}
