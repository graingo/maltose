# Maltose CI 与发布说明

本文件适用于 workflow 和发布脚本，并补充仓库根目录的 `AGENTS.md`。

## CI 变更

- 保持对根模块、使用当前工作区的全部嵌套模块、作为独立发布消费者的全部嵌套模块进行验证。
- 保持 `GOWORK=off`、`go mod tidy -diff`、`-mod=readonly`、竞态检测、lint、漏洞扫描和根模块 75% 覆盖率门槛，除非任务明确修改这些策略。
- 保持 Redis、MySQL、Apollo 和 Nacos 服务装配与使用它们的测试一致。
- 多步骤脚本使用严格 Shell 模式，成功和失败时都执行资源清理。
- 第三方 Action 和工具使用明确版本。审查其权限，没有具体需求时不扩大 token 权限。

## 发布变更

- 保持发布流程可恢复，并验证不可变 tag 的内容。
- 保持发布顺序为根模块、一级模块、`contrib/observability`。
- 发布前验证所有模块都不包含仓库本地 `replace`。
- tag 和 GitHub Release 是外部可见的不可变产物。修改发布状态前先解析准确的版本和目标。
- 保持发布操作幂等，使失败任务能够安全恢复。

## 验证

修改 Shell 脚本后运行语法检查，并执行其可用的只读验证模式。修改 workflow 后，对照 `.github/workflows/test.yaml` 和 `.github/workflows/release.yml` 检查本地命令顺序，确认现有模块和服务仍然全部包含在验证范围内。
