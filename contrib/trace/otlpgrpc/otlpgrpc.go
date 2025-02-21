// Package otlpgrpc provides trace implementation using OpenTelemetry gRPC protocol.
package otlpgrpc

import (
	"context"
	"time"

	"github.com/mingzaily/maltose/frame/g"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	tracerHostnameTagKey = "hostname"
)

// Init 初始化并注册 otlpgrpc 到全局 TracerProvider
//
// serviceName: 服务名称
// endpoint: OTLP gRPC 接收端点
//
// 返回的 shutdown 函数用于等待导出的 trace spans 上传完成
func Init(serviceName, endpoint string) (func(ctx context.Context), error) {
	ctx := context.Background()

	traceExp, err := otlptrace.New(ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(endpoint), // Replace the otel Agent Addr with the access point obtained in the prerequisite。
		),
	)
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
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
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
