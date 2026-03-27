name: Bug Report
description: 报告框架异常、错误行为或与预期不符的问题
title: "[BUG] "
labels: ["bug"]
body:

- type: markdown
  attributes:
  value: |
  请尽量提供可复现的信息，这会帮助我们更快定位问题。
  创建后仓库会自动进行 AI 初步分析，请尽量描述清楚。

- type: textarea
  id: description
  attributes:
  label: 问题描述
  description: 请简要说明你遇到了什么问题
  placeholder: 例如：注册中间件后，某些路由请求返回状态码异常
  validations:
  required: true

- type: textarea
  id: steps
  attributes:
  label: 复现步骤
  description: 请尽量写清楚如何复现
  placeholder: | 1. 初始化应用 2. 注册路由和中间件 3. 发起请求 4. 返回结果与预期不一致
  validations:
  required: false

- type: textarea
  id: expected
  attributes:
  label: 期望结果
  description: 你原本期待发生什么
  placeholder: 例如：应正常进入处理函数并返回 200
  validations:
  required: false

- type: textarea
  id: actual
  attributes:
  label: 实际结果
  description: 实际发生了什么
  placeholder: 例如：请求被提前中断，返回 500
  validations:
  required: false

- type: textarea
  id: code
  attributes:
  label: 最小复现代码
  description: 如有可能，请提供最小可复现示例
  render: go
  validations:
  required: false

- type: input
  id: version
  attributes:
  label: 框架版本
  description: 如果知道，请填写版本号
  placeholder: 例如：v0.2.2
  validations:
  required: false

- type: input
  id: go_version
  attributes:
  label: Go 版本
  placeholder: 例如：go1.24.1
  validations:
  required: false

- type: input
  id: environment
  attributes:
  label: 运行环境
  description: 请填写系统或运行环境
  placeholder: 例如：macOS 15 / Ubuntu 24.04 / Docker
  validations:
  required: false

- type: textarea
  id: logs
  attributes:
  label: 日志或报错信息
  description: 如有，请粘贴关键错误信息或 panic 栈
  render: shell
  validations:
  required: false

- type: textarea
  id: extra
  attributes:
  label: 补充说明
  description: 其他有助于定位问题的信息
  validations:
  required: false
