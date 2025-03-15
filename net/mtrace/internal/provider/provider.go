package provider

import (
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// TracerProvider is a wrapper for sdkTrace.TracerProvider.
type TracerProvider struct {
	*sdkTrace.TracerProvider
}

// New returns a new configured TracerProvider.
// The default configuration includes:
// - ParentBased(AlwaysSample) sampler
// - IDGenerator based on unix nano timestamp and random number
// - resource.Default() Resource
// - default SpanLimits
func New() *TracerProvider {
	return &TracerProvider{
		TracerProvider: sdkTrace.NewTracerProvider(
			sdkTrace.WithIDGenerator(NewIDGenerator()),
		),
	}
}
