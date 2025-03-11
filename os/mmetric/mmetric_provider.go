// Package mmetric 提供指标收集和监控的提供者实现
package mmetric

import "context"

// noopProvider 是一个空操作提供者，不执行任何实际操作
// 用于在未设置实际提供者或禁用指标收集时使用
type noopProvider struct{}

// newNoopProvider 创建一个新的空操作提供者
func newNoopProvider() *noopProvider {
	return &noopProvider{}
}

// Meter 实现 Provider 接口，返回空操作计量器
func (p *noopProvider) Meter(option MeterOption) Meter {
	return &noopMeter{}
}

// Shutdown 实现 Provider 接口，不执行任何操作
func (p *noopProvider) Shutdown(ctx context.Context) error {
	return nil
}

// noopMeter 是一个空操作计量器，不执行任何实际操作
type noopMeter struct{}

// Counter 创建空操作计数器
func (m *noopMeter) Counter(name string, option MetricOption) (Counter, error) {
	return &noopCounter{}, nil
}

// MustCounter 创建空操作计数器
func (m *noopMeter) MustCounter(name string, option MetricOption) Counter {
	return &noopCounter{}
}

// UpDownCounter 创建空操作上下计数器
func (m *noopMeter) UpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	return &noopUpDownCounter{}, nil
}

// MustUpDownCounter 创建空操作上下计数器
func (m *noopMeter) MustUpDownCounter(name string, option MetricOption) UpDownCounter {
	return &noopUpDownCounter{}
}

// Histogram 创建空操作直方图
func (m *noopMeter) Histogram(name string, option MetricOption) (Histogram, error) {
	return &noopHistogram{}, nil
}

// MustHistogram 创建空操作直方图
func (m *noopMeter) MustHistogram(name string, option MetricOption) Histogram {
	return &noopHistogram{}
}

// noopCounter 是一个空操作计数器，不执行任何实际操作
type noopCounter struct{}

// Add 不执行任何操作
func (c *noopCounter) Add(ctx context.Context, value float64, opts ...Option) {}

// Inc 不执行任何操作
func (c *noopCounter) Inc(ctx context.Context, opts ...Option) {}

// noopUpDownCounter 是一个空操作上下计数器，不执行任何实际操作
type noopUpDownCounter struct{}

// Add 不执行任何操作
func (c *noopUpDownCounter) Add(ctx context.Context, value float64, opts ...Option) {}

// Inc 不执行任何操作
func (c *noopUpDownCounter) Inc(ctx context.Context, opts ...Option) {}

// Dec 不执行任何操作
func (c *noopUpDownCounter) Dec(ctx context.Context, opts ...Option) {}

// noopHistogram 是一个空操作直方图，不执行任何实际操作
type noopHistogram struct{}

// Record 不执行任何操作
func (h *noopHistogram) Record(value float64, opts ...Option) {}
