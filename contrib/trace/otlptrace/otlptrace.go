package otlptrace

import (
	"context"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/frame/m"
	"github.com/graingo/maltose/net/mipv4"
	"github.com/graingo/maltose/net/mtrace"
	"github.com/graingo/maltose/os/mlog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	tracerHostnameTagKey = "hostname"
)

// Init initializes an OTLP trace exporter and registers it as the global tracer provider.
func Init(endpoint string, opts ...Option) (func(context.Context), error) {
	o := defaultOptions()
	o.endpoint = endpoint
	for _, opt := range opts {
		opt(&o)
	}

	ctx := context.Background()

	res, err := createResource(o)
	if err != nil {
		return nil, merror.Wrap(err, "failed to create resource")
	}

	var exporter *otlptrace.Exporter
	if o.protocol == ProtocolGRPC {
		exporter, err = createGRPCExporter(ctx, o)
	} else {
		exporter, err = createHTTPExporter(ctx, o)
	}
	if err != nil {
		return nil, merror.Wrap(err, "failed to create exporter")
	}

	// Create a new tracer provider with a batch span processor and the given exporter.
	tracerProvider := mtrace.NewProvider(
		sdktrace.WithSampler(o.sampler),
		// resource configuration: set the resource information associated with spans.
		// The resource usually contains information about the entity producing telemetry,
		// such as the service name, version, and environment. This information is
		// associated with all spans created by the provider.
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithResource
		sdktrace.WithResource(res),
		// Span processor configuration: use batch processing to handle spans.
		// BatchSpanProcessor will:
		// - cache spans in memory
		// - batch send to the backend collector
		// - improve performance, reduce network requests
		// - asynchronous processing, not blocking the application
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithSpanProcessor
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)

	// Set the global tracer provider, so a tracer can be retrieved using otel.Tracer(name).
	mtrace.SetProvider(tracerProvider)

	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, o.timeout)
		defer cancel()
		if tp, ok := tracerProvider.(*sdktrace.TracerProvider); ok {
			log := m.Log().With(mlog.String(maltose.COMPONENT, "otlptrace"))
			if err := tp.Shutdown(ctx); err != nil {
				log.Errorf(ctx, err, "failed to shutdown tracer provider")
			} else {
				log.Infow(ctx, "tracer provider shutdown successfully")
			}
		}
	}, nil
}

func createResource(opts options) (*resource.Resource, error) {
	intranetIPArray, err := mipv4.GetIntranetIPArray()
	hostIP := "NoHostIpFound"
	if err != nil {
		return nil, err
	} else if len(intranetIPArray) == 0 {
		if intranetIPArray, err = mipv4.GetIPArray(); err != nil {
			return nil, err
		}
	}
	if len(intranetIPArray) > 0 {
		hostIP = intranetIPArray[0]
	}

	// create attributes, containing metadata associated with trace data
	// see: https://pkg.go.dev/go.opentelemetry.io/otel/attribute#KeyValue
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(opts.serviceName),
		semconv.ServiceVersionKey.String(opts.serviceVersion),
		semconv.DeploymentEnvironmentKey.String(opts.environment),
		semconv.HostNameKey.String(hostIP),
		attribute.String(tracerHostnameTagKey, hostIP),
	}
	for k, v := range opts.resourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	// create resource, containing metadata associated with trace data
	// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/resource#New
	return resource.New(context.Background(),
		// get resource information from environment variables
		resource.WithFromEnv(),
		// add process information
		resource.WithProcess(),
		// add SDK information
		resource.WithTelemetrySDK(),
		// add host information
		resource.WithHost(),
		// add custom attributes
		resource.WithAttributes(attrs...),
	)
}

func createGRPCExporter(ctx context.Context, opts options) (*otlptrace.Exporter, error) {
	// OTLP gRPC exporter
	// responsible for sending trace data through the gRPC protocol to the OpenTelemetry Collector
	// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc#NewClient
	grpcOpts := []otlptracegrpc.Option{
		// set endpoint address, specify the host name and port of the collector
		otlptracegrpc.WithEndpoint(opts.endpoint),
		// set timeout for the exporter
		otlptracegrpc.WithTimeout(opts.timeout),
	}
	if opts.insecure {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
	}
	client := otlptracegrpc.NewClient(grpcOpts...)
	return otlptrace.New(ctx, client)
}

func createHTTPExporter(ctx context.Context, opts options) (*otlptrace.Exporter, error) {
	httpOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(opts.endpoint),
		otlptracehttp.WithTimeout(opts.timeout),
	}
	if opts.insecure {
		httpOpts = append(httpOpts, otlptracehttp.WithInsecure())
	}
	if opts.urlPath != "" {
		httpOpts = append(httpOpts, otlptracehttp.WithURLPath(opts.urlPath))
	}
	if opts.compression > 0 {
		httpOpts = append(httpOpts, otlptracehttp.WithCompression(otlptracehttp.Compression(opts.compression)))
	}
	client := otlptracehttp.NewClient(httpOpts...)
	return otlptrace.New(ctx, client)
}
