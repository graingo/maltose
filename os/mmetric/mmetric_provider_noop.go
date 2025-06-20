package mmetric

import "context"

// noopProvider is a noop provider that does not perform any actual operations
// It is used when no actual provider is set or metric collection is disabled
type noopProvider struct{}

// newNoopProvider creates a new noop provider
func newNoopProvider() *noopProvider {
	return &noopProvider{}
}

// Meter implements the Provider interface, returns a noop meter
func (p *noopProvider) Meter(option MeterOption) Meter {
	return &noopMeter{}
}

// Shutdown implements the Provider interface, does not perform any operations
func (p *noopProvider) Shutdown(ctx context.Context) error {
	return nil
}

// noopMeter is a noop meter that does not perform any actual operations
type noopMeter struct{}

// Counter creates a noop counter
func (m *noopMeter) Counter(name string, option MetricOption) (Counter, error) {
	return &noopCounter{}, nil
}

// MustCounter creates a noop counter
func (m *noopMeter) MustCounter(name string, option MetricOption) Counter {
	return &noopCounter{}
}

// UpDownCounter creates a noop up-down counter
func (m *noopMeter) UpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	return &noopUpDownCounter{}, nil
}

// MustUpDownCounter creates a noop up-down counter
func (m *noopMeter) MustUpDownCounter(name string, option MetricOption) UpDownCounter {
	return &noopUpDownCounter{}
}

// Histogram creates a noop histogram
func (m *noopMeter) Histogram(name string, option MetricOption) (Histogram, error) {
	return &noopHistogram{}, nil
}

// MustHistogram creates a noop histogram
func (m *noopMeter) MustHistogram(name string, option MetricOption) Histogram {
	return &noopHistogram{}
}

// noopCounter is a noop counter that does not perform any actual operations
type noopCounter struct{}

// Add does not perform any operations
func (c *noopCounter) Add(ctx context.Context, value float64, opts ...Option) {}

// Inc does not perform any operations
func (c *noopCounter) Inc(ctx context.Context, opts ...Option) {}

// noopUpDownCounter is a noop up-down counter that does not perform any actual operations
type noopUpDownCounter struct{}

// Add does not perform any operations
func (c *noopUpDownCounter) Add(ctx context.Context, value float64, opts ...Option) {}

// Inc does not perform any operations
func (c *noopUpDownCounter) Inc(ctx context.Context, opts ...Option) {}

// Dec does not perform any operations
func (c *noopUpDownCounter) Dec(ctx context.Context, opts ...Option) {}

// noopHistogram is a noop histogram that does not perform any actual operations
type noopHistogram struct{}

// Record does not perform any operations
func (h *noopHistogram) Record(value float64, opts ...Option) {}
