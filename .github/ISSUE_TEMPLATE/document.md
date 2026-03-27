name: Documentation
description: 提交文档错误、缺失、示例不准确或说明不清晰的问题
title: "[Doc] "
labels: ["documentation"]
body:

- type: markdown
  attributes:
  value: |
  如果你发现 README、使用说明、示例代码或 API 文档内容有误，请在这里提交。
  创建后仓库会自动进行 AI 初步分析。

- type: textarea
  id: doc_problem
  attributes:
  label: 文档问题描述
  description: 请说明文档哪里有问题
  placeholder: 例如：README 中的路由注册示例与当前 API 不一致
  validations:
  required: true

- type: input
  id: location
  attributes:
  label: 文档位置
  description: 请尽量提供文件名、章节名或链接
  placeholder: 例如：README.md / 快速开始 / Middleware 章节
  validations:
  required: false

- type: textarea
  id: current
  attributes:
  label: 当前内容或现象
  description: 如有必要，请贴出当前文档内容或说明问题
  validations:
  required: false

- type: textarea
  id: expected
  attributes:
  label: 你期望的内容
  description: 你希望文档如何修改或补充
  placeholder: 例如：补充路由组嵌套和中间件执行顺序的说明
  validations:
  required: false

- type: textarea
  id: extra
  attributes:
  label: 补充说明
  description: 其他你觉得有帮助的信息
  validations:
  required: false
