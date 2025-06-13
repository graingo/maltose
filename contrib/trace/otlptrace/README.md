# OpenTelemetry OTLP Trace 导出器

本包提供了通过 OTLP 导出 OpenTelemetry 追踪数据的功能，同时支持 gRPC 和 HTTP 协议。

## 功能特性

- 通过 gRPC 或 HTTP 将追踪数据导出到 OpenTelemetry Collector。
- 支持自定义服务名称和资源属性。
- 自动收集主机信息作为追踪属性。
- 与 Maltose 框架无缝集成。

## 安装

```bash
go get github.com/graingo/maltose/contrib/trace/otlptrace
```

## 使用方法

### 使用 gRPC (默认)

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/graingo/maltose/contrib/trace/otlptrace"
    "go.opentelemetry.io/otel"
)

func main() {
    // 初始化 OTLP gRPC 追踪提供者
    // 参数: collector 地址
    shutdown, err := otlptrace.Init("localhost:4317",
        otlptrace.WithServiceName("my-service-grpc"),
    )
    if err != nil {
        log.Fatalf("无法初始化追踪提供者: %v", err)
    }
    defer shutdown(context.Background())

    // 获取追踪器
    tracer := otel.Tracer("my-component")

    // 创建一个根 Span
    _, span := tracer.Start(context.Background(), "main-operation")
    defer span.End()

    // ... 你的业务逻辑
    time.Sleep(time.Millisecond * 100)
}
```

### 使用 HTTP

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/graingo/maltose/contrib/trace/otlptrace"
    "go.opentelemetry.io/otel"
)

func main() {
    // 初始化 OTLP HTTP 追踪提供者
    shutdown, err := otlptrace.Init("localhost:4318",
        otlptrace.WithServiceName("my-service-http"),
        otlptrace.WithProtocol(otlptrace.ProtocolHTTP),
        otlptrace.WithURLPath("/v1/traces"),
    )
    if err != nil {
        log.Fatalf("无法初始化追踪提供者: %v", err)
    }
    defer shutdown(context.Background())

    // 获取追踪器
    tracer := otel.Tracer("my-component")

    // 创建一个根 Span
    _, span := tracer.Start(context.Background(), "main-operation")
    defer span.End()

    // ... 你的业务逻辑
    time.Sleep(time.Millisecond * 100)
}
```

## 配置选项

- `WithServiceName(name string)`: 设置服务名称。
- `WithServiceVersion(version string)`: 设置服务版本。
- `WithEnvironment(env string)`: 设置部署环境。
- `WithProtocol(protocol Protocol)`: 设置协议 (`otlptrace.ProtocolGRPC` 或 `otlptrace.ProtocolHTTP`)。
- `WithTimeout(timeout time.Duration)`: 设置导出超时时间。
- `WithInsecure(insecure bool)`: 启用/禁用非安全连接。
- `WithURLPath(path string)`: 设置 HTTP 导出器的 URL 路径。
- `WithCompression(level int)`: 设置 HTTP 导出器的压缩级别。
- `WithResourceAttribute(key, value string)`: 添加自定义资源属性。
- `WithSampler(sampler trace.Sampler)`: 设置追踪采样器 (例如, `trace.TraceIDRatioBased(0.1)` 用于 10% 的采样率)。

## 注意事项

1.  **gRPC** 是 OpenTelemetry 推荐的传输协议，它提供更好的性能和稳定性。默认端口是 `4317`。
2.  **HTTP** 适用于不支持 gRPC 的环境。默认端口是 `4318`。
3.  请确保您的 OpenTelemetry Collector 已正确配置，以在所选的协议和端口上接收追踪数据。
4.  在生产环境中，建议配置适当的采样策略以避免产生过多的追踪数据。
