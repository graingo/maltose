// Package mmetric 提供指标收集和监控的抽象接口
package mmetric

import (
	"context"
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
	// Counter 创建计数器
	Counter(name string, option MetricOption) (Counter, error)

	// MustCounter 创建计数器，如果创建失败则 panic
	MustCounter(name string, option MetricOption) Counter

	// UpDownCounter 创建上下计数器
	UpDownCounter(name string, option MetricOption) (UpDownCounter, error)

	// MustUpDownCounter 创建上下计数器，如果创建失败则 panic
	MustUpDownCounter(name string, option MetricOption) UpDownCounter

	// Histogram 创建直方图
	Histogram(name string, option MetricOption) (Histogram, error)

	// MustHistogram 创建直方图，如果创建失败则 panic
	MustHistogram(name string, option MetricOption) Histogram
}

// Counter 计数器接口，用于记录单调递增的值
type Counter interface {
	// Add 添加指定值
	Add(ctx context.Context, value float64, opts ...Option)

	// Inc 增加 1
	Inc(ctx context.Context, opts ...Option)
}

// UpDownCounter 上下计数器接口，用于记录可增可减的值
type UpDownCounter interface {
	// Add 添加指定值（可为负数）
	Add(ctx context.Context, value float64, opts ...Option)

	// Inc 增加 1
	Inc(ctx context.Context, opts ...Option)

	// Dec 减少 1
	Dec(ctx context.Context, opts ...Option)
}

// Histogram 直方图接口，用于记录值的分布情况
type Histogram interface {
	// Record 记录一个值
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
	Buckets    []float64  // 直方图桶（可能不被所有实现支持）
}

// Option 操作选项，用于记录指标时添加标签
type Option struct {
	Attributes AttributeMap // 属性映射
}

// Attributes 属性列表，键值对形式
type Attributes map[string]string

// AttributeMap 属性映射，支持多种类型的值
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

// 全局变量
var (
	enabled                  = true              // 是否启用指标收集
	defaultProvider          = newNoopProvider() // 默认的空操作提供者
	activeProvider  Provider = defaultProvider   // 当前活跃的提供者，修改为 Provider 接口类型
)

// IsEnabled 检查是否启用了指标收集
func IsEnabled() bool {
	return enabled
}

// SetEnabled 设置是否启用指标收集
func SetEnabled(e bool) {
	enabled = e
}

// SetProvider 设置活跃的提供者
func SetProvider(provider Provider) {
	activeProvider = provider
}

// GetProvider 获取当前活跃的提供者
func GetProvider() Provider {
	if !enabled {
		return defaultProvider
	}
	return activeProvider
}
