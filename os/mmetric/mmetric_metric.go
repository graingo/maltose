package mmetric

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// counterWrapper is a wrapper for an OpenTelemetry Counter.
type counterWrapper struct {
	counter metric.Float64Counter
}

// Add adds a value to the counter.
func (c *counterWrapper) Add(ctx context.Context, value float64, opts ...Option) {
	c.counter.Add(ctx, value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Inc increments the counter by 1.
func (c *counterWrapper) Inc(ctx context.Context, opts ...Option) {
	c.counter.Add(ctx, 1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// upDownCounterWrapper is a wrapper for an OpenTelemetry UpDownCounter.
type upDownCounterWrapper struct {
	counter metric.Float64UpDownCounter
}

// Add adds a value to the counter.
func (c *upDownCounterWrapper) Add(ctx context.Context, value float64, opts ...Option) {
	c.counter.Add(ctx, value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Inc increments the counter by 1.
func (c *upDownCounterWrapper) Inc(ctx context.Context, opts ...Option) {
	c.counter.Add(ctx, 1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// Dec decrements the counter by 1.
func (c *upDownCounterWrapper) Dec(ctx context.Context, opts ...Option) {
	c.counter.Add(ctx, -1, metric.WithAttributes(optionsToAttributes(opts)...))
}

// histogramWrapper is a wrapper for an OpenTelemetry Histogram.
type histogramWrapper struct {
	histogram metric.Float64Histogram
}

// Record records a value in the histogram.
func (h *histogramWrapper) Record(value float64, opts ...Option) {
	h.histogram.Record(context.Background(), value, metric.WithAttributes(optionsToAttributes(opts)...))
}

// optionsToAttributes converts a slice of Option into a slice of attribute.KeyValue.
func optionsToAttributes(opts []Option) []attribute.KeyValue {
	if len(opts) == 0 {
		return nil
	}
	var totalSize int
	for _, opt := range opts {
		totalSize += len(opt.Attributes)
	}
	if totalSize == 0 {
		return nil
	}
	attributes := make([]attribute.KeyValue, 0, totalSize)
	for _, opt := range opts {
		attributes = append(attributes, opt.Attributes...)
	}
	return attributes
}
