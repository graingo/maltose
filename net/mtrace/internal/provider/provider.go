package provider

import (
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// TracerProvider is a wrapper for sdkTrace.TracerProvider.
type TracerProvider struct {
	*sdkTrace.TracerProvider
}

// New returns a new configured TracerProvider.
// It acts as a lightweight wrapper around the OpenTelemetry SDK's NewTracerProvider,
// ensuring a consistent creation pattern within the Maltose framework.
// By default, it uses the OpenTelemetry SDK's default IDGenerator (random IDs).
// Users can override this and other options by providing sdkTrace.TracerProviderOption.
func New(opts ...sdkTrace.TracerProviderOption) *TracerProvider {
	return &TracerProvider{
		TracerProvider: sdkTrace.NewTracerProvider(opts...),
	}
}
