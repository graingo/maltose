# OpenTelemetry HTTP Trace Exporter

本包提供了通过 HTTP 协议导出 OpenTelemetry 链路追踪数据的功能。
The

## 功能特性

- 通过 HTTP 协议将链路追踪数据发送到 OpenTelemetry Collector
- 支持自定义服务名称和资源属性
- 自动收集主机信息作为追踪属性
- 与 maltose 框架无缝集成

## 使用方法

### 基本用法

```go
package main

import (
    "context"
    "time"

    "github.com/graingo/maltose/contrib/trace/otlphttp"
    "github.com/graingo/maltose/frame/m"
    "github.com/graingo/maltose/net/mtrace"
    "go.opentelemetry.io/otel"
)

func main() {
    // 初始化 OTLP HTTP 链路追踪提供者
    // 参数: 服务名称, 收集器地址, 路径
    shutdown, err := otlphttp.Init("my-service", "localhost:4318", "/v1/traces")
    if err != nil {
        panic(err)
    }
    defer shutdown(context.Background())

    // 获取追踪器
    tracer := otel.Tracer("my-component")

    // 创建一个根 span
    ctx, span := tracer.Start(context.Background(), "main-operation")
    defer span.End()

    // 执行业务逻辑...
    time.Sleep(time.Millisecond * 100)

    // 创建子 span
    ctx, childSpan := tracer.Start(ctx, "child-operation")
    // 设置属性
    childSpan.SetAttributes(mtrace.String("key", "value"))
    // 记录事件
    childSpan.AddEvent("processing-item", mtrace.Int("item-id", 42))

    // 执行子操作...
    time.Sleep(time.Millisecond * 50)

    // 结束子 span
    childSpan.End()
}
```

### 与 Maltose HTTP 服务集成

```go
package main

import (
    "github.com/graingo/maltose/contrib/trace/otlphttp"
    "github.com/graingo/maltose/frame/m"
    "github.com/graingo/maltose/net/mhttp"
)

func main() {
    // 初始化 OTLP HTTP 链路追踪提供者
    shutdown, err := otlphttp.Init("my-http-service", "collector:4318", "/v1/traces")
    if err != nil {
        panic(err)
    }
    defer shutdown(m.Ctx())

    // 创建 HTTP 服务
    s := mhttp.NewServer()

    // 注册路由
    s.BindHandler("/hello", func(r *mhttp.Request) {
        r.Response.WriteString("Hello, World!")
    })

    // 启动服务器
    s.SetPort(8080)
    s.Run()
}
```

## 链路追踪最佳实践

### 命名规范

为了保持一致性和可读性，建议遵循以下命名规范：

- **Span 名称**：使用动词+名词格式，如 `get_user`、`process_payment`
- **属性键**：使用小写字母和下划线，如 `user_id`、`request_size`

### 错误处理

在 Span 中记录错误信息：

```go
import "go.opentelemetry.io/otel/codes"

// 记录错误
if err != nil {
    span.SetStatus(codes.Error, err.Error())
    span.RecordError(err)
}
```

### 设置适当的采样率

在生产环境中，可能需要降低采样率以减少数据量：

```go
// 在 Init 函数之前设置采样率
sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)) // 10% 采样率
```

## 注意事项

1. HTTP 协议适用于不支持 gRPC 的环境，如某些受限网络环境
2. 确保 OpenTelemetry Collector 已正确配置，能够接收 HTTP 协议的追踪数据
3. 默认端口为 4318，与 gRPC 协议的 4317 不同
4. 相比 gRPC，HTTP 协议可能在高负载下性能略低，但在大多数场景下差异不明显

## 版本兼容性

本包使用 OpenTelemetry SDK v1.34.0 及相关组件，与 maltose 框架兼容。
