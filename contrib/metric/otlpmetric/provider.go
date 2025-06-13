package otlpmetric

import (
	"context"

	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// providerWrapper is a wrapper for a provider.
type providerWrapper struct {
	provider *sdkmetric.MeterProvider
}

// newProviderWrapper creates a new provider wrapper.
func newProviderWrapper(provider *sdkmetric.MeterProvider) *providerWrapper {
	return &providerWrapper{provider: provider}
}

// Meter implements the mmetric.Provider interface.
func (p *providerWrapper) Meter(option mmetric.MeterOption) mmetric.Meter {
	// Create an OpenTelemetry Meter.
	meter := p.provider.Meter(
		option.Instrument,
		metric.WithInstrumentationVersion(option.InstrumentVersion),
	)

	// Wrap it in our Meter interface.
	return &meterWrapper{meter: meter}
}

// Shutdown implements the mmetric.Provider interface.
func (p *providerWrapper) Shutdown(ctx context.Context) error {
	return p.provider.Shutdown(ctx)
}
