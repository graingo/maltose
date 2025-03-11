package otlpmetric

import (
	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/metric"
)

// 计量器包装器
type meterWrapper struct {
	meter metric.Meter
}

// Counter 创建计数器
func (m *meterWrapper) Counter(name string, option mmetric.MetricOption) (mmetric.Counter, error) {
	counter, err := m.meter.Float64Counter(
		name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &counterWrapper{counter: counter}, nil
}

// MustCounter 创建计数器，如果出错则 panic
func (m *meterWrapper) MustCounter(name string, option mmetric.MetricOption) mmetric.Counter {
	counter, err := m.Counter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

// UpDownCounter 创建上下计数器
func (m *meterWrapper) UpDownCounter(name string, option mmetric.MetricOption) (mmetric.UpDownCounter, error) {
	counter, err := m.meter.Float64UpDownCounter(
		name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &upDownCounterWrapper{counter: counter}, nil
}

// MustUpDownCounter 创建上下计数器，如果出错则 panic
func (m *meterWrapper) MustUpDownCounter(name string, option mmetric.MetricOption) mmetric.UpDownCounter {
	counter, err := m.UpDownCounter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

// Histogram 创建直方图
func (m *meterWrapper) Histogram(name string, option mmetric.MetricOption) (mmetric.Histogram, error) {
	var opts []metric.Float64HistogramOption

	// 设置描述和单位
	opts = append(opts,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)

	// 设置桶（注意：OpenTelemetry v1.34.0 不再支持显式设置桶边界）
	// 桶配置现在由收集器或后端系统处理

	// 创建直方图
	histogram, err := m.meter.Float64Histogram(name, opts...)
	if err != nil {
		return nil, err
	}

	return &histogramWrapper{histogram: histogram}, nil
}

// MustHistogram 创建直方图，如果出错则 panic
func (m *meterWrapper) MustHistogram(name string, option mmetric.MetricOption) mmetric.Histogram {
	histogram, err := m.Histogram(name, option)
	if err != nil {
		panic(err)
	}
	return histogram
}
