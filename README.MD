<div align=center>
<h1>MALTOSE</h1>

基于 Gin 打造的轻量级企业级开发框架

[![Go Reference](https://pkg.go.dev/badge/github.com/graingo/maltose.svg)](https://pkg.go.dev/github.com/graingo/maltose)
[![License](https://img.shields.io/github/license/graingo/maltose.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/graingo/maltose)](https://goreportcard.com/report/github.com/graingo/maltose)

</div>

# 特性

- 基于 [Gin](https://github.com/gin-gonic/gin) 高性能 HTTP 框架
- 完全兼容 Gin 的中间件生态
- 集成 OpenTelemetry 分布式追踪
- 基于 Logrus 的结构化日志系统
- 类 GoFrame 的元数据路由注册
- 支持自动生成 OpenAPI/Swagger 文档
- 简洁优雅的项目结构
- 规范的代码分层架构
- 易于扩展的模块化设计

# 快速开始

## 安装

```bash
go get -u github.com/graingo/maltose
```

## 创建项目

通过我们提供的快速启动模板创建新项目：

```bash
# 克隆快速启动模板
git clone https://github.com/graingo/maltose-quickstart.git myproject

# 进入项目目录
cd myproject

# 修改go.mod中的模块名
# 将 github.com/graingo/maltose-quickstart 替换为你的项目包名

# 运行项目
go run main.go
```

## Hello World

```go
package main

import (
    "github.com/graingo/maltose/frame/m"
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

## 框架源码结构

```
maltose/
├── frame/          # 框架核心
│   ├── m/          # 框架入口
│   └── mins/       # 框架实例管理
├── net/            # 网络相关
│   ├── mhttp/      # HTTP 服务实现
│   └── mclient/    # HTTP 客户端
├── database/       # 数据库相关
│   └── mdb/        # 数据库操作封装
├── os/             # 系统相关
│   ├── mcfg/       # 配置管理
│   └── mlog/       # 日志系统
├── util/           # 工具包
│   └── mmeta/      # 元数据工具
└── contrib/        # 扩展集成
    ├── trace/      # 链路追踪
    ├── metric/     # 指标监控
    └── config/     # 配置中心
```

## 最佳工程实践目录结构

```
project/
├── api/            # API 定义和请求/响应结构
│   └── v1/         # API 版本
├── cmd/            # 应用程序入口
│   └── server.go   # 服务器启动
├── config/         # 配置文件
├── internal/       # 内部代码，不对外暴露
│   ├── controller/ # 控制器，处理请求和响应
│   ├── service/    # 服务层，实现业务逻辑
│   ├── dao/        # 数据访问对象，处理数据库操作
│   └── model/      # 模型定义
│       ├── entity/ # 实体定义，映射数据库表结构
│       └── do/     # 数据对象，用于业务数据传输
├── route/          # 路由定义
│   └── route.go    # 路由注册
├── utility/        # 工具函数
└── main.go         # 主入口文件
```

# 功能清单

- [x] mhttp 服务器 - HTTP 服务框架
- [x] mcfg 配置 - 配置管理
- [x] mtrace 链路跟踪 - 分布式追踪
- [x] mclient HTTP 请求 - HTTP 客户端
- [x] mdb 数据库 - 数据库操作封装
- [ ] mqueue 异步任务和队列 - 消息队列
- [ ] mcache 缓存 - 缓存管理
- [ ] mgrpc gRPC 支持 - RPC 服务
- [ ] mcli 脚手架工具 - 基于数据库表结构生成代码

# 贡献指南

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

# 开源许可

`Maltose` 采用 [MIT 许可证](LICENSE) 开源，永久免费。
