package mmetric

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// NewMeterOption creates a new meter option
func NewMeterOption() MeterOption {
	return MeterOption{}
}

// WithInstrument sets the instrument name
func (o MeterOption) WithInstrument(instrument string) MeterOption {
	o.Instrument = instrument
	return o
}

// WithInstrumentVersion sets the instrument version
func (o MeterOption) WithInstrumentVersion(version string) MeterOption {
	o.InstrumentVersion = version
	return o
}

// WithMeterAttributes sets the attributes
func (o MeterOption) WithMeterAttributes(attrs Attributes) MeterOption {
	o.Attributes = attrs
	return o
}

// NewMetricOption creates a new metric option
func NewMetricOption() MetricOption {
	return MetricOption{}
}

// WithHelp sets the help information
func (o MetricOption) WithHelp(help string) MetricOption {
	o.Help = help
	return o
}

// WithUnit sets the unit
func (o MetricOption) WithUnit(unit string) MetricOption {
	o.Unit = unit
	return o
}

// WithMetricAttributes sets the attributes
func (o MetricOption) WithMetricAttributes(attrs Attributes) MetricOption {
	o.Attributes = attrs
	return o
}

// WithBuckets sets the histogram buckets
func (o MetricOption) WithBuckets(buckets []float64) MetricOption {
	o.Buckets = buckets
	return o
}

// WithAttributes creates an Option with the given attributes.
// This is a convenience function for creating attributes for a single metric observation.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return Option{
		Attributes: attrs,
	}
}

// GetMeter creates a Meter with the specified instrument name.
// It uses the global default provider.
func GetMeter(name string) Meter {
	p := GetProvider()
	return p.Meter(MeterOption{Instrument: name})
}

// NewCounter creates a new Counter metric.
// It uses the global default provider.
func NewCounter(name string, option MetricOption) (Counter, error) {
	meter := GetMeter(name)
	return meter.Counter(name, option)
}

// NewMustCounter creates a new Counter metric and panics if an error occurs.
func NewMustCounter(name string, option MetricOption) Counter {
	meter := GetMeter(name)
	return meter.MustCounter(name, option)
}

// NewUpDownCounter creates a new UpDownCounter metric.
func NewUpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	meter := GetMeter(name)
	return meter.UpDownCounter(name, option)
}

// NewMustUpDownCounter creates a new UpDownCounter metric and panics if an error occurs.
func NewMustUpDownCounter(name string, option MetricOption) UpDownCounter {
	meter := GetMeter(name)
	return meter.MustUpDownCounter(name, option)
}

// NewHistogram creates a new Histogram metric.
func NewHistogram(name string, option MetricOption) (Histogram, error) {
	meter := GetMeter(name)
	return meter.Histogram(name, option)
}

// NewMustHistogram creates a new Histogram metric and panics if an error occurs.
func NewMustHistogram(name string, option MetricOption) Histogram {
	meter := GetMeter(name)
	return meter.MustHistogram(name, option)
}

// Shutdown gracefully shuts down the global metric provider.
func Shutdown(ctx context.Context) error {
	p := GetProvider()
	return p.Shutdown(ctx)
}
