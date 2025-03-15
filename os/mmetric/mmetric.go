// Package mmetric 提供指标收集和监控的抽象接口
package mmetric

import (
	"context"
)

// Provider is the interface for metric providers
type Provider interface {
	Meter(option MeterOption) Meter     // Meter creates a meter
	Shutdown(ctx context.Context) error // Shutdown shuts down the provider
}

// Meter is the interface for meters
type Meter interface {
	Counter(name string, option MetricOption) (Counter, error)             // Counter creates a counter
	MustCounter(name string, option MetricOption) Counter                  // MustCounter creates a counter, panics if creation fails
	UpDownCounter(name string, option MetricOption) (UpDownCounter, error) // UpDownCounter creates an up-down counter
	MustUpDownCounter(name string, option MetricOption) UpDownCounter      // MustUpDownCounter creates an up-down counter, panics if creation fails
	Histogram(name string, option MetricOption) (Histogram, error)         // Histogram creates a histogram
	MustHistogram(name string, option MetricOption) Histogram              // MustHistogram creates a histogram, panics if creation fails
}

// Counter is the interface for counters
type Counter interface {
	Add(ctx context.Context, value float64, opts ...Option) // Add adds a specified value
	Inc(ctx context.Context, opts ...Option)                // Inc increments the counter by 1
}

// UpDownCounter is the interface for up-down counters
type UpDownCounter interface {
	Add(ctx context.Context, value float64, opts ...Option) // Add adds a specified value (can be negative)
	Inc(ctx context.Context, opts ...Option)                // Inc increments the counter by 1
	Dec(ctx context.Context, opts ...Option)                // Dec decrements the counter by 1
}

// Histogram is the interface for histograms
type Histogram interface {
	Record(value float64, opts ...Option) // Record records a value
}

// MeterOption is the option for meters
type MeterOption struct {
	Instrument        string     // Instrument name
	InstrumentVersion string     // Instrument version
	Attributes        Attributes // Attributes
}

// MetricOption is the option for metrics
type MetricOption struct {
	Help       string     // Help information
	Unit       string     // Unit
	Attributes Attributes // Attributes
	Buckets    []float64  // Buckets (may not be supported by all implementations)
}

// Option is the option for recording metrics
type Option struct {
	Attributes AttributeMap // Attributes map
}

// Attributes is a list of attributes, key-value pairs
type Attributes map[string]string

// AttributeMap is a map of attributes, supports multiple types of values
type AttributeMap map[string]interface{}

// Sets sets multiple attributes
func (m AttributeMap) Sets(attrs AttributeMap) {
	for k, v := range attrs {
		m[k] = v
	}
}

// Pick picks specific attributes
func (m AttributeMap) Pick(keys ...string) AttributeMap {
	result := make(AttributeMap)
	for _, key := range keys {
		if v, ok := m[key]; ok {
			result[key] = v
		}
	}
	return result
}

// Global variables
var (
	enabled                  = true              // Whether to enable metric collection
	defaultProvider          = newNoopProvider() // Default noop provider
	activeProvider  Provider = defaultProvider   // Current active provider, modified to Provider interface type
)

// IsEnabled checks if metric collection is enabled
func IsEnabled() bool {
	return enabled
}

// SetEnabled sets whether metric collection is enabled
func SetEnabled(e bool) {
	enabled = e
}

// SetProvider sets the active provider
func SetProvider(provider Provider) {
	activeProvider = provider
	enabled = true
}

// GetProvider gets the current active provider
func GetProvider() Provider {
	if !enabled {
		return defaultProvider
	}
	return activeProvider
}
