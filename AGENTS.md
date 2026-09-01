# Maltose 仓库说明

## 项目范围

Maltose 是基于 Gin 构建的轻量级 Go Web 框架。本仓库包含框架根模块和多个独立发布的嵌套模块。

- 根模块为 `github.com/graingo/maltose`。
- `cmd/maltose/` 是独立发布的 CLI 模块。
- `contrib/config/apollo/`、`contrib/config/nacos/`、`contrib/metric/otlpmetric/`、`contrib/trace/otlptrace/` 和 `contrib/observability/` 是独立发布的模块。
- `container/`、`database/`、`errors/`、`frame/`、`net/`、`os/` 和 `util/` 下的包属于根模块。

修改 `cmd/maltose/`、`contrib/` 或 `.github/` 下的文件前，先读取对应子树中距离目标最近的 `AGENTS.md`。这些规则与本文件共同生效。

## 开发原则

- 根因修复，直接用正确实现替换错误逻辑。不要添加临时补丁、补偿性判断或保留错误行为的兼容分支。
- 代码、注释、文档、错误提示和结论使用直接、肯定、准确的表达。
- 保持控制流清晰。简单行为使用直接逻辑；稳定的业务规则或扩展点再使用策略、组合或工厂。
- 保持包职责清楚，创建新抽象前先复用 Maltose 现有组件。
- 保持公共 API 小而明确。导出标识符、配置键、默认行为、错误语义、生成产物和文档示例都属于兼容契约。
- 有意修改公共契约时，在同一变更中同步更新测试、文档、示例和迁移说明。
- 不要直接修改标记为 `DO NOT EDIT` 的文件；修改生成器或源定义后重新生成。
- 使用 `gofmt` 或 `goimports` 格式化 Go 代码。包名沿用 `mhttp`、`mdb`、`mlog` 等简短小写形式。

## 模块与依赖规则

- 保持所有嵌套模块能够作为独立 Go 模块使用。
- 不要在提交的 `go.mod` 中添加仓库本地 `replace`。
- 验证独立发布模块的依赖解析时使用 `GOWORK=off`。
- 保持依赖方向无环。`contrib/observability` 可以组合 OTLP metric 和 trace 模块；叶子模块不得反向依赖 observability 聚合模块。
- 标准库和现有依赖无法满足需求时再添加生产依赖。
- 框架包不承载具体应用的业务规则。

## 测试与验证

根据变更范围选择验证命令：

- 根框架：`go test -race -coverprofile=coverage.out ./...`
- Lint：`golangci-lint run --verbose`
- CLI 模块：`(cd cmd/maltose && GOWORK=off go test -race -mod=readonly ./...)`
- 使用当前工作区验证全部嵌套模块：`.github/scripts/test-local-modules.sh`
- 独立模块完整性：在每个受影响模块中运行 `GOWORK=off go mod tidy -diff` 和 `GOWORK=off go test -race -mod=readonly ./...`
- Redis 集成测试：`go test -race -tags=integration ./database/mredis ./contrib/cache/redis`

Apollo 和 Nacos 集成测试依赖真实服务。服务不可用时明确说明，不要把对应测试报告为通过。根模块 CI 覆盖率门槛为 75%。

测试文件与被测包放在同一目录，优先使用表驱动测试。测试可观察行为和失败语义，不只验证实现细节。

## Review 要求

开发完成后执行两轮审查，发现问题后先完成修复再交付。

1. 审查业务行为和契约：主流程、生命周期、失败分支、兼容性、并发、事务、重试、数据一致性、边界、性能和敏感数据。
2. 审查设计和可维护性：包职责、命名、控制流、重复代码、无用代码、临时分支、现有能力复用和抽象成本。

测试和静态分析用于支持审查，不能替代人工判断。最终分别说明业务审查、设计审查、验证结果和遗留风险。

## 文档与提交

- 公共 API、配置项、CLI 行为、生成结构或用户工作流发生变化时，更新 `maltose-docs`。
- 推荐应用结构或脚手架生成行为发生变化时，更新 `maltose-quickstart`。
- 提交保持聚焦，使用 `feat:`、`fix:`、`refactor:`、`chore:` 等 Conventional Commit 前缀。
- 默认分支为 `master`。任务未明确要求切换分支时保持当前分支，不要在 Maltose 提交中混入兄弟仓库的改动。
