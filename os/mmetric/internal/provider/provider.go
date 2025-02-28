package provider

import (
	"context"
)

// Provider 默认的指标提供者
type Provider struct{}

// New 创建新的默认提供者
func New() *Provider {
	return &Provider{}
}

// Meter 创建计量器
func (p *Provider) Meter(option interface{}) interface{} {
	return &noopMeter{}
}

// Shutdown 关闭提供者
func (p *Provider) Shutdown(ctx context.Context) error {
	return nil
}

// 空操作计量器
type noopMeter struct{}

func (m *noopMeter) Counter(name string, option interface{}) (interface{}, error) {
	return &noopCounter{}, nil
}

func (m *noopMeter) MustCounter(name string, option interface{}) interface{} {
	return &noopCounter{}
}

func (m *noopMeter) UpDownCounter(name string, option interface{}) (interface{}, error) {
	return &noopUpDownCounter{}, nil
}

func (m *noopMeter) MustUpDownCounter(name string, option interface{}) interface{} {
	return &noopUpDownCounter{}
}

func (m *noopMeter) Histogram(name string, option interface{}) (interface{}, error) {
	return &noopHistogram{}, nil
}

func (m *noopMeter) MustHistogram(name string, option interface{}) interface{} {
	return &noopHistogram{}
}

// 空操作计数器
type noopCounter struct{}

func (c *noopCounter) Add(ctx context.Context, value float64, opts ...interface{}) {}

func (c *noopCounter) Inc(ctx context.Context, opts ...interface{}) {}

// 空操作上下计数器
type noopUpDownCounter struct{}

func (c *noopUpDownCounter) Add(ctx context.Context, value float64, opts ...interface{}) {}

func (c *noopUpDownCounter) Inc(ctx context.Context, opts ...interface{}) {}

func (c *noopUpDownCounter) Dec(ctx context.Context, opts ...interface{}) {}

// 空操作直方图
type noopHistogram struct{}

func (h *noopHistogram) Record(value float64, opts ...interface{}) {}
