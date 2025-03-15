package otlphttp

import (
	"context"
	"time"

	"github.com/graingo/maltose/frame/m"
	"github.com/graingo/maltose/net/mipv4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	tracerHostnameTagKey = "hostname"
)

// Init initializes and registers otlphttp to the global TracerProvider
//
// Parameters:
//   - serviceName: service name
//   - endpoint: OTLP receiver endpoint
//   - path: receiver path
//
// Returns:
//   - shutdown: shutdown function, used to wait for the exported trace spans to be uploaded, call this function at the end of the program to ensure that the recent trace data is not lost
//   - err: error information
func Init(serviceName, endpoint, path string) (func(ctx context.Context), error) {
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

	// create OTLP HTTP exporter
	// responsible for sending trace data through the HTTP protocol to the OpenTelemetry Collector
	traceExp, err := otlptrace.New(ctx, otlptracehttp.NewClient(
		// set endpoint address, specify the host name and port of the collector
		otlptracehttp.WithEndpoint(endpoint),
		// set URL path, specify the target path for trace data sending
		otlptracehttp.WithURLPath(path),
		// use insecure connection, suitable for development environment or internal network
		otlptracehttp.WithInsecure(),
		// set compression level, reduce network transmission data量
		otlptracehttp.WithCompression(1),
	))
	if err != nil {
		return nil, err
	}

	// create resource, containing metadata associated with trace data
	res, err := resource.New(ctx,
		// get resource information from environment variables
		resource.WithFromEnv(),
		// add process information
		resource.WithProcess(),
		// add SDK information
		resource.WithTelemetrySDK(),
		// add host information
		resource.WithHost(),
		// add custom attributes
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.HostNameKey.String(hostIP),
			attribute.String(tracerHostnameTagKey, hostIP),
		),
	)
	if err != nil {
		return nil, err
	}

	// create tracer provider
	tracerProvider := trace.NewTracerProvider(
		// sampler configuration: set to always sample
		// AlwaysSample will sample all trace data
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#AlwaysSample
		trace.WithSampler(trace.AlwaysSample()),
		// resource configuration: set the resource information associated with spans
		// the resource usually contains:
		// - service name
		// - host name
		// - IP address
		// - environment identifier
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithResource
		trace.WithResource(res),
		// Span processor configuration: use batch processing to handle spans
		// BatchSpanProcessor will:
		// - cache spans in memory
		// - batch send to the backend collector
		// - improve performance, reduce network requests
		// - asynchronous processing, not blocking the application
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithSpanProcessor
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExp)),
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

	// return shutdown function
	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err = tracerProvider.Shutdown(ctx); err != nil {
			m.Log().Errorf(ctx, "Shutdown tracerProvider failed: %+v", err)
		} else {
			m.Log().Debug(ctx, "Shutdown tracerProvider success")
		}
	}, nil
}
