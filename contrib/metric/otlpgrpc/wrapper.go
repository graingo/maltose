package otlpgrpc

import (
	"context"
	"fmt"

	"github.com/mingzaily/maltose/os/mmetric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type counterWrapper struct {
	counter metric.Float64Counter
}

type upDownCounterWrapper struct {
	counter metric.Float64UpDownCounter
}

type histogramWrapper struct {
	histogram metric.Float64Histogram
}

func convertAttributes(attrs mmetric.AttributeMap) []attribute.KeyValue {
	if len(attrs) == 0 {
		return nil
	}
	kvs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		// 根据值的类型使用不同的属性创建方法
		switch val := v.(type) {
		case string:
			kvs = append(kvs, attribute.String(k, val))
		case int:
			kvs = append(kvs, attribute.Int(k, val))
		case int64:
			kvs = append(kvs, attribute.Int64(k, val))
		case float64:
			kvs = append(kvs, attribute.Float64(k, val))
		case bool:
			kvs = append(kvs, attribute.Bool(k, val))
		default:
			// 对于其他类型，转换为字符串
			kvs = append(kvs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
	return kvs
}

func (c *counterWrapper) Add(ctx context.Context, value float64, opts ...mmetric.Option) {
	var attrs []attribute.KeyValue
	if len(opts) > 0 {
		attrs = convertAttributes(opts[0].Attributes)
	}
	c.counter.Add(ctx, value, metric.WithAttributes(attrs...))
}

func (c *counterWrapper) Inc(ctx context.Context, opts ...mmetric.Option) {
	c.Add(ctx, 1, opts...)
}

func (c *upDownCounterWrapper) Add(ctx context.Context, value float64, opts ...mmetric.Option) {
	var attrs []attribute.KeyValue
	if len(opts) > 0 {
		attrs = convertAttributes(opts[0].Attributes)
	}
	c.counter.Add(ctx, value, metric.WithAttributes(attrs...))
}

func (c *upDownCounterWrapper) Inc(ctx context.Context, opts ...mmetric.Option) {
	c.Add(ctx, 1, opts...)
}

func (c *upDownCounterWrapper) Dec(ctx context.Context, opts ...mmetric.Option) {
	c.Add(ctx, -1, opts...)
}

func (h *histogramWrapper) Record(value float64, opts ...mmetric.Option) {
	var attrs []attribute.KeyValue
	if len(opts) > 0 {
		attrs = convertAttributes(opts[0].Attributes)
	}
	h.histogram.Record(context.Background(), value, metric.WithAttributes(attrs...))
}