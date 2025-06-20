package mmetric

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

const (
	// DefaultInstrument an instrument name for the default meter.
	DefaultInstrument = "github.com/graingo/maltose/os/mmetric"
)

var (
	// defaultProvider is the default provider for mmetric.
	// It is a no-op provider by default, which means that no metrics will be collected
	// unless a real provider is set using the SetProvider function.
	defaultProvider Provider = newNoopProvider()
)

// Attributes is a slice of attribute.KeyValue. It is used to add metadata to metrics.
// Using attribute.KeyValue allows for strongly-typed attributes, which is recommended
// by OpenTelemetry for better performance and correctness.
//
// Example:
//
//	attrs := mmetric.Attributes{
//		attribute.String("request.method", "GET"),
//		attribute.Int("response.status_code", 200),
//	}
type Attributes []attribute.KeyValue

// MeterOption is the option for creating a new meter. A meter is responsible for creating
// instruments (e.g., counters, histograms).
type MeterOption struct {
	// Instrument is the name of the instrumentation library.
	Instrument string
	// InstrumentVersion is the version of the instrumentation library.
	InstrumentVersion string
	// Attributes is a list of attributes that will be attached to all metrics created by this meter.
	Attributes Attributes
}

// MetricOption is the option for creating a new metric instrument (e.g., Counter, Histogram).
type MetricOption struct {
	// Help provides a brief description of the metric. It is used by some backends
	// to display help text in GUIs.
	Help string
	// Unit specifies the unit of the metric. It should follow the UCUM standard.
	// See: https://unitsofmeasure.org/ucum.html
	Unit string
	// Attributes is a list of attributes that will be attached to the metric.
	Attributes Attributes
	// Buckets defines the bucket boundaries for a Histogram. If not set, a default
	// set of buckets will be used by the provider.
	// This is only applicable to Histogram metrics.
	Buckets []float64
}

// Option is the option for a single metric operation, like Add or Inc.
type Option struct {
	// Attributes is a list of attributes that will be attached to this specific metric observation.
	// These attributes are combined with the attributes from the Meter and the Instrument.
	Attributes Attributes
}

// Provider is the interface for a metric provider. It is responsible for creating meters.
// This is an abstraction over OpenTelemetry's MeterProvider.
type Provider interface {
	// Meter creates a new meter with the given options.
	Meter(option MeterOption) Meter
	// Shutdown gracefully shuts down the provider, ensuring all buffered metrics are exported.
	Shutdown(ctx context.Context) error
}

// Meter is the interface for a metric meter. It is responsible for creating instruments.
// This is an abstraction over OpenTelemetry's Meter.
type Meter interface {
	// Counter creates a new counter metric. A counter is a metric that only goes up.
	Counter(name string, option MetricOption) (Counter, error)
	// MustCounter is like Counter but panics on error.
	MustCounter(name string, option MetricOption) Counter
	// UpDownCounter creates a new up-down counter. This metric can go up and down.
	UpDownCounter(name string, option MetricOption) (UpDownCounter, error)
	// MustUpDownCounter is like UpDownCounter but panics on error.
	MustUpDownCounter(name string, option MetricOption) UpDownCounter
	// Histogram creates a new histogram metric. Histograms are used to measure the
	// distribution of a set of values.
	Histogram(name string, option MetricOption) (Histogram, error)
	// MustHistogram is like Histogram but panics on error.
	MustHistogram(name string, option MetricOption) Histogram
}

// Counter is an interface for a counter metric.
type Counter interface {
	// Add adds a value to the counter. The value must be non-negative.
	Add(ctx context.Context, value float64, opts ...Option)
	// Inc increments the counter by 1.
	Inc(ctx context.Context, opts ...Option)
}

// UpDownCounter is an interface for an up-down counter metric.
type UpDownCounter interface {
	// Add adds a value to the counter. The value can be positive or negative.
	Add(ctx context.Context, value float64, opts ...Option)
	// Inc increments the counter by 1.
	Inc(ctx context.Context, opts ...Option)
	// Dec decrements the counter by 1.
	Dec(ctx context.Context, opts ...Option)
}

// Histogram is an interface for a histogram metric.
type Histogram interface {
	// Record records a value in the histogram.
	Record(value float64, opts ...Option)
}

// SetProvider sets the global metric provider.
// This should be called once at the beginning of the application.
func SetProvider(p Provider) {
	defaultProvider = p
}

// GetProvider returns the global metric provider.
func GetProvider() Provider {
	return defaultProvider
}
