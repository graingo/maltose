package otlpmetric

import (
	"context"

	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// 提供者包装器
type providerWrapper struct {
	provider *sdkmetric.MeterProvider
}

// 创建新的提供者包装器
func newProviderWrapper(provider *sdkmetric.MeterProvider) *providerWrapper {
	return &providerWrapper{provider: provider}
}

// Meter 实现 mmetric.Provider 接口
func (p *providerWrapper) Meter(option mmetric.MeterOption) mmetric.Meter {
	// 创建 OpenTelemetry Meter
	meter := p.provider.Meter(
		option.Instrument,
		metric.WithInstrumentationVersion(option.InstrumentVersion),
	)

	// 包装为我们的 Meter 接口
	return &meterWrapper{meter: meter}
}

// Shutdown 实现 mmetric.Provider 接口
func (p *providerWrapper) Shutdown(ctx context.Context) error {
	return p.provider.Shutdown(ctx)
}
