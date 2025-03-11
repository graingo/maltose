package otlpmetric

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// 计数器包装器
type counterWrapper struct {
	counter metric.Float64Counter
}

// Add 添加值
func (c *counterWrapper) Add(ctx context.Context, value float64, opts ...mmetric.Option) {
	// 使用 metric.WithAttributes 将属性转换为 AddOption
	c.counter.Add(ctx, value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Inc 增加 1
func (c *counterWrapper) Inc(ctx context.Context, opts ...mmetric.Option) {
	// 使用 metric.WithAttributes 将属性转换为 AddOption
	c.counter.Add(ctx, 1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// 上下计数器包装器
type upDownCounterWrapper struct {
	counter metric.Float64UpDownCounter
}

// Add 添加值
func (c *upDownCounterWrapper) Add(ctx context.Context, value float64, opts ...mmetric.Option) {
	// 使用 metric.WithAttributes 将属性转换为 AddOption
	c.counter.Add(ctx, value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Inc 增加 1
func (c *upDownCounterWrapper) Inc(ctx context.Context, opts ...mmetric.Option) {
	// 使用 metric.WithAttributes 将属性转换为 AddOption
	c.counter.Add(ctx, 1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Dec 减少 1
func (c *upDownCounterWrapper) Dec(ctx context.Context, opts ...mmetric.Option) {
	// 使用 metric.WithAttributes 将属性转换为 AddOption
	c.counter.Add(ctx, -1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// 直方图包装器
type histogramWrapper struct {
	histogram metric.Float64Histogram
}

// Record 记录值
func (h *histogramWrapper) Record(value float64, opts ...mmetric.Option) {
	// 使用 metric.WithAttributes 将属性转换为 RecordOption
	h.histogram.Record(context.Background(), value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// 将 mmetric.Option 转换为 attribute.KeyValue
func optionsToAttributes(opts []mmetric.Option) []attribute.KeyValue {
	if len(opts) == 0 {
		return nil
	}

	// 合并所有属性
	mergedAttrs := make(mmetric.AttributeMap)
	for _, opt := range opts {
		mergedAttrs.Sets(opt.Attributes)
	}

	// 转换为 attribute.KeyValue
	attrs := make([]attribute.KeyValue, 0, len(mergedAttrs))
	for k, v := range mergedAttrs {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	return attrs
}
