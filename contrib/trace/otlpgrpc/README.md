# OpenTelemetry gRPC Trace Exporter

本包提供了通过 gRPC 协议导出 OpenTelemetry 链路追踪数据的功能。

## 功能特性

- 通过 gRPC 协议将链路追踪数据发送到 OpenTelemetry Collector
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

    "github.com/graingo/maltose/contrib/trace/otlpgrpc"
    "github.com/graingo/maltose/frame/m"
    "github.com/graingo/maltose/net/mtrace"
    "go.opentelemetry.io/otel"
)

func main() {
    // 初始化 OTLP gRPC 链路追踪提供者
    // 参数: 服务名称, 收集器地址
    shutdown, err := otlpgrpc.Init("my-service", "localhost:4317")
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

### 与 HTTP 服务集成

```go
package main

import (
    "net/http"

    "github.com/graingo/maltose/contrib/trace/otlpgrpc"
    "github.com/graingo/maltose/frame/m"
    "github.com/graingo/maltose/net/mtrace"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
    // 初始化 OTLP gRPC 链路追踪提供者
    shutdown, err := otlpgrpc.Init("my-http-service", "collector:4317")
    if err != nil {
        panic(err)
    }
    defer shutdown(m.Ctx())

    // 创建带有追踪的 HTTP 处理函数
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // 使用 otelhttp 包装 HTTP 处理函数
    tracedHandler := otelhttp.NewHandler(handler, "hello-handler")

    // 启动 HTTP 服务器
    http.Handle("/hello", tracedHandler)
    http.ListenAndServe(":8080", nil)
}
```

## 链路追踪概念

### Span

Span 是链路追踪的基本单位，表示一个操作或工作单元。每个 Span 包含以下信息：

- 名称：描述操作的名称
- 开始和结束时间：操作的持续时间
- 属性：键值对形式的元数据
- 事件：带时间戳的日志
- 状态：操作是否成功
- 父 Span：表示调用关系的层次结构

### 上下文传播

上下文传播允许在不同服务之间传递追踪信息，实现分布式追踪。Maltose 框架自动处理了上下文传播，您只需要确保在服务调用时传递 context。

## 注意事项

1. gRPC 是 OpenTelemetry 推荐的传输协议，具有更好的性能和稳定性
2. 确保 OpenTelemetry Collector 已正确配置，能够接收 gRPC 协议的追踪数据
3. 默认端口为 4317，与 HTTP 协议的 4318 不同
4. 在生产环境中，建议配置适当的采样策略，避免产生过多的追踪数据

## 版本兼容性

本包使用 OpenTelemetry SDK v1.34.0 及相关组件，与 maltose 框架兼容。
