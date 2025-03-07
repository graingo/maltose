package otlpgrpc

import (
	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/metric"
)

type meter struct {
	metric.Meter
}

func newMeter(m metric.Meter) mmetric.Meter {
	return &meter{Meter: m}
}

func (m *meter) Counter(name string, option mmetric.MetricOption) (mmetric.Counter, error) {
	counter, err := m.Meter.Float64Counter(name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &counterWrapper{counter: counter}, nil
}

func (m *meter) MustCounter(name string, option mmetric.MetricOption) mmetric.Counter {
	counter, err := m.Counter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

func (m *meter) UpDownCounter(name string, option mmetric.MetricOption) (mmetric.UpDownCounter, error) {
	counter, err := m.Meter.Float64UpDownCounter(name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &upDownCounterWrapper{counter: counter}, nil
}

func (m *meter) MustUpDownCounter(name string, option mmetric.MetricOption) mmetric.UpDownCounter {
	counter, err := m.UpDownCounter(name, option)
	if err != nil {
		panic(err)
	}
	return counter
}

func (m *meter) Histogram(name string, option mmetric.MetricOption) (mmetric.Histogram, error) {
	histogram, err := m.Meter.Float64Histogram(name,
		metric.WithDescription(option.Help),
		metric.WithUnit(option.Unit),
	)
	if err != nil {
		return nil, err
	}
	return &histogramWrapper{histogram: histogram}, nil
}

func (m *meter) MustHistogram(name string, option mmetric.MetricOption) mmetric.Histogram {
	histogram, err := m.Histogram(name, option)
	if err != nil {
		panic(err)
	}
	return histogram
}
