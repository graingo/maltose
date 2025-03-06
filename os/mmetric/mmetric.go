package mmetric

import (
	"context"
	"sync"
)

// Provider 指标提供者接口
type Provider interface {
	// Meter 创建计量器
	Meter(option MeterOption) Meter

	// Shutdown 关闭提供者
	Shutdown(ctx context.Context) error
}

// Meter 计量器接口
type Meter interface {
	Counter(name string, option MetricOption) (Counter, error)
	MustCounter(name string, option MetricOption) Counter
	UpDownCounter(name string, option MetricOption) (UpDownCounter, error)
	MustUpDownCounter(name string, option MetricOption) UpDownCounter
	Histogram(name string, option MetricOption) (Histogram, error)
	MustHistogram(name string, option MetricOption) Histogram
}

// Counter 计数器接口
type Counter interface {
	Add(ctx context.Context, value float64, opts ...Option)
	Inc(ctx context.Context, opts ...Option)
}

// UpDownCounter 上下计数器接口
type UpDownCounter interface {
	Add(ctx context.Context, value float64, opts ...Option)
	Inc(ctx context.Context, opts ...Option)
	Dec(ctx context.Context, opts ...Option)
}

// Histogram 直方图接口
type Histogram interface {
	Record(value float64, opts ...Option)
}

// MeterOption 计量器选项
type MeterOption struct {
	Instrument        string     // 仪表名称
	InstrumentVersion string     // 仪表版本
	Attributes        Attributes // 属性
}

// MetricOption 指标选项
type MetricOption struct {
	Help       string     // 帮助信息
	Unit       string     // 单位
	Attributes Attributes // 属性
	Buckets    []float64  // 直方图桶
}

// Option 操作选项
type Option struct {
	Attributes AttributeMap
}

// Attributes 属性列表
type Attributes map[string]string

// AttributeMap 属性映射
type AttributeMap map[string]interface{}

// Sets 设置多个属性
func (m AttributeMap) Sets(attrs AttributeMap) {
	for k, v := range attrs {
		m[k] = v
	}
}

// Pick 选择特定属性
func (m AttributeMap) Pick(keys ...string) AttributeMap {
	result := make(AttributeMap)
	for _, key := range keys {
		if v, ok := m[key]; ok {
			result[key] = v
		}
	}
	return result
}

// 移除原来的 provider 包导入和默认提供者初始化
var (
	globalProviderMu sync.RWMutex
	globalProvider   Provider
	enabled          = true
	defaultProvider  = newNoopProvider()
)

// SetGlobalProvider 设置全局提供者
func SetGlobalProvider(provider Provider) {
	globalProviderMu.Lock()
	defer globalProviderMu.Unlock()
	globalProvider = provider
}

// GetGlobalProvider 获取全局提供者
func GetGlobalProvider() Provider {
	globalProviderMu.RLock()
	defer globalProviderMu.RUnlock()

	if globalProvider == nil {
		return defaultProvider
	}
	return globalProvider
}

// IsEnabled 是否启用指标
func IsEnabled() bool {
	return enabled
}

// SetEnabled 设置是否启用指标
func SetEnabled(e bool) {
	enabled = e
}
