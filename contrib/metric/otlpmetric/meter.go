package otlpmetric

import (
	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/metric"
)

// meterWrapper is a wrapper for a meter.
type meterWrapper struct {
	meter metric.Meter
}

// Counter creates a counter.
func (m *meterWrapper) Counter(name string, option mmetric.MetricOption) (mmetric.Counter, error) {
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

// MustCounter creates a counter and panics on error.
func (m *meterWrapper) MustCounter(name string, option mmetric.MetricOption) mmetric.Counter {
	counter, err := m.Counter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

// UpDownCounter creates an up-down counter.
func (m *meterWrapper) UpDownCounter(name string, option mmetric.MetricOption) (mmetric.UpDownCounter, error) {
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

// MustUpDownCounter creates an up-down counter and panics on error.
func (m *meterWrapper) MustUpDownCounter(name string, option mmetric.MetricOption) mmetric.UpDownCounter {
	counter, err := m.UpDownCounter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

// Histogram creates a histogram.
func (m *meterWrapper) Histogram(name string, option mmetric.MetricOption) (mmetric.Histogram, error) {
	var opts []metric.Float64HistogramOption

	// Set description and unit.
	opts = append(opts,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)

	// Set buckets (Note: OpenTelemetry v1.34.0 no longer supports explicit bucket boundaries).
	// Bucket configuration is now handled by the collector or backend system.

	// Create the histogram.
	histogram, err := m.meter.Float64Histogram(name, opts...)
	if err != nil {
		return nil, err
	}

	return &histogramWrapper{histogram: histogram}, nil
}

// MustHistogram creates a histogram and panics on error.
func (m *meterWrapper) MustHistogram(name string, option mmetric.MetricOption) mmetric.Histogram {
	histogram, err := m.Histogram(name, option)
	if err != nil {
		panic(err)
	}
	return histogram
}
