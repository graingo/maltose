package observability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDisabledReturnsNoopProvider(t *testing.T) {
	provider, err := New(context.Background(), Config{})
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.NoError(t, provider.Close())
}

func TestNewValidatesEnabledSignals(t *testing.T) {
	_, err := New(context.Background(), Config{Enabled: true})
	require.EqualError(t, err, "observability requires at least one enabled signal")

	_, err = New(context.Background(), Config{
		Enabled: true,
		Trace:   TraceConfig{Enabled: true},
	})
	require.EqualError(t, err, "observability.trace.endpoint is required")

	_, err = New(context.Background(), Config{
		Enabled: true,
		Metric: MetricConfig{
			Enabled:  true,
			Endpoint: "localhost:4317",
			Protocol: "invalid",
		},
	})
	require.EqualError(t, err, `unsupported observability.metric.protocol "invalid"`)
}

func TestFromConfigLoadsDefaults(t *testing.T) {
	adapter, err := mcfg.NewAdapterContent(`
observability:
  enabled: false
  service_name: checkout
`, "yaml")
	require.NoError(t, err)

	provider, err := FromConfig(context.Background(), mcfg.NewWithAdapter(adapter))
	require.NoError(t, err)
	assert.Equal(t, defaultShutdownTimeout, provider.shutdownTimeout)
}

func TestProviderShutdownIsIdempotent(t *testing.T) {
	metricErr := errors.New("metric shutdown failed")
	var traceCalls, metricCalls int
	provider := &Provider{
		shutdownTimeout: time.Second,
		traceShutdown: func(context.Context) {
			traceCalls++
		},
		metricShutdown: func(context.Context) error {
			metricCalls++
			return metricErr
		},
	}

	require.ErrorIs(t, provider.Shutdown(context.Background()), metricErr)
	require.ErrorIs(t, provider.Shutdown(context.Background()), metricErr)
	assert.Equal(t, 1, traceCalls)
	assert.Equal(t, 1, metricCalls)
}
