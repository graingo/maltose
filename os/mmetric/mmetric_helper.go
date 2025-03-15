package mmetric

import (
	"context"
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

// NewOption creates a new option
func NewOption() Option {
	return Option{
		Attributes: make(AttributeMap),
	}
}

// WithOptionAttributes sets the attributes
func (o Option) WithOptionAttributes(attrs AttributeMap) Option {
	o.Attributes = attrs
	return o
}

// GetMeter gets the meter with the specified name
func GetMeter(name string) Meter {
	return GetProvider().Meter(NewMeterOption().WithInstrument(name))
}

// NewCounter creates a counter
func NewCounter(name string, option MetricOption) (Counter, error) {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.Counter(name, option)
}

// NewMustCounter creates a counter, panics if it fails
func NewMustCounter(name string, option MetricOption) Counter {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.MustCounter(name, option)
}

// NewUpDownCounter creates an up-down counter
func NewUpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.UpDownCounter(name, option)
}

// NewMustUpDownCounter creates an up-down counter, panics if it fails
func NewMustUpDownCounter(name string, option MetricOption) UpDownCounter {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.MustUpDownCounter(name, option)
}

// NewHistogram creates a histogram
func NewHistogram(name string, option MetricOption) (Histogram, error) {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.Histogram(name, option)
}

// NewMustHistogram creates a histogram, panics if it fails
func NewMustHistogram(name string, option MetricOption) Histogram {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.MustHistogram(name, option)
}

// Shutdown shuts down the provider
func Shutdown(ctx context.Context) error {
	return GetProvider().Shutdown(ctx)
}
