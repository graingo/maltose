package mmetric

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// otelProvider is a wrapper for a metric.MeterProvider that implements the Provider interface.
type otelProvider struct {
	provider metric.MeterProvider
}

// Meter implements the Provider interface.
func (p *otelProvider) Meter(option MeterOption) Meter {
	if option.Instrument == "" {
		option.Instrument = defaultInstrument
	}
	meter := p.provider.Meter(
		option.Instrument,
		metric.WithInstrumentationVersion(option.InstrumentVersion),
	)
	return &meterWrapper{
		meter:      meter,
		attributes: option.Attributes,
	}
}

// Shutdown implements the Provider interface.
// Note that the default OpenTelemetry provider does not have a Shutdown method.
// This will panic if the underlying provider is not an SDK provider that has a Shutdown method.
func (p *otelProvider) Shutdown(ctx context.Context) error {
	if prov, ok := p.provider.(interface {
		Shutdown(context.Context) error
	}); ok {
		return prov.Shutdown(ctx)
	}
	return nil
}

// meterWrapper is a wrapper for an OpenTelemetry Meter that implements the Meter interface.
type meterWrapper struct {
	meter      metric.Meter
	attributes Attributes
}

// Counter creates a new Counter metric instrument.
func (m *meterWrapper) Counter(name string, option MetricOption) (Counter, error) {
	counter, err := m.meter.Float64Counter(
		name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &counterWrapper{
		counter:    counter,
		attributes: append(m.attributes, option.Attributes...),
	}, nil
}

// MustCounter creates a new Counter, panicking on error.
func (m *meterWrapper) MustCounter(name string, option MetricOption) Counter {
	counter, err := m.Counter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

// UpDownCounter creates a new UpDownCounter metric instrument.
func (m *meterWrapper) UpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	counter, err := m.meter.Float64UpDownCounter(
		name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &upDownCounterWrapper{
		counter:    counter,
		attributes: append(m.attributes, option.Attributes...),
	}, nil
}

// MustUpDownCounter creates a new UpDownCounter, panicking on error.
func (m *meterWrapper) MustUpDownCounter(name string, option MetricOption) UpDownCounter {
	counter, err := m.UpDownCounter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

// Histogram creates a new Histogram metric instrument.
func (m *meterWrapper) Histogram(name string, option MetricOption) (Histogram, error) {
	histogram, err := m.meter.Float64Histogram(
		name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
		metric.WithExplicitBucketBoundaries(option.Buckets...),
	)
	if err != nil {
		return nil, err
	}
	return &histogramWrapper{
		histogram:  histogram,
		attributes: append(m.attributes, option.Attributes...),
	}, nil
}

// MustHistogram creates a new Histogram, panicking on error.
func (m *meterWrapper) MustHistogram(name string, option MetricOption) Histogram {
	histogram, err := m.Histogram(name, option)
	if err != nil {
		panic(err)
	}
	return histogram
}

// counterWrapper is a wrapper for an OpenTelemetry Counter.
type counterWrapper struct {
	counter    metric.Float64Counter
	attributes Attributes
}

// Add adds a value to the counter.
func (c *counterWrapper) Add(ctx context.Context, value float64, opts ...Option) {
	c.counter.Add(ctx, value, metric.WithAttributes(
		append(c.attributes, optionsToAttributes(opts)...)...,
	))
}

// Inc increments the counter by 1.
func (c *counterWrapper) Inc(ctx context.Context, opts ...Option) {
	c.counter.Add(ctx, 1, metric.WithAttributes(
		append(c.attributes, optionsToAttributes(opts)...)...,
	))
}

// upDownCounterWrapper is a wrapper for an OpenTelemetry UpDownCounter.
type upDownCounterWrapper struct {
	counter    metric.Float64UpDownCounter
	attributes Attributes
}

// Add adds a value to the counter.
func (c *upDownCounterWrapper) Add(ctx context.Context, value float64, opts ...Option) {
	c.counter.Add(ctx, value, metric.WithAttributes(
		append(c.attributes, optionsToAttributes(opts)...)...,
	))
}

// Inc increments the counter by 1.
func (c *upDownCounterWrapper) Inc(ctx context.Context, opts ...Option) {
	c.counter.Add(ctx, 1, metric.WithAttributes(
		append(c.attributes, optionsToAttributes(opts)...)...,
	))
}

// Dec decrements the counter by 1.
func (c *upDownCounterWrapper) Dec(ctx context.Context, opts ...Option) {
	c.counter.Add(ctx, -1, metric.WithAttributes(
		append(c.attributes, optionsToAttributes(opts)...)...,
	))
}

// histogramWrapper is a wrapper for an OpenTelemetry Histogram.
type histogramWrapper struct {
	histogram  metric.Float64Histogram
	attributes Attributes
}

// Record records a value in the histogram.
func (h *histogramWrapper) Record(value float64, opts ...Option) {
	h.histogram.Record(context.Background(), value, metric.WithAttributes(
		append(h.attributes, optionsToAttributes(opts)...)...,
	))
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
