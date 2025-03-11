package otlpmetric

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/net/mipv4"
	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// 协议类型
type Protocol string

const (
	ProtocolGRPC         Protocol = "grpc" // gRPC 协议
	ProtocolHTTP         Protocol = "http" // HTTP 协议
	metricHostnameTagKey          = "hostname"
)

// Init 初始化 OpenTelemetry 指标
//
// 参数:
//   - endpoint: 收集器端点地址，例如 "collector:4317"
//   - opts: 配置选项
//
// 返回:
//   - shutdown: 关闭函数，用于优雅关闭
//   - err: 错误信息
func Init(endpoint string, opts ...Option) (func(context.Context) error, error) {
	// 应用选项
	o := defaultOptions()
	o.endpoint = endpoint
	for _, opt := range opts {
		opt(&o)
	}

	// 创建资源
	res, err := createResource(o)
	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	// 根据协议创建导出器
	var exporter sdkmetric.Exporter
	if o.protocol == ProtocolGRPC {
		exporter, err = createGRPCExporter(o)
	} else {
		exporter, err = createHTTPExporter(o)
	}
	if err != nil {
		return nil, fmt.Errorf("创建导出器失败: %w", err)
	}

	// 创建读取器
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		// 导出间隔配置：设置指标数据导出的时间间隔
		// WithInterval 控制：
		// - 指标数据收集和导出的频率
		// - 较短的间隔提供更实时的数据，但增加系统负载
		// - 较长的间隔减少系统负载，但数据实时性降低
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#WithInterval
		sdkmetric.WithInterval(o.exportInterval),
	)

	// 创建指标提供者
	provider := sdkmetric.NewMeterProvider(
		// 资源配置：设置与指标关联的资源信息
		// 资源通常包含：
		// - 服务名称
		// - 主机名
		// - IP地址
		// - 环境标识
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#WithResource
		sdkmetric.WithResource(res),
		// 读取器配置：设置如何读取和导出指标数据
		// PeriodicReader 会：
		// - 定期收集指标数据
		// - 批量发送到后端收集器
		// - 异步处理，不阻塞应用程序
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#WithReader
		sdkmetric.WithReader(reader),
	)

	// 设置全局提供者
	otel.SetMeterProvider(provider)

	// 创建并设置我们的提供者包装器
	wrapper := newProviderWrapper(provider)
	mmetric.SetProvider(wrapper)

	// 返回关闭函数
	return provider.Shutdown, nil
}

// 创建资源
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

	// 创建资源
	return resource.New(ctx,
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
			semconv.ServiceNameKey.String(opts.serviceName),
			semconv.HostNameKey.String(hostIP),
			attribute.String(metricHostnameTagKey, hostIP),
		),
	)
}

// 创建 gRPC 导出器
func createGRPCExporter(opts options) (sdkmetric.Exporter, error) {
	ctx := context.Background()

	// 设置选项
	grpcOpts := []otlpmetricgrpc.Option{
		// 设置端点地址
		// WithEndpoint 指定：
		// - 收集器的主机名和端口
		// - 不包含协议前缀
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#WithEndpoint
		otlpmetricgrpc.WithEndpoint(opts.endpoint),
		// 设置超时时间
		// WithTimeout 控制：
		// - 导出请求的最大等待时间
		// - 超过此时间将取消请求
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#WithTimeout
		otlpmetricgrpc.WithTimeout(opts.timeout),
	}

	// 设置安全选项
	if opts.insecure {
		// 使用非安全连接
		// WithInsecure 设置：
		// - 使用非 TLS 连接
		// - 适用于开发环境或内部网络
		// - 生产环境建议使用 TLS
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#WithInsecure
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithInsecure())
	}

	// 创建导出器
	return otlpmetricgrpc.New(ctx, grpcOpts...)
}

// 创建 HTTP 导出器
func createHTTPExporter(opts options) (sdkmetric.Exporter, error) {
	ctx := context.Background()

	// 设置选项
	httpOpts := []otlpmetrichttp.Option{
		// 设置端点地址
		// WithEndpoint 指定：
		// - 收集器的主机名和端口
		// - 不包含协议前缀
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithEndpoint
		otlpmetrichttp.WithEndpoint(opts.endpoint),
		// 设置超时时间
		// WithTimeout 控制：
		// - 导出请求的最大等待时间
		// - 超过此时间将取消请求
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithTimeout
		otlpmetrichttp.WithTimeout(opts.timeout),
	}

	// 设置 URL 路径
	if opts.urlPath != "" {
		// 自定义 URL 路径
		// WithURLPath 设置：
		// - 指标数据发送的目标路径
		// - 默认为 "/v1/metrics"
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithURLPath
		httpOpts = append(httpOpts, otlpmetrichttp.WithURLPath(opts.urlPath))
	}

	// 设置安全选项
	if opts.insecure {
		// 使用非安全连接
		// WithInsecure 设置：
		// - 使用 HTTP 而非 HTTPS
		// - 适用于开发环境或内部网络
		// - 生产环境建议使用 HTTPS
		// see: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#WithInsecure
		httpOpts = append(httpOpts, otlpmetrichttp.WithInsecure())
	}

	// 创建导出器
	return otlpmetrichttp.New(ctx, httpOpts...)
}
