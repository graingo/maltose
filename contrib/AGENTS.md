# Maltose Contrib 说明

本文件适用于 `contrib/` 下的适配器和可观测性集成，并补充仓库根目录的 `AGENTS.md`。

## 模块边界

- 每个包含 `go.mod` 的目录都是独立发布、独立使用的模块。
- 不要提交仓库本地 `replace`。
- 每个叶子适配器只负责一个外部系统或协议。
- `contrib/observability` 是 trace 和 metric 集成的组合层。保持发布顺序为根模块、叶子模块、observability。
- 只有集成明确暴露底层能力时，Maltose 公共接口才可以返回供应商专用类型。
- 配置默认值、环境变量行为、生命周期、关闭、重试和错误语义都属于公共契约。

## 外部服务与遥测

- 将可单元测试的适配器行为与真实服务集成测试分开。
- 调用供应商服务时完整传递 `context.Context` 的取消和 deadline。
- 确保客户端、exporter 和后台任务按生命周期关闭。
- 控制遥测数据规模，默认不添加高基数属性。
- 不要记录凭证、token、连接串或包含秘密的配置内容。

## 验证

在每个受影响的嵌套模块中运行：

```bash
GOWORK=off go mod tidy -diff
GOWORK=off go test -race -mod=readonly ./...
```

然后从仓库根目录运行 `.github/scripts/test-local-modules.sh`，使用当前框架源码验证模块，同时保持提交的模块文件不变。

Apollo 和 Nacos 集成测试依赖真实服务。当前 Nacos 真实服务集成测试不使用 `-race`，因为固定版本 SDK 的 reconnect 状态存在已知竞态；可单元测试的适配器代码继续启用竞态检测。
