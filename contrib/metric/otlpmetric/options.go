package otlpmetric

import (
	"time"
)

// Option 配置选项函数类型
type Option func(*options)

// options 配置选项结构体
type options struct {
	// 服务信息
	serviceName    string // 服务名称
	serviceVersion string // 服务版本
	environment    string // 环境名称

	// 导出配置
	protocol       Protocol      // 使用的协议
	exportInterval time.Duration // 导出间隔
	timeout        time.Duration // 超时时间
	insecure       bool          // 是否使用非安全连接
	endpoint       string        // 端点地址
	urlPath        string        // URL 路径（仅用于 HTTP 协议）

	// 资源属性
	resourceAttributes map[string]string // 自定义资源属性
}

// defaultOptions 返回默认选项
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

// WithServiceName 设置服务名称
func WithServiceName(name string) Option {
	return func(o *options) {
		o.serviceName = name
	}
}

// WithServiceVersion 设置服务版本
func WithServiceVersion(version string) Option {
	return func(o *options) {
		o.serviceVersion = version
	}
}

// WithEnvironment 设置环境
func WithEnvironment(env string) Option {
	return func(o *options) {
		o.environment = env
	}
}

// WithProtocol 设置协议
func WithProtocol(protocol Protocol) Option {
	return func(o *options) {
		o.protocol = protocol
	}
}

// WithExportInterval 设置导出间隔
func WithExportInterval(interval time.Duration) Option {
	return func(o *options) {
		o.exportInterval = interval
	}
}

// WithTimeout 设置超时
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// WithInsecure 设置是否使用非安全连接
func WithInsecure(insecure bool) Option {
	return func(o *options) {
		o.insecure = insecure
	}
}

// WithURLPath 设置 URL 路径（仅用于 HTTP 协议）
func WithURLPath(path string) Option {
	return func(o *options) {
		o.urlPath = path
	}
}

// WithResourceAttribute 添加资源属性
func WithResourceAttribute(key, value string) Option {
	return func(o *options) {
		o.resourceAttributes[key] = value
	}
}
