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

## 工程实践与代码生成

Maltose 提倡关注点分离 (SoC) 和约定优于配置 (CoC) 的原则。为了在提升开发效率的同时保持代码结构清晰、易于维护，我们设计了一套受 GoFrame 启发的工程结构，并提供了强大的代码生成工具 `maltose gen dao`。

### 推荐目录结构

```
project-name/
├── api/                # API 接口定义 (对外暴露的 Request/Response)
├── internal/           # 核心业务逻辑 (不对外暴露)
│   ├── controller/     # 接口层: 负责请求校验、调用服务、返回响应
│   ├── model/          # 业务模型 (Service 层的输入/输出，手动维护)
│   │   ├── entity/     # 数据模型 (由工具生成，与数据库表一一对应)
│   │   └── do/         # 领域对象 (手动维护，用于复杂查询或自定义数据结构)
│   ├── service/        # 业务层: 负责业务逻辑的实现与封装
│   ├── dao/            # 数据访问层 (由工具首次生成，可手动扩展)
│   │   └── internal/   # 内部 DAO (由工具生成，包含基础 CRUD)
│   └── ...             # 其他业务相关目录
├── .env                # 配置文件 (用于数据库连接等)
└── main.go             # 项目入口
```

### 分层设计哲学

这套结构的核心思想是 **将自动生成的代码与开发者手写的代码安全地分离开**，使得您可以随时重新生成代码以适应数据库变更，而不用担心覆盖自己的业务逻辑。

#### Model 层详解

- **`internal/model/entity/`**

  - **来源**: 由 `maltose gen dao` 工具根据数据库表结构自动生成。
  - **用途**: 作为 GORM 的模型，与数据库表严格一一对应。
  - **规则**: **不应手动修改**。文件头部有 `DO NOT EDIT` 标识。

- **`internal/model/do/` (Domain Object)**

  - **来源**: 开发者手动创建。
  - **用途**: 定义非标准的数据结构。最常见的场景是存放 **多表联合查询** 的结果。
  - **规则**: 完全由开发者维护。

- **`internal/model/`**
  - **来源**: 开发者手动创建。
  - **用途**: 定义 **业务模型**，作为 `Service` 层公共方法的输入（Input）和输出（Output）参数。它构成了 `Controller` 和 `Service` 之间的"胶水层"，实现业务逻辑与特定接口的解耦。
  - **规则**: 完全由开发者维护。`Service` 层应避免直接使用 `api` 或 `entity` 中的结构体。

#### DAO 层详解

- **`internal/dao/internal/`**

  - **来源**: 由 `maltose gen dao` 工具自动生成。
  - **用途**: 包含了每张表最基础的 CRUD 方法。
  - **规则**: **不应手动修改**，每次代码生成时都会被覆盖。

- **`internal/dao/`**
  - **来源**: 由工具在 **第一次** 生成时创建。
  - **用途**: 封装 `internal` 中的 DAO，并作为业务的统一数据访问入口。**所有自定义的数据库操作**（如复杂的链式查询、手写 SQL 等）都应在此层中添加为新方法。
  - **规则**: **可以随意修改和扩展**。由于工具不会覆盖此目录下的已有文件，您的自定义逻辑是安全的。

通过这种设计，开发者可以将精力聚焦在 `service` 层的业务逻辑、`dao` 层的复杂查询以及 `controller` 层的接口处理上，而将繁琐的基础模型和 DAO 代码交给 `maltose` 工具自动完成。

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
