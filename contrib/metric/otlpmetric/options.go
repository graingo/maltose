package otlpmetric

import (
	"time"
)

// Option is a function type for configuration options.
type Option func(*options)

// options is the configuration options struct.
type options struct {
	// Service information.
	serviceName    string // Service name.
	serviceVersion string // Service version.
	environment    string // Environment name.

	// Export configuration.
	protocol       Protocol      // Protocol to use.
	exportInterval time.Duration // Export interval.
	timeout        time.Duration // Timeout.
	insecure       bool          // Whether to use an insecure connection.
	endpoint       string        // Endpoint address.
	urlPath        string        // URL path (for HTTP protocol only).

	// Resource attributes.
	resourceAttributes map[string]string // Custom resource attributes.
}

// defaultOptions returns the default options.
func defaultOptions() options {
	return options{
		serviceName:        "maltose-service",
		serviceVersion:     "1.0.0",
		environment:        "production",
		protocol:           ProtocolGRPC,
		exportInterval:     10 * time.Second,
		timeout:            10 * time.Second,
		insecure:           true,
		resourceAttributes: map[string]string{},
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

// WithEnvironment sets the environment.
func WithEnvironment(env string) Option {
	return func(o *options) {
		o.environment = env
	}
}

// WithProtocol sets the protocol.
func WithProtocol(protocol Protocol) Option {
	return func(o *options) {
		o.protocol = protocol
	}
}

// WithExportInterval sets the export interval.
func WithExportInterval(interval time.Duration) Option {
	return func(o *options) {
		o.exportInterval = interval
	}
}

// WithTimeout sets the timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// WithInsecure sets whether to use an insecure connection.
func WithInsecure(insecure bool) Option {
	return func(o *options) {
		o.insecure = insecure
	}
}

// WithURLPath sets the URL path (for HTTP protocol only).
func WithURLPath(path string) Option {
	return func(o *options) {
		o.urlPath = path
	}
}

// WithResourceAttribute adds a resource attribute.
func WithResourceAttribute(key, value string) Option {
	return func(o *options) {
		o.resourceAttributes[key] = value
	}
}
