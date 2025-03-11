package otlpgrpc

import (
	"context"
	"time"

	"github.com/graingo/maltose/frame/m"
	"github.com/graingo/maltose/net/mipv4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
// 参数:
//   - serviceName: 服务名称
//   - endpoint: OTLP 接收端点
//
// 返回:
//   - shutdown: 关闭函数，用于等待导出的 trace spans 上传完成，在程序结束时调用此函数可确保最近的追踪数据不会丢失
//   - err: 错误信息
func Init(serviceName, endpoint string) (func(ctx context.Context), error) {
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

	// 创建 OTLP gRPC 导出器
	// 负责将追踪数据通过 gRPC 协议发送到 OpenTelemetry Collector
	traceExp, err := otlptrace.New(ctx,
		otlptracegrpc.NewClient(
			// 使用非安全连接，适用于开发环境或内部网络
			otlptracegrpc.WithInsecure(),
			// 设置端点地址，指定收集器的主机名和端口
			otlptracegrpc.WithEndpoint(endpoint),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建资源，包含与追踪数据关联的元数据
	res, err := resource.New(ctx,
		// 从环境变量中获取资源信息
		resource.WithFromEnv(),
		// 添加进程信息
		resource.WithProcess(),
		// 添加 SDK 信息
		resource.WithTelemetrySDK(),
		// 添加主机信息
		resource.WithHost(),
		// 添加自定义属性
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.HostNameKey.String(hostIP),
			attribute.String(tracerHostnameTagKey, hostIP),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建追踪提供者
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

	// 设置全局文本映射传播器，用于在不同服务之间传递追踪上下文
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		// 实现 W3C TraceContext 规范，用于传递追踪 ID 和 span ID
		propagation.TraceContext{},
		// 实现 W3C Baggage 规范，用于传递键值对元数据
		propagation.Baggage{},
	))

	// 设置全局追踪提供者
	otel.SetTracerProvider(tracerProvider)

	// 返回关闭函数
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
