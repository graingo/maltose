name: Feature Request
description: 提交新功能建议、API 设计建议或框架能力改进建议
title: "[Feature] "
labels: ["enhancement"]
body:

- type: markdown
  attributes:
  value: |
  请尽量先描述问题，再描述你希望的方案。
  创建后仓库会自动进行 AI 初步分析。

- type: textarea
  id: problem
  attributes:
  label: 你想解决的问题
  description: 请先说明当前痛点或使用障碍
  placeholder: 例如：目前缺少统一的请求上下文扩展机制，导致插件开发不方便
  validations:
  required: true

- type: textarea
  id: proposal
  attributes:
  label: 你建议的方案
  description: 请说明你希望新增什么能力或改进什么 API
  placeholder: 例如：增加 Context 扩展接口，允许用户在请求生命周期中挂载自定义数据
  validations:
  required: false

- type: textarea
  id: alternatives
  attributes:
  label: 替代方案
  description: 你是否考虑过其他实现方式
  placeholder: 例如：也可以通过 middleware 注入，但侵入性较高
  validations:
  required: false

- type: textarea
  id: compatibility
  attributes:
  label: 兼容性影响
  description: 该建议是否可能影响现有 API 或行为
  placeholder: 例如：希望保持现有 Handler 签名不变
  validations:
  required: false

- type: textarea
  id: benefits
  attributes:
  label: 预期收益
  description: 这项改动会帮助解决什么问题
  placeholder: 例如：降低插件开发成本，提升可扩展性
  validations:
  required: false

- type: textarea
  id: extra
  attributes:
  label: 补充说明
  description: 其他你觉得有帮助的信息
  validations:
  required: false
