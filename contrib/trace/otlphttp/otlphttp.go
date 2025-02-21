// Package otlphttp provides trace implementation using OpenTelemetry HTTP protocol.
package otlphttp

import (
	"context"
	"time"

	"github.com/mingzaily/maltose/frame/g"
	"go.opentelemetry.io/otel"
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

// Init 初始化并注册 otlphttp 到全局 TracerProvider
//
// serviceName: 服务名称
// endpoint: OTLP 接收端点
// path: 接收路径
//
// 返回的 shutdown 函数用于等待导出的 trace spans 上传完成
func Init(serviceName, endpoint, path string) (func(ctx context.Context), error) {
	ctx := context.Background()

	traceExp, err := otlptrace.New(ctx, otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithCompression(1),
	))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		// AlwaysSample is a sampler that samples every trace.
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#AlwaysSample
		trace.WithSampler(trace.AlwaysSample()),
		// WithResource returns a trace option that sets the resource to be associated with spans.
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithResource
		trace.WithResource(res),
		// WithSpanProcessor returns a trace option that sets the span processor to be used by the trace provider.
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithSpanProcessor
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExp)),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetTracerProvider(tracerProvider)

	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err = tracerProvider.Shutdown(ctx); err != nil {
			g.Log().Errorf(ctx, "Shutdown tracerProvider failed: %+v", err)
		} else {
			g.Log().Debug(ctx, "Shutdown tracerProvider success")
		}
	}, nil
}
