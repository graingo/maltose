// Package mmetric 提供指标收集和监控的辅助函数
package mmetric

import (
	"context"
)

// ===== MeterOption 辅助函数 =====

// NewMeterOption 创建新的计量器选项
func NewMeterOption() MeterOption {
	return MeterOption{}
}

// WithInstrument 设置仪表名称
func (o MeterOption) WithInstrument(instrument string) MeterOption {
	o.Instrument = instrument
	return o
}

// WithInstrumentVersion 设置仪表版本
func (o MeterOption) WithInstrumentVersion(version string) MeterOption {
	o.InstrumentVersion = version
	return o
}

// WithMeterAttributes 设置属性
func (o MeterOption) WithMeterAttributes(attrs Attributes) MeterOption {
	o.Attributes = attrs
	return o
}

// ===== MetricOption 辅助函数 =====

// NewMetricOption 创建新的指标选项
func NewMetricOption() MetricOption {
	return MetricOption{}
}

// WithHelp 设置帮助信息
func (o MetricOption) WithHelp(help string) MetricOption {
	o.Help = help
	return o
}

// WithUnit 设置单位
func (o MetricOption) WithUnit(unit string) MetricOption {
	o.Unit = unit
	return o
}

// WithMetricAttributes 设置属性
func (o MetricOption) WithMetricAttributes(attrs Attributes) MetricOption {
	o.Attributes = attrs
	return o
}

// WithBuckets 设置直方图桶
func (o MetricOption) WithBuckets(buckets []float64) MetricOption {
	o.Buckets = buckets
	return o
}

// ===== Option 辅助函数 =====

// NewOption 创建新的操作选项
func NewOption() Option {
	return Option{
		Attributes: make(AttributeMap),
	}
}

// WithOptionAttributes 设置属性
func (o Option) WithOptionAttributes(attrs AttributeMap) Option {
	o.Attributes = attrs
	return o
}

// ===== 便捷函数 =====

// GetMeter 获取指定名称的计量器
func GetMeter(name string) Meter {
	return GetProvider().Meter(NewMeterOption().WithInstrument(name))
}

// NewCounter 创建计数器
func NewCounter(name string, option MetricOption) (Counter, error) {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.Counter(name, option)
}

// NewMustCounter 创建计数器，如果出错则 panic
func NewMustCounter(name string, option MetricOption) Counter {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.MustCounter(name, option)
}

// NewUpDownCounter 创建上下计数器
func NewUpDownCounter(name string, option MetricOption) (UpDownCounter, error) {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.UpDownCounter(name, option)
}

// NewMustUpDownCounter 创建上下计数器，如果出错则 panic
func NewMustUpDownCounter(name string, option MetricOption) UpDownCounter {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.MustUpDownCounter(name, option)
}

// NewHistogram 创建直方图
func NewHistogram(name string, option MetricOption) (Histogram, error) {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.Histogram(name, option)
}

// NewMustHistogram 创建直方图，如果出错则 panic
func NewMustHistogram(name string, option MetricOption) Histogram {
	meter := GetProvider().Meter(NewMeterOption().WithInstrument(name))
	return meter.MustHistogram(name, option)
}

// Shutdown 关闭提供者
func Shutdown(ctx context.Context) error {
	return GetProvider().Shutdown(ctx)
}
