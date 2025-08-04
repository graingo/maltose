package mmetric_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mmetric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// setupTestProvider creates a test MeterProvider with a manual reader
// to allow for manual collection and verification of metrics.
func setupTestProvider(t *testing.T) (mmetric.Provider, *metric.ManualReader) {
	t.Helper()

	reader := metric.NewManualReader()
	// Using NewProvider from our package, which wraps the OTel SDK function.
	provider := mmetric.NewProvider(metric.WithReader(reader))

	// Set this as the global provider for the duration of the test.
	mmetric.SetProvider(provider)

	t.Cleanup(func() {
		// Shutdown the provider to flush any remaining metrics.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := provider.Shutdown(ctx)
		assert.NoError(t, err)
		// Reset to the default global provider after the test.
		mmetric.SetProvider(metric.NewMeterProvider())
	})

	// Use our mmetric.Provider interface.
	return mmetric.GetProvider(), reader
}

// findSumMetric searches for a metric that aggregates as a Sum (like Counter, UpDownCounter)
// and returns its data points.
func findSumMetric[N int64 | float64](t *testing.T, data metricdata.ResourceMetrics, name string) []metricdata.DataPoint[N] {
	t.Helper()
	for _, sm := range data.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				agg, ok := m.Data.(metricdata.Sum[N])
				if !ok {
					t.Fatalf("metric '%s' has unexpected aggregation type %T, expected Sum", name, m.Data)
				}
				return agg.DataPoints
			}
		}
	}
	t.Fatalf("metric '%s' not found in collected data", name)
	return nil
}

// findHistogramMetric searches for a Histogram metric and returns its data points.
func findHistogramMetric[N int64 | float64](t *testing.T, data metricdata.ResourceMetrics, name string) []metricdata.HistogramDataPoint[N] {
	t.Helper()
	for _, sm := range data.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				agg, ok := m.Data.(metricdata.Histogram[N])
				if !ok {
					t.Fatalf("metric '%s' has unexpected aggregation type %T, expected Histogram", name, m.Data)
				}
				return agg.DataPoints
			}
		}
	}
	t.Fatalf("metric '%s' not found in collected data", name)
	return nil
}

// assertAttributes asserts that the given attribute set contains all the expected attributes.
func assertAttributes(t *testing.T, set attribute.Set, expected ...attribute.KeyValue) {
	t.Helper()
	for _, kv := range expected {
		val, ok := set.Value(kv.Key)
		require.True(t, ok, "expected attribute '%s' not found", kv.Key)
		assert.Equal(t, kv.Value, val, "attribute '%s' has incorrect value", kv.Key)
	}
}

