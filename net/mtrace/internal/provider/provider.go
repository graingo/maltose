package provider

import (
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// TracerProvider 是 sdkTrace.TracerProvider 的包装器
type TracerProvider struct {
	*sdkTrace.TracerProvider
}

// New 返回一个新的配置好的 TracerProvider，默认没有 SpanProcessor
// 默认配置包括：
// - ParentBased(AlwaysSample) 采样器
// - 基于 unix nano 时间戳和随机数的 IDGenerator
// - resource.Default() Resource
// - 默认的 SpanLimits
func New() *TracerProvider {
	return &TracerProvider{
		TracerProvider: sdkTrace.NewTracerProvider(
			sdkTrace.WithIDGenerator(NewIDGenerator()),
		),
	}
}
