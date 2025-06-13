package otlptrace

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/frame/m"
	"github.com/graingo/maltose/net/mipv4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
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
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var exporter *otlptrace.Exporter
	if o.protocol == ProtocolGRPC {
		exporter, err = createGRPCExporter(ctx, o)
	} else {
		exporter, err = createHTTPExporter(ctx, o)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
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

	// set global text map propagator, for passing trace context between different services
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		// implement W3C TraceContext specification, for passing trace ID and span ID
		propagation.TraceContext{},
		// implement W3C Baggage specification, for passing key-value pair metadata
		propagation.Baggage{},
	))

	// set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, o.timeout)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			m.Log().Errorf(ctx, "failed to shutdown tracer provider: %+v", err)
		} else {
			m.Log().Debug(ctx, "tracer provider shutdown successfully")
		}
	}, nil
}

func createResource(opts options) (*resource.Resource, error) {
	intranetIPArray, err := mipv4.GetIntranetIpArray()
	hostIP := "NoHostIpFound"
	if err != nil {
		// Do not return error, just log it.
		m.Log().Errorf(context.Background(), "failed to get intranet ip array: %v", err)
	} else if len(intranetIPArray) == 0 {
		if intranetIPArray, err = mipv4.GetIpArray(); err != nil {
			m.Log().Errorf(context.Background(), "failed to get ip array: %v", err)
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
