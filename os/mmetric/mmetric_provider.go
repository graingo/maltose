package mmetric

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// provider is a wrapper for a sdkmetric.MeterProvider that implements the Provider interface.
type provider struct {
	provider *sdkmetric.MeterProvider
}

// NewProvider creates a new Provider with the given OpenTelemetry SDK options.
// It wraps the standard OpenTelemetry MeterProvider in our custom interface.
func NewProvider(opts ...sdkmetric.Option) Provider {
	p := sdkmetric.NewMeterProvider(opts...)
	return &provider{provider: p}
}

// Meter implements the Provider interface.
func (p *provider) Meter(option MeterOption) Meter {
	meter := p.provider.Meter(
		option.Instrument,
		metric.WithInstrumentationVersion(option.InstrumentVersion),
	)
	return &meterWrapper{meter: meter}
}

// Shutdown implements the Provider interface.
func (p *provider) Shutdown(ctx context.Context) error {
	return p.provider.Shutdown(ctx)
}

// meterWrapper is a wrapper for an OpenTelemetry Meter that implements the Meter interface.
type meterWrapper struct {
	meter metric.Meter
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
	return &counterWrapper{counter: counter}, nil
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
	return &upDownCounterWrapper{counter: counter}, nil
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
	return &histogramWrapper{histogram: histogram}, nil
}

// MustHistogram creates a new Histogram, panicking on error.
func (m *meterWrapper) MustHistogram(name string, option MetricOption) Histogram {
	histogram, err := m.Histogram(name, option)
	if err != nil {
		panic(err)
	}
	return histogram
}
