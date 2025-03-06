package otlphttp

import (
	"context"
	"time"

	"github.com/savorelle/maltose/frame/m"
	"github.com/savorelle/maltose/net/mipv4"
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

// Init 初始化并注册 otlphttp 到全局 TracerProvider
//
// serviceName: 服务名称
// endpoint: OTLP 接收端点
// path: 接收路径
//
// 返回的 shutdown 函数用于等待导出的 trace spans 上传完成
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
			semconv.HostNameKey.String(hostIP),
			attribute.String(tracerHostnameTagKey, hostIP),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		// 采样器配置：设置为始终采样
		// AlwaysSample 会采样所有的追踪数据
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#AlwaysSample
		trace.WithSampler(trace.AlwaysSample()),
		// 资源配置：设置与 spans 关联的资源信息
		// 资源通常包含：
		// - 服务名称
		// - 主机名
		// - IP地址
		// - 环境标识
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#WithResource
		trace.WithResource(res),
		// Span处理器配置：使用批处理方式处理 spans
		// BatchSpanProcessor 会：
		// - 将 spans 缓存在内存中
		// - 批量发送到后端收集器
		// - 提高性能，减少网络请求
		// - 异步处理，不阻塞应用程序
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
			m.Log().Errorf(ctx, "Shutdown tracerProvider failed: %+v", err)
		} else {
			m.Log().Debug(ctx, "Shutdown tracerProvider success")
		}
	}, nil
}
