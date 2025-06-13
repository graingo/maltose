# OpenTelemetry 指标导出器

本包提供了基于 OpenTelemetry 的指标收集和导出功能，支持通过 gRPC 或 HTTP 协议将指标数据发送到 OpenTelemetry Collector。

## 功能特点

- 支持 gRPC 和 HTTP 协议导出
- 自动周期性导出指标数据
- 支持多种指标类型：Counter、UpDownCounter、Histogram
- 与 Maltose 框架的 mmetric 包无缝集成
- 提供丰富的配置选项

## 安装

```bash
go get github.com/graingo/maltose/contrib/metric/otlpmetric
```

## 基本用法

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/graingo/maltose/contrib/metric/otlpmetric"
	"github.com/graingo/maltose/os/mmetric"
)

func main() {
	// 初始化指标导出器
	shutdown, err := otlpmetric.Init("localhost:4317",
		otlpmetric.WithServiceName("my-service"),
		otlpmetric.WithServiceVersion("1.0.0"),
		otlpmetric.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatalf("初始化指标失败: %v", err)
	}
	defer shutdown(context.Background())

	// 创建计数器
	counter := mmetric.MustCounter("my_counter", mmetric.MetricOption{
		Help: "这是一个示例计数器",
		Unit: "1",
	})

	// 增加计数器值
	counter.Inc(context.Background(), mmetric.WithAttributes(map[string]interface{}{
		"label1": "value1",
		"label2": "value2",
	}))

	// 等待指标导出
	time.Sleep(15 * time.Second)
}
```

## 高级用法

### 在 HTTP 服务中使用

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/graingo/maltose/contrib/metric/otlpmetric"
	"github.com/graingo/maltose/os/mmetric"
)

func main() {
	// 初始化指标导出器
	shutdown, err := otlpmetric.Init("localhost:4317",
		otlpmetric.WithServiceName("http-service"),
		otlpmetric.WithExportInterval(5 * time.Second),
	)
	if err != nil {
		log.Fatalf("初始化指标失败: %v", err)
	}
	defer shutdown(context.Background())

	// 创建请求计数器
	requestCounter := mmetric.MustCounter("http_requests_total", mmetric.MetricOption{
		Help: "HTTP 请求总数",
		Unit: "1",
	})

	// 创建请求延迟直方图
	latencyHistogram := mmetric.MustHistogram("http_request_duration_seconds", mmetric.MetricOption{
		Help: "HTTP 请求延迟（秒）",
		Unit: "s",
	})

	// 创建中间件
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// 处理请求
		w.Write([]byte("Hello, World!"))

		// 记录指标
		duration := time.Since(startTime).Seconds()
		requestCounter.Inc(r.Context(), mmetric.WithAttributes(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		}))
		latencyHistogram.Record(duration, mmetric.WithAttributes(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		}))
	})

	// 启动服务器
	log.Println("服务器启动在 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 配置选项

本包提供了多种配置选项，可以根据需要进行设置：

- `WithServiceName(name string)`: 设置服务名称
- `WithServiceVersion(version string)`: 设置服务版本
- `WithEnvironment(env string)`: 设置环境名称
- `WithProtocol(protocol Protocol)`: 设置协议（ProtocolGRPC 或 ProtocolHTTP）
- `WithExportInterval(interval time.Duration)`: 设置导出间隔
- `WithTimeout(timeout time.Duration)`: 设置超时时间
- `WithInsecure(insecure bool)`: 设置是否使用非安全连接
- `WithURLPath(path string)`: 设置 URL 路径（仅用于 HTTP 协议）
- `WithResourceAttribute(key, value string)`: 添加资源属性

## 最佳实践

1. 设置合理的导出间隔，避免过于频繁的导出影响性能
2. 使用有意义的指标名称和标签，便于后续查询和分析
3. 在生产环境中，建议使用安全连接（设置 WithInsecure(false)）
4. 为服务设置正确的名称、版本和环境信息，便于区分不同的服务和环境
5. 使用 Must\* 方法创建指标，简化错误处理
