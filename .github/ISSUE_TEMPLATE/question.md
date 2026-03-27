name: Question
description: 提交使用问题、设计疑问或配置相关咨询
title: "[Question] "
labels: ["question"]
body:

- type: markdown
  attributes:
  value: |
  如果你不确定这是不是 bug，也可以先作为 Question 提交。
  创建后仓库会自动进行 AI 初步分析。

- type: textarea
  id: question
  attributes:
  label: 问题内容
  description: 请尽量具体描述你想咨询的问题
  placeholder: 例如：如何为某个路由组单独应用中间件？
  validations:
  required: true

- type: textarea
  id: context
  attributes:
  label: 使用场景
  description: 请说明你的业务场景或上下文
  placeholder: 例如：我想在后台管理路由启用鉴权，但不影响公开接口
  validations:
  required: false

- type: textarea
  id: tried
  attributes:
  label: 已尝试的方式
  description: 你已经试过哪些方式或查阅了哪些资料
  placeholder: 例如：我已经阅读 README 和示例，但仍不清楚推荐做法
  validations:
  required: false

- type: input
  id: version
  attributes:
  label: 框架版本
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

- type: textarea
  id: code
  attributes:
  label: 相关代码或配置
  description: 如有必要，请贴出最小示例
  render: go
  validations:
  required: false

- type: textarea
  id: extra
  attributes:
  label: 补充说明
  description: 其他你觉得有帮助的信息
  validations:
  required: false
