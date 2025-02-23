<div align=center>
<h1>MALTOSE</h1>

基于 Gin 打造的轻量级企业级开发框架

[![Go Reference](https://pkg.go.dev/badge/github.com/mingzaily/maltose.svg)](https://pkg.go.dev/github.com/mingzaily/maltose)
[![License](https://img.shields.io/github/license/mingzaily/maltose.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mingzaily/maltose)](https://goreportcard.com/report/github.com/mingzaily/maltose)

</div>

# 特性

- 基于 [Gin](https://github.com/gin-gonic/gin) 高性能 HTTP 框架
- 完全兼容 Gin 的中间件生态
- 集成 OpenTelemetry 分布式追踪
- 基于 Logrus 的结构化日志系统
- 类 GoFrame 的元数据路由注册
- 支持自动生成 OpenAPI/Swagger 文档
- 简洁优雅的项目结构

# 快速开始

## 安装

```bash
go get -u github.com/mingzaily/maltose
```

## Hello World

```go
package main

import (
    "github.com/mingzaily/maltose/frame/m"
)

type HelloReq struct {
    m.Meta `path:"/hello" method:"GET"`
    Name   string `form:"name"`
}

type HelloRes struct {
    Message string `json:"message"`
}

type HelloController struct{}

func (h *HelloController) Hello(ctx context.Context, req *HelloReq) (*HelloRes, error) {
    return &HelloRes{
        Message: "Hello, " + req.Name,
    }, nil
}

func main() {
    s := m.Server()
    s.Bind(&HelloController{})
    s.Run()
}
```

# 配置示例

```yaml
# config.yaml
server:
  address: ":8080"
  serverRoot: "./public"
  logger:
    path: "./logs"
    level: "debug"
    format: "text"
```

# 组件说明

## HTTP 服务 (mhttp)

- 基于 Gin 的 HTTP 服务框架
- 支持元数据路由注册
- 集成追踪和日志中间件
- 完整的请求生命周期管理

## 日志系统 (mlog)

- 基于 Logrus 的结构化日志
- 支持多输出目标
- 集成 OpenTelemetry 追踪
- 灵活的日志格式配置

## 配置管理 (mcfg)

- 支持 YAML/JSON 配置
- 多环境配置管理
- 配置热重载
- 配置变更通知

## 分布式追踪 (mtrace)

- OpenTelemetry 集成
- 支持多个追踪后端
- 自动注入追踪上下文
- 完整的链路追踪

# 项目结构

```
maltose/
├── frame/          # 框架核心
│   ├── m/         # 框架入口
│   └── mins/      # 框架实例管理
├── net/           # 网络相关
│   └── mhttp/     # HTTP 服务实现
├── os/            # 系统相关
│   ├── mcfg/      # 配置管理
│   └── mlog/      # 日志系统
└── util/          # 工具包
    └── mmeta/     # 元数据工具
```

# 贡献指南

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

# 开源许可

`Maltose` 采用 [MIT 许可证](LICENSE) 开源，永久免费。
