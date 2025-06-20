package otlpmetric

import (
	"context"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mipv4"
	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Init initializes OpenTelemetry metrics using the otlpmetric exporter.
// It configures and creates an exporter (gRPC or HTTP), a resource, and a periodic reader,
// then uses them to create and set a global mmetric.Provider.
func Init(endpoint string, opts ...Option) (func(context.Context) error, error) {
	// Apply user-provided options.
	o := defaultOptions()
	o.endpoint = endpoint
	for _, opt := range opts {
		opt(&o)
	}

	// Create a resource with service and host information.
	res, err := createResource(o)
	if err != nil {
		return nil, merror.Wrap(err, "failed to create resource")
	}

	// Create an exporter based on the configured protocol.
	var exporter sdkmetric.Exporter
	if o.protocol == ProtocolGRPC {
		exporter, err = createGRPCExporter(o)
	} else {
		exporter, err = createHTTPExporter(o)
	}
	if err != nil {
		return nil, merror.Wrap(err, "failed to create exporter")
	}

	// Create a periodic reader to export metrics at a fixed interval.
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		// Export interval configuration: Set the interval for exporting metric data.
		// WithInterval controls:
		// - The frequency of metric data collection and export.
		// - Shorter intervals provide more real-time data but increase system load.
		// - Longer intervals reduce system load but decrease data real-time performance.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#WithInterval
		sdkmetric.WithInterval(o.exportInterval),
	)

	// Use the new mmetric.NewProvider to create and wrap the OTel provider.
	provider := mmetric.NewProvider(
		// Resource configuration: Set resource information associated with metrics.
		// Resources usually include:
		// - Service name
		// - Hostname
		// - IP address
		// - Environment identifier
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#WithResource
		sdkmetric.WithResource(res),
		// Reader configuration: Set how to read and export metric data.
		// PeriodicReader will:
		// - Periodically collect metric data.
		// - Batch send to the backend collector.
		// - Process asynchronously, without blocking the application.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#WithReader
		sdkmetric.WithReader(reader),
	)

	// Set the created provider as the global provider for the mmetric package.
	mmetric.SetProvider(provider)

	// Return the shutdown function from our provider wrapper.
	return provider.Shutdown, nil
}

// createResource creates a resource.
func createResource(opts options) (*resource.Resource, error) {
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

	// Create a new resource.
	return resource.New(ctx,
		// Get resource information from environment variables.
		resource.WithFromEnv(),
		// Add process information.
		resource.WithProcess(),
		// Add SDK information.
		resource.WithTelemetrySDK(),
		// Add host information.
		resource.WithHost(),
		// Add custom attributes.
		resource.WithAttributes(
			semconv.ServiceNameKey.String(opts.serviceName),
			semconv.HostNameKey.String(hostIP),
		),
	)
}

// createGRPCExporter creates a gRPC exporter.
func createGRPCExporter(opts options) (sdkmetric.Exporter, error) {
	ctx := context.Background()

	// Set options.
	grpcOpts := []otlpmetricgrpc.Option{
		// Set the endpoint address.
		// WithEndpoint specifies:
		// - The hostname and port of the collector.
		// - It does not include the protocol prefix.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#WithEndpoint
		otlpmetricgrpc.WithEndpoint(opts.endpoint),
		// Set the timeout.
		// WithTimeout controls:
		// - The maximum waiting time for an export request.
		// - The request will be canceled if it exceeds this time.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#WithTimeout
		otlpmetricgrpc.WithTimeout(opts.timeout),
	}

	// Set security options.
	if opts.insecure {
		// Use an insecure connection.
		// WithInsecure sets:
		// - The use of a non-TLS connection.
		// - Suitable for development environments or internal networks.
		// - It is recommended to use TLS in production environments.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#WithInsecure
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithInsecure())
	}

	// Create the exporter.
	return otlpmetricgrpc.New(ctx, grpcOpts...)
}

// createHTTPExporter creates an HTTP exporter.
func createHTTPExporter(opts options) (sdkmetric.Exporter, error) {
	ctx := context.Background()

	// Set options.
	httpOpts := []otlpmetrichttp.Option{
		// Set the endpoint address.
		// WithEndpoint specifies:
		// - The hostname and port of the collector.
		// - It does not include the protocol prefix.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithEndpoint
		otlpmetrichttp.WithEndpoint(opts.endpoint),
		// Set the timeout.
		// WithTimeout controls:
		// - The maximum waiting time for an export request.
		// - The request will be canceled if it exceeds this time.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithTimeout
		otlpmetrichttp.WithTimeout(opts.timeout),
	}

	// Set the URL path.
	if opts.urlPath != "" {
		// Custom URL path.
		// WithURLPath sets:
		// - The target path for sending metric data.
		// - Defaults to "/v1/metrics".
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithURLPath
		httpOpts = append(httpOpts, otlpmetrichttp.WithURLPath(opts.urlPath))
	}

	// Set security options.
	if opts.insecure {
		// Use an insecure connection.
		// WithInsecure sets:
		// - The use of HTTP instead of HTTPS.
		// - Suitable for development environments or internal networks.
		// - It is recommended to use HTTPS in production environments.
		// See: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithInsecure
		httpOpts = append(httpOpts, otlpmetrichttp.WithInsecure())
	}

	// Create the exporter.
	return otlpmetrichttp.New(ctx, httpOpts...)
}
