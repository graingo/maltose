<div align="center">
  <a href="https://graingo.github.io/maltose-docs/">
    <img width="128" src="https://graingo.github.io/maltose-docs/logo.svg" alt="Maltose Logo">
  </a>
  <h1>Maltose</h1>
  <p><strong>基于 Gin 打造的轻量级、企业级 Go Web 开发框架</strong></p>
  <p>
    <a href="https://graingo.github.io/maltose-docs/"><strong>查看文档</strong></a> ·
    <a href="https://github.com/Graingo/maltose/issues/new/choose">提交 Bug</a> ·
    <a href="https://github.com/Graingo/maltose/issues/new/choose">功能需求</a>
  </p>
</div>

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/graingo/maltose.svg)](https://pkg.go.dev/github.com/graingo/maltose)
[![Go Report Card](https://goreportcard.com/badge/github.com/graingo/maltose)](https://goreportcard.com/report/github.com/graingo/maltose)
[![License](https://img.shields.io/github/license/graingo/maltose.svg)](LICENSE)

</div>

Maltose 是一款基于 [Gin](https://github.com/gin-gonic/gin) 设计的、旨在提升团队开发效率和代码质量的 Go Web 框架。它通过约定优于配置、清晰的分层架构和强大的代码生成工具，帮助开发者快速构建可维护、可扩展的企业级应用。

## 核心特性

- 🚀 **高性能**: 基于 Gin 核心，保证了卓越的路由和中间件性能。
- 🧱 **模块化设计**: 清晰的模块划分，如 `mhttp`, `mcfg`, `mlog`, `mdb`，易于扩展和替换。
- 🔧 **代码生成**: 内置强大的 CLI 工具，可一键生成 `model`, `dao`, `service` 等业务代码。
- 🔭 **可观测性**: 原生集成 OpenTelemetry，提供开箱即用的链路追踪和指标监控能力。
- 🤝 **Gin 生态兼容**: 完全兼容海量的 Gin 中间件，无缝迁移。
- 📚 **文档完备**: 提供内容详尽、结构清晰的官方文档。

## 快速开始

### 安装

```bash
go get -u github.com/graingo/maltose
```

### 创建项目

我们强烈建议您通过 [maltose-quickstart](https://github.com/graingo/maltose-quickstart) 模板来开始您的新项目，它已经为您配置好了推荐的项目结构。

```bash
# 克隆快速启动模板
git clone https://github.com/graingo/maltose-quickstart.git my-app

# 进入项目目录
cd my-app

# (重要) 修改 go.mod 中的模块名为您自己的项目路径
go mod edit -module "your/module/path"

# 整理依赖
go mod tidy

# 运行
go run main.go
```

## 了解更多

我们强烈建议您阅读 **[官方文档](https://graingo.github.io/maltose-docs/)** 以获得关于框架的完整介绍、核心概念、组件使用和高级特性的详细指南。

## 设计参考

Maltose 在设计上深受优秀的开源项目 [GoFrame](https://github.com/gogf/gf) 的启发，我们对其在模块化、接口化、工程化方面的设计哲学表示敬意和感谢。

与 GoFrame 追求大而全的全功能套件不同，Maltose 的目标是成为一个更加 **轻量级、专注** 的框架。我们选择基于拥有庞大生态和高性能的 [Gin](https://github.com/gin-gonic/gin) 作为 Web 核心，专注于后端 API 的开发场景，致力于为开发者提供一套既有 GoFrame 工程化优点，又能充分享受 Gin 生态红利的开发体验。

## 致谢

Maltose 的诞生离不开以下优秀开源项目的启发和支持：

- **[GoFrame](https://github.com/gogf/gf)**: 提供了卓越的工程化思想和分层架构设计的灵感。
- **[Gin](https://github.com/gin-gonic/gin)**: 提供了稳定、高性能的 Web 核心。

在此向以上项目及其所有贡献者表示诚挚的感谢！

## 贡献

我们欢迎任何形式的贡献！您可以：

- 提交 Bug 或提出功能建议 [Issues](https://github.com/Graingo/maltose/issues/new/choose).
- 通过 [Pull Requests](https://github.com/Graingo/maltose/pulls) 提交您的代码。

## 许可证

MIT 许可证

Enjoy it!

### 结构体转换 (Struct Conversion)

`mconv` 包在 `complex` 子包中提供了一个强大的 `Struct` 函数 (以及它返回错误的版本 `StructE`)。这个工具专为将 `map[string]interface{}` 或其他 `struct` 灵活、高性能地转换为目标 `struct` 而设计。它利用缓存机制来优化重复转换，从而实现显著的速度提升。

**主要特性:**

- **简单转换**: 直接将数据映射到结构体的字段。
- **标签驱动映射**: 使用 `mconv` 标签来映射不同名称的字段。
- **不区分大小写**: 自动匹配源 `map` 中的键和结构体字段，不限制大小写。
- **嵌套结构体**: 递归地转换嵌套的 `map` 或 `struct`。
- **通过钩子扩展**: 提供自定义的 `HookFunc` 函数来处理特殊的转换逻辑。
- **高性能**: 缓存结构体的分析结果，使得后续的转换非常快。

**基础用法**

```go
package main

import (
	"fmt"
	"github.com/graingo/mconv/complex"
)

func main() {
	type User struct {
		ID   int    `mconv:"user_id"`
		Name string `mconv:"user_name"`
	}

	source := map[string]interface{}{
		"user_id":   123,
		"USER_NAME": "Alice", // 不区分大小写匹配
	}

	var user User
	err := complex.StructE(source, &user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", user)
	// Output: {ID:123 Name:Alice}
}
```

**使用钩子 (Hooks)**

您可以使用钩子注入自定义的转换逻辑。例如，将一个整型的状态转换为字符串。

```go
package main

import (
	"fmt"
	"github.com/graingo/mconv/complex"
	"reflect"
)

func main() {
	type Post struct {
		Title  string
		Status string `mconv:"status"`
	}

	// 这个钩子将整型 status 转换为字符串表示
	intStatusToStringHook := func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.Int && to.Kind() == reflect.String {
			i, _ := data.(int)
			switch i {
			case 0:
				return "Draft", nil
			case 1:
				return "Published", nil
			default:
				return "Unknown", nil
			}
		}
		return data, nil
	}

	source := map[string]interface{}{
		"Title":  "Hello World",
		"status": 1,
	}

	var post Post
	err := complex.StructE(source, &post, intStatusToStringHook)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", post)
	// Output: {Title:Hello World Status:Published}
}
```

默认情况下, `mconv` 包含一个内置钩子，用于处理从 `string` 到 `time.Time` 的转换。
