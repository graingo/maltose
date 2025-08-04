<div align="center">
  <a href="https://graingo.github.io/maltose-docs/">
    <img width="128" src="https://graingo.github.io/maltose-docs/logo.svg" alt="Maltose Logo">
  </a>
  <h1>Maltose</h1>
  <p><strong>Lightweight, Enterprise-grade Go Web Development Framework Built on Gin</strong></p>
  <p>
    <a href="https://graingo.github.io/maltose-docs/"><strong>View Documentation</strong></a> ·
    <a href="https://github.com/Graingo/maltose/issues/new/choose">Report Bug</a> ·
    <a href="https://github.com/Graingo/maltose/issues/new/choose">Request Feature</a>
  </p>
</div>

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/graingo/maltose.svg)](https://pkg.go.dev/github.com/graingo/maltose)
[![Go Report Card](https://goreportcard.com/badge/github.com/graingo/maltose)](https://goreportcard.com/report/github.com/graingo/maltose)
[![GitHub release](https://img.shields.io/github/release/graingo/maltose.svg)](https://github.com/graingo/maltose/releases)
[![Go version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![codecov](https://codecov.io/gh/graingo/maltose/graph/badge.svg?token=jpR53wyp8d)](https://codecov.io/gh/graingo/maltose)
[![License](https://img.shields.io/github/license/graingo/maltose.svg)](LICENSE)

**🇺🇸 English** | **[🇨🇳 简体中文](README.MD)**

</div>

Maltose is a Go Web framework based on [Gin](https://github.com/gin-gonic/gin), designed to improve team development efficiency and code quality. Through convention over configuration, clear layered architecture, and powerful code generation tools, it helps developers quickly build maintainable and scalable enterprise-grade applications.

## Core Features

- 🚀 **High Performance**: Built on Gin core, ensuring excellent routing and middleware performance.
- 🧱 **Modular Design**: Clear module separation like `mhttp`, `mcfg`, `mlog`, `mdb`, easy to extend and replace.
- 🔧 **Code Generation**: Built-in powerful CLI tools for one-click generation of `model`, `dao`, `service` and other business code.
- 🔭 **Observability**: Native OpenTelemetry integration, providing out-of-the-box distributed tracing and metrics monitoring.
- 🤝 **Gin Ecosystem Compatible**: Fully compatible with the vast Gin middleware ecosystem, seamless migration.
- 📚 **Complete Documentation**: Provides comprehensive and well-structured official documentation.

## Quick Start

### Installation

```bash
go get -u github.com/graingo/maltose
```

### Create Project

We strongly recommend starting your new project with the [maltose-quickstart](https://github.com/graingo/maltose-quickstart) template, which has the recommended project structure already configured for you.

```bash
# Clone the quickstart template
git clone https://github.com/graingo/maltose-quickstart.git my-app

# Enter project directory
cd my-app

# (Important) Modify the module name in go.mod to your own project path
go mod edit -module "your/module/path"

# Tidy dependencies
go mod tidy

# Run
go run main.go
```

## Learn More

We strongly recommend reading the **[Official Documentation](https://graingo.github.io/maltose-docs/)** for a complete introduction to the framework, core concepts, component usage, and detailed guides on advanced features.

## Design References

Maltose is deeply inspired by the excellent open-source project [GoFrame](https://github.com/gogf/gf) in its design. We pay respect and gratitude to its design philosophy in modularity, interface-oriented programming, and engineering practices.

Unlike GoFrame's pursuit of a comprehensive full-featured suite, Maltose aims to be a more **lightweight and focused** framework. We chose to build on [Gin](https://github.com/gin-gonic/gin), which has a huge ecosystem and high performance, as our Web core, focusing on backend API development scenarios, and are committed to providing developers with a development experience that has both the engineering advantages of GoFrame and can fully enjoy the Gin ecosystem benefits.

## Acknowledgements

The birth of Maltose would not be possible without the inspiration and support of the following excellent open-source projects:

- **[GoFrame](https://github.com/gogf/gf)**: Provided excellent engineering ideas and inspiration for layered architecture design.
- **[Gin](https://github.com/gin-gonic/gin)**: Provided a stable, high-performance Web core.

We hereby express our sincere gratitude to the above projects and all their contributors!

## Contributing

We welcome contributions of any kind! You can:

- Submit bugs or feature suggestions via [Issues](https://github.com/Graingo/maltose/issues/new/choose).
- Submit your code through [Pull Requests](https://github.com/Graingo/maltose/pulls).

## License

MIT License