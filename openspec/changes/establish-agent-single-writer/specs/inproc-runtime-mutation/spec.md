## ADDED Requirements

### Requirement: Inproc 工具只能通过受控 RuntimePort 操作状态

系统 MUST 向 inproc 工具提供脱敏 Snapshot 和受控 mutation 端口，不得向工具暴露 API Key、LLM client、输出 channel 或完整 Runtime 指针。

#### Scenario: 工具接收脱敏快照

- **WHEN** Engine 调用 inproc 工具
- **THEN** 工具只能读取其所需 Session 标识、工作区能力、已加载 Skill 和显式提供的上下文数据
- **AND** 快照不包含凭据、LLM client 或内部事件通道

### Requirement: Mutation 仅在成功且未超时时应用

系统 MUST 在 inproc 工具成功返回且调用上下文仍有效时，按确定顺序应用暂存 mutation；工具失败、panic 或超时时 MUST 丢弃全部 mutation。

#### Scenario: 成功提交 Skill mutation

- **WHEN** inproc Skill 工具成功暂存加载 Skill 的 mutation
- **THEN** Engine 在校验 Skill 存在后应用 mutation，并在后续 LLM 请求中反映新状态

#### Scenario: 超时后丢弃 mutation

- **GIVEN** inproc 工具已暂存修改上下文的 mutation
- **WHEN** 工具在 Engine 超时后仍未正常返回
- **THEN** Engine 丢弃该工具的全部 mutation，Runtime/context 保持不变

#### Scenario: 工具执行失败

- **WHEN** inproc 工具返回错误
- **THEN** Engine 不应用该工具的任何已暂存 mutation，并返回可识别的工具错误

### Requirement: Snapshot 在同一工具调用内保持一致

系统 MUST 为单次 inproc 工具调用提供一致性 Snapshot；工具执行期间的并发状态变化不得被该工具观察到。

#### Scenario: 工具执行期间其他 Turn 状态变化

- **WHEN** 工具基于开始时 Snapshot 执行且 Session 状态随后发生变化
- **THEN** 该工具仍只看到原 Snapshot，最终 mutation 由 Engine 基于最新状态重新校验

