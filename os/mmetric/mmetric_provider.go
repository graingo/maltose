package mmetric

import "context"

// 添加默认的空操作提供者实现
type noopProvider struct{}

func newNoopProvider() *noopProvider {
	return &noopProvider{}
}

func (p *noopProvider) Meter(option MeterOption) Meter {
	return &noopMeter{}
}

func (p *noopProvider) Shutdown(ctx context.Context) error {
	return nil
}

type noopMeter struct{}

func (m *noopMeter) Counter(name string, option MetricOption) (Counter, error) {
	return &noopCounter{}, nil
}

func (m *noopMeter) MustCounter(name string, option MetricOption) Counter {
	return &noopCounter{}
}

func (m *noopMeter) UpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	return &noopUpDownCounter{}, nil
}

func (m *noopMeter) MustUpDownCounter(name string, option MetricOption) UpDownCounter {
	return &noopUpDownCounter{}
}

func (m *noopMeter) Histogram(name string, option MetricOption) (Histogram, error) {
	return &noopHistogram{}, nil
}

func (m *noopMeter) MustHistogram(name string, option MetricOption) Histogram {
	return &noopHistogram{}
}

type noopCounter struct{}

func (c *noopCounter) Add(ctx context.Context, value float64, opts ...Option) {}

func (c *noopCounter) Inc(ctx context.Context, opts ...Option) {}

type noopUpDownCounter struct{}

func (c *noopUpDownCounter) Add(ctx context.Context, value float64, opts ...Option) {}

func (c *noopUpDownCounter) Inc(ctx context.Context, opts ...Option) {}

func (c *noopUpDownCounter) Dec(ctx context.Context, opts ...Option) {}

type noopHistogram struct{}

func (h *noopHistogram) Record(value float64, opts ...Option) {}
