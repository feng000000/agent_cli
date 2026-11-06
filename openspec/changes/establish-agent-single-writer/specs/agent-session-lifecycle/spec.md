## ADDED Requirements

### Requirement: Session 同时只执行一个 Turn

系统 MUST 将同一 Session 的并发用户输入排队，并按接收顺序逐个执行 Turn；第二个输入不得在第一个 Turn 完成前启动新的 Agent 循环。

#### Scenario: 连续输入进入队列

- **WHEN** 第一个 Turn 仍在等待 LLM 或执行工具时用户提交第二条输入
- **THEN** 第二条输入保持在队列中，且不会启动第二个并发 Agent 循环
- **AND** 第一个 Turn 完成后系统按顺序开始处理第二条输入

#### Scenario: Turn 完成后释放执行权

- **WHEN** 当前 Turn 成功、失败或被取消
- **THEN** 系统清理该 Turn 的临时状态并将 Session 恢复到可接受下一输入的状态

### Requirement: Turn 生命周期可观察

系统 SHALL 使用显式 Turn 状态表示准备、等待 LLM、执行工具、完成、失败和取消，并通过事件输出报告状态变化。

#### Scenario: 工具执行阶段转换

- **WHEN** LLM 响应包含有效工具调用
- **THEN** 系统依次进入工具执行状态并在所有工具结果写回后重新等待 LLM

#### Scenario: Turn 取消

- **WHEN** 用户取消正在执行的 Turn
- **THEN** 系统停止发起新的 LLM 请求和工具调用并报告取消结果

### Requirement: 状态与进程依赖分离

系统 MUST 将可持久化会话状态与 LLM client、输入队列、输出 channel、取消 Context 等进程依赖分开；加载 Session 时 MUST 重新注入进程依赖。

#### Scenario: 保存并重新加载 Session

- **WHEN** 系统保存包含消息和 usage 的 Session 并在新进程中加载
- **THEN** 对话历史与可持久化元数据被恢复
- **AND** 进程依赖被重新创建且消息发送、LLM 调用和取消仍然可用

#### Scenario: 持久化失败

- **WHEN** 原子写入或 schema 解析失败
- **THEN** 系统返回可识别的错误且不把部分写入内容视为有效 Session

