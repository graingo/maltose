# Maltose CLI 说明

本文件适用于独立发布的 `github.com/graingo/maltose/cmd/maltose` 模块，并补充仓库根目录的 `AGENTS.md`。

## CLI 契约

- 命令名、flag、默认值、位置参数处理、退出行为、生成路径、生成源码和错误信息都属于面向用户的契约。
- 保持 Cobra 命令校验明确，在文件系统或数据库操作开始前拒绝不支持的输入。
- 生成器返回可操作的错误，不要写入不完整或语法无效的 Go 文件。
- 保留已有文件，除非命令明确定义了追加或重新生成行为。
- `cli/` 负责命令装配和输入校验，`internal/gen/` 负责 Go 源码生成，`internal/openapi/` 负责 OpenAPI 解析与生成。

## 生成源码

- 修改 `internal/gen/template.go` 中的模板或对应生成逻辑，不要把手工修改生成结果作为根因修复。
- 保持可编辑产物与标记为 `DO NOT EDIT` 的产物之间的区别。
- 写入或追加生成代码前，先完成格式化和语法解析。
- 模板变化时，验证生成的包名、import、receiver 类型、指针语义、零值处理、文件权限和重复运行行为。
- 推荐生成的应用结构发生变化时，更新 `maltose-quickstart`。

## 验证

在本模块目录运行：

```bash
GOWORK=off go mod tidy -diff
GOWORK=off go test -race -mod=readonly ./...
```

生成器变更需要增加基于临时目录的测试，并检查或编译生成结果。命令变更需要更新命令契约测试。CLI 使用了根模块的新 API 时，从仓库根目录运行 `.github/scripts/test-local-modules.sh`。