func TestProviderAndMeter(t *testing.T) {
	t.Run("get_and_set_provider", func(t *testing.T) {
		provider, _ := setupTestProvider(t)
		require.NotNil(t, provider)

		// GetProvider should return the provider we just set.
		// Note: We can't compare them directly, but we can check if it works.
		meter := provider.Meter(mmetric.MeterOption{Instrument: "test_instrument"})
		require.NotNil(t, meter)
	})

	t.Run("meter_with_attributes", func(t *testing.T) {
		provider, reader := setupTestProvider(t)
		meterAttrs := mmetric.Attributes{attribute.String("service", "test-app")}

		meter := provider.Meter(mmetric.MeterOption{
			Instrument: "test_meter_with_attrs",
			Attributes: meterAttrs,
		})

		counter, err := meter.Counter("test_counter", mmetric.MetricOption{})
		require.NoError(t, err)
		counter.Inc(context.Background())

		data := metricdata.ResourceMetrics{}
		err = reader.Collect(context.Background(), &data)
		require.NoError(t, err)

		points := findSumMetric[float64](t, data, "test_counter")
		require.Len(t, points, 1)
		assertAttributes(t, points[0].Attributes, meterAttrs...)
	})

	t.Run("shutdown_provider", func(t *testing.T) {
		reader := metric.NewManualReader()
		provider := mmetric.NewProvider(metric.WithReader(reader))
		mmetric.SetProvider(provider)

		p := mmetric.GetProvider()
		err := p.Shutdown(context.Background())
		assert.NoError(t, err)

		// Test shutdown on a No-Op provider, which should also not error.
		mmetric.SetProvider(metric.NewMeterProvider())
		p = mmetric.GetProvider()
		err = p.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}

func TestInstruments(t *testing.T) {
	t.Run("counter_with_merged_attributes", func(t *testing.T) {
		provider, reader := setupTestProvider(t)
		meterAttrs := mmetric.Attributes{attribute.String("meter_key", "meter_val")}
		metricAttrs := mmetric.Attributes{attribute.String("metric_key", "metric_val")}
		opAttrs := mmetric.Attributes{attribute.String("op_key", "op_val")}

		meter := provider.Meter(mmetric.MeterOption{Instrument: "test_instrument", Attributes: meterAttrs})
		counter, err := meter.Counter("test.counter", mmetric.MetricOption{Attributes: metricAttrs})
		require.NoError(t, err)

		counter.Add(context.Background(), 5, mmetric.WithAttributes(opAttrs...))
		counter.Inc(context.Background(), mmetric.WithAttributes(opAttrs...)) // total 6

		data := metricdata.ResourceMetrics{}
		err = reader.Collect(context.Background(), &data)
		require.NoError(t, err)

		points := findSumMetric[float64](t, data, "test.counter")
		require.Len(t, points, 1)

		assert.Equal(t, float64(6), points[0].Value)
		assertAttributes(t, points[0].Attributes, meterAttrs[0], metricAttrs[0], opAttrs[0])
	})

	t.Run("up_down_counter_with_merged_attributes", func(t *testing.T) {
		provider, reader := setupTestProvider(t)
		meterAttrs := mmetric.Attributes{attribute.Int("meter_id", 1)}
		metricAttrs := mmetric.Attributes{attribute.Bool("metric_bool", true)}
		opAttrs := mmetric.Attributes{attribute.String("op_key", "op_val")}

		meter := provider.Meter(mmetric.MeterOption{Instrument: "test_instrument", Attributes: meterAttrs})
		counter, err := meter.UpDownCounter("test.updown_counter", mmetric.MetricOption{Attributes: metricAttrs})
		require.NoError(t, err)

		counter.Add(context.Background(), 10)
		counter.Inc(context.Background(), mmetric.WithAttributes(opAttrs...)) // 11
		counter.Dec(context.Background())                                     // 10
		counter.Add(context.Background(), -3)                                 // 7

		data := metricdata.ResourceMetrics{}
		err = reader.Collect(context.Background(), &data)
		require.NoError(t, err)

		points := findSumMetric[float64](t, data, "test.updown_counter")
		// We now expect 2 data points, because operations with the same attribute
		// set are merged by the SDK.
		require.Len(t, points, 2)

		// Check both data points.
		var opPointFound, mergedPointFound bool
		for _, p := range points {
			if p.Attributes.HasValue("op_key") {
				// This is the point from the Inc() call with operation attributes.
				assert.Equal(t, float64(1), p.Value)
				assertAttributes(t, p.Attributes, meterAttrs[0], metricAttrs[0], opAttrs[0])
				opPointFound = true
			} else {
				// This is the merged point from the other three calls (10 - 1 - 3).
				assert.Equal(t, float64(6), p.Value)
				assertAttributes(t, p.Attributes, meterAttrs[0], metricAttrs[0])
				mergedPointFound = true
			}
		}
		require.True(t, opPointFound, "data point with operation attributes not found")
		require.True(t, mergedPointFound, "merged data point not found")
	})

	t.Run("histogram_with_merged_attributes", func(t *testing.T) {
		provider, reader := setupTestProvider(t)
		meterAttrs := mmetric.Attributes{attribute.String("meter_key", "meter_val")}
		metricAttrs := mmetric.Attributes{attribute.String("metric_key", "metric_val")}
		opAttrs := mmetric.Attributes{attribute.String("op_key", "op_val")}

		buckets := []float64{10, 20, 30}
		meter := provider.Meter(mmetric.MeterOption{Instrument: "test_instrument", Attributes: meterAttrs})
		histogram, err := meter.Histogram("test.histogram", mmetric.MetricOption{
			Attributes: metricAttrs,
			Unit:       "ms",
			Buckets:    buckets,
		})
		require.NoError(t, err)

		histogram.Record(15, mmetric.WithAttributes(opAttrs...))
		histogram.Record(25)

		data := metricdata.ResourceMetrics{}
		err = reader.Collect(context.Background(), &data)
		require.NoError(t, err)

		points := findHistogramMetric[float64](t, data, "test.histogram")
		require.Len(t, points, 2)

		var found bool
		for _, p := range points {
			if p.Attributes.HasValue("op_key") {
				assert.Equal(t, uint64(1), p.Count)
				assert.Equal(t, float64(15), p.Sum)
				// The number of bucket counts should be one more than the number of boundaries.
				assert.Len(t, p.BucketCounts, len(buckets)+1)
				// The value 15 should fall into the second bucket (10, 20].
				assert.Equal(t, uint64(1), p.BucketCounts[1])
				assertAttributes(t, p.Attributes, meterAttrs[0], metricAttrs[0], opAttrs[0])
				found = true
			}
		}
		require.True(t, found, "data point with operation attributes not found")
	})
}

// errorMeter is a mock meter that always returns an error, for testing Must... functions.
type errorMeter struct{}

func (e *errorMeter) Counter(string, mmetric.MetricOption) (mmetric.Counter, error) {
	return nil, errors.New("mock error")
}
func (e *errorMeter) MustCounter(name string, option mmetric.MetricOption) mmetric.Counter {
	c, err := e.Counter(name, option)
	if err != nil {
		panic(err)
	}
	return c
}
func (e *errorMeter) UpDownCounter(string, mmetric.MetricOption) (mmetric.UpDownCounter, error) {
	return nil, errors.New("mock error")
}
func (e *errorMeter) MustUpDownCounter(name string, option mmetric.MetricOption) mmetric.UpDownCounter {
	c, err := e.UpDownCounter(name, option)
	if err != nil {
		panic(err)
	}
	return c
}
func (e *errorMeter) Histogram(string, mmetric.MetricOption) (mmetric.Histogram, error) {
	return nil, errors.New("mock error")
}
func (e *errorMeter) MustHistogram(name string, option mmetric.MetricOption) mmetric.Histogram {
	h, err := e.Histogram(name, option)
	if err != nil {
		panic(err)
	}
	return h
}

func TestMustPanics(t *testing.T) {
	meter := &errorMeter{}

	t.Run("must_counter_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			meter.MustCounter("any", mmetric.MetricOption{})
		})
	})

	t.Run("must_up_down_counter_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			meter.MustUpDownCounter("any", mmetric.MetricOption{})
		})
	})

	t.Run("must_histogram_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			meter.MustHistogram("any", mmetric.MetricOption{})
		})
	})
}

