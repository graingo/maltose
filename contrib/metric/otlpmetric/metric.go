package otlpmetric

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// counterWrapper is a wrapper for a counter.
type counterWrapper struct {
	counter metric.Float64Counter
}

// Add adds a value.
func (c *counterWrapper) Add(ctx context.Context, value float64, opts ...mmetric.Option) {
	// Convert attributes to AddOption using metric.WithAttributes.
	c.counter.Add(ctx, value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Inc increments by 1.
func (c *counterWrapper) Inc(ctx context.Context, opts ...mmetric.Option) {
	// Convert attributes to AddOption using metric.WithAttributes.
	c.counter.Add(ctx, 1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// upDownCounterWrapper is a wrapper for an up-down counter.
type upDownCounterWrapper struct {
	counter metric.Float64UpDownCounter
}

// Add adds a value.
func (c *upDownCounterWrapper) Add(ctx context.Context, value float64, opts ...mmetric.Option) {
	// Convert attributes to AddOption using metric.WithAttributes.
	c.counter.Add(ctx, value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Inc increments by 1.
func (c *upDownCounterWrapper) Inc(ctx context.Context, opts ...mmetric.Option) {
	// Convert attributes to AddOption using metric.WithAttributes.
	c.counter.Add(ctx, 1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Dec decrements by 1.
func (c *upDownCounterWrapper) Dec(ctx context.Context, opts ...mmetric.Option) {
	// Convert attributes to AddOption using metric.WithAttributes.
	c.counter.Add(ctx, -1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// histogramWrapper is a wrapper for a histogram.
type histogramWrapper struct {
	histogram metric.Float64Histogram
}

// Record records a value.
func (h *histogramWrapper) Record(value float64, opts ...mmetric.Option) {
	// Convert attributes to RecordOption using metric.WithAttributes.
	h.histogram.Record(context.Background(), value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// optionsToAttributes converts mmetric.Option to attribute.KeyValue.
func optionsToAttributes(opts []mmetric.Option) []attribute.KeyValue {
	if len(opts) == 0 {
		return nil
	}

	// Merge all attributes.
	mergedAttrs := make(mmetric.AttributeMap)
	for _, opt := range opts {
		mergedAttrs.Sets(opt.Attributes)
	}

	// Convert to attribute.KeyValue.
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
