package otlptrace

import (
	"time"

	"go.opentelemetry.io/otel/sdk/trace"
)

// Protocol defines the protocol for the exporter.
type Protocol string

const (
	// ProtocolGRPC is the gRPC protocol.
	ProtocolGRPC Protocol = "grpc"
	// ProtocolHTTP is the HTTP protocol.
	ProtocolHTTP Protocol = "http"
)

// Option is a function that configures the tracer.
type Option func(*options)

// options holds the configuration for the tracer.
type options struct {
	// Service information.
	serviceName    string // Service name.
	serviceVersion string // Service version.
	environment    string // Environment name.

	// Export configuration.
	protocol    Protocol      // Protocol to use.
	timeout     time.Duration // Timeout for export.
	insecure    bool          // Whether to use an insecure connection.
	endpoint    string        // Exporter endpoint.
	urlPath     string        // URL path for HTTP exporter.
	compression int           // Compression level for HTTP exporter.

	// Resource attributes.
	resourceAttributes map[string]string // Custom resource attributes.

	// Sampler configuration.
	sampler trace.Sampler // Sampler to use for tracing.
}

// defaultOptions returns the default configuration options.
func defaultOptions() options {
	return options{
		serviceName:        "maltose-service",
		serviceVersion:     "1.0.0",
		environment:        "production",
		protocol:           ProtocolGRPC,
		timeout:            10 * time.Second,
		insecure:           true,
		compression:        1, // Default to gzip compression for HTTP
		resourceAttributes: make(map[string]string),
		sampler:            trace.AlwaysSample(),
	}
}

// WithServiceName sets the service name.
func WithServiceName(name string) Option {
	return func(o *options) {
		o.serviceName = name
	}
}

// WithServiceVersion sets the service version.
func WithServiceVersion(version string) Option {
	return func(o *options) {
		o.serviceVersion = version
	}
}

// WithEnvironment sets the deployment environment.
func WithEnvironment(env string) Option {
	return func(o *options) {
		o.environment = env
	}
}

// WithProtocol sets the exporter protocol.
func WithProtocol(protocol Protocol) Option {
	return func(o *options) {
		o.protocol = protocol
	}
}

// WithTimeout sets the export timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// WithInsecure enables or disables insecure connection.
func WithInsecure(insecure bool) Option {
	return func(o *options) {
		o.insecure = insecure
	}
}

// WithURLPath sets the URL path for the HTTP exporter.
func WithURLPath(path string) Option {
	return func(o *options) {
		o.urlPath = path
	}
}

// WithCompression sets the compression level for the HTTP exporter.
func WithCompression(level int) Option {
	return func(o *options) {
		o.compression = level
	}
}

// WithResourceAttribute adds a custom resource attribute.
func WithResourceAttribute(key, value string) Option {
	return func(o *options) {
		o.resourceAttributes[key] = value
	}
}

// WithSampler sets the sampler for tracing.
func WithSampler(sampler trace.Sampler) Option {
	return func(o *options) {
		o.sampler = sampler
	}
}