func TestHelpers(t *testing.T) {
	t.Run("package_level_helpers", func(t *testing.T) {
		_, reader := setupTestProvider(t)

		counter := mmetric.NewMustCounter("pkg.counter", mmetric.MetricOption{})
		counter.Inc(context.Background())

		data := metricdata.ResourceMetrics{}
		err := reader.Collect(context.Background(), &data)
		require.NoError(t, err)

		points := findSumMetric[float64](t, data, "pkg.counter")
		require.Len(t, points, 1)
		assert.Equal(t, float64(1), points[0].Value)
	})

	t.Run("option_helpers", func(t *testing.T) {
		meterOpt := mmetric.NewMeterOption().
			WithInstrument("inst").
			WithInstrumentVersion("v1").
			WithMeterAttributes(mmetric.Attributes{attribute.Bool("a", true)})

		assert.Equal(t, "inst", meterOpt.Instrument)
		assert.Equal(t, "v1", meterOpt.InstrumentVersion)
		assert.Len(t, meterOpt.Attributes, 1)

		metricOpt := mmetric.NewMetricOption().
			WithHelp("help").
			WithUnit("ms").
			WithBuckets([]float64{1, 2}).
			WithMetricAttributes(mmetric.Attributes{attribute.Bool("b", true)})

		assert.Equal(t, "help", metricOpt.Help)
		assert.Equal(t, "ms", metricOpt.Unit)
		assert.True(t, slices.Equal([]float64{1, 2}, metricOpt.Buckets))
		assert.Len(t, metricOpt.Attributes, 1)

		opOpt := mmetric.WithAttributes(attribute.Int("c", 3))
		assert.Len(t, opOpt.Attributes, 1)
		assert.Equal(t, attribute.Int("c", 3), opOpt.Attributes[0])
	})

	t.Run("get_meter_helper", func(t *testing.T) {
		_, reader := setupTestProvider(t)
		meter := mmetric.GetMeter("my-instrument-name")
		require.NotNil(t, meter)
		counter := meter.MustCounter("my.counter", mmetric.MetricOption{})
		counter.Inc(context.Background())

		data := metricdata.ResourceMetrics{}
		err := reader.Collect(context.Background(), &data)
		require.NoError(t, err)

		// Check that the meter was created with the correct instrument name (scope name)
		var found bool
		for _, sm := range data.ScopeMetrics {
			if sm.Scope.Name == "my-instrument-name" {
				found = true
				break
			}
		}
		assert.True(t, found, "meter with expected instrument name not found")
	})

	t.Run("shutdown_helper", func(t *testing.T) {
		// This test ensures the package-level Shutdown helper works.
		// It creates a real provider, shuts it down, and asserts no error.
		reader := metric.NewManualReader()
		provider := mmetric.NewProvider(metric.WithReader(reader))
		mmetric.SetProvider(provider)

		// Calling the package-level helper
		err := mmetric.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}
