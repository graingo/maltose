# Maltose Observability Bootstrap

This module initializes the existing OTLP trace and metric exporters from one configuration node and exposes a single lifecycle resource.

```yaml
observability:
  enabled: true
  service_name: checkout-api
  service_version: v1.2.0
  environment: production
  insecure: true
  shutdown_timeout: 10s
  trace:
    enabled: true
    endpoint: otel-collector:4317
    protocol: grpc
    sample_ratio: 0.1
  metric:
    enabled: true
    endpoint: otel-collector:4317
    protocol: grpc
    export_interval: 10s
```

```go
telemetry, err := observability.FromConfig(ctx, m.Config())
if err != nil {
	return err
}

app := m.NewApp(
	m.WithServer(server),
	m.WithCloser(telemetry),
)
return app.Run()
```

Set `enabled: false` or omit the node to receive a no-op provider that is safe to close.
