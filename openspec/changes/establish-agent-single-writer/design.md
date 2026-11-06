## Context

当前 `Session` 聚合 `Meta`、`ToolState` 和一个包含 services、conversation、skill state、response 与 context size 的 `Runtime`。多个包直接修改这些字段，SimpleUI 还会为每条输入启动新的 Agent goroutine。状态所有权、并发控制和持久化边界都没有定义，因此任何功能改动都容易跨多个包并产生竞争或冲突。

## Goals / Non-Goals

**Goals:**

- 建立单 Session 单活动 Turn 的执行模型。
- 将持久化状态与进程依赖分离。
- 只允许 Agent Engine 修改 Session 聚合。
- 为 inproc 工具提供脱敏 Snapshot 和事务化 RuntimePort。

**Non-Goals:**

- 不实现 outproc worker、审批或沙箱策略。
- 不恢复中断中的活动 Turn。
- 不改变 LLM API 数据和工具定义格式。

## Decisions

- **采用单写者 Engine，而不是扩大共享锁。** Engine 串行处理输入队列和 Turn；LLM/tool 的异步结果回到 Engine 后再修改状态。替代方案是让所有组件共享 Session mutex，但仍无法约束 mutation 顺序和超时后的迟到写入。
- **拆分 `SessionMeta`、`Conversation`、`TurnState` 和 `RuntimeServices`。** Meta 保存稳定配置；Conversation 保存消息和 usage；TurnState 保存阶段、工具调用和取消；Services 保存 LLM client、队列、事件 sink 和 Context。这样持久化与依赖注入可独立测试。
- **Turn 使用显式状态机。** 路由不再依据 `Runtime.Response != nil` 隐式判断。工具和执行状态的事件由 Engine 发出。
- **输入进入 Session 队列。** API/CLI 提交输入后等待该 Turn 完成，而不是为每条输入启动 Agent 循环。多个 Session 可以并行，同一 Session 不并行。
- **Inproc 使用 `Snapshot + Stage` 事务。** 工具获得不可变脱敏 Snapshot，并把 mutation 暂存到 invocation scope。Engine 在工具成功且未超时后应用，且重新基于最新状态校验。替代方案是提供完整 Runtime 指针和写锁，但无法限制权限与迟到 mutation。
- **纯 LLM 类型与 client 拆分。** `pkg/llm/types` 保存 Message/Tool/ToolCall/Usage；DeepSeek client 独立，`tool.contract` 只依赖纯类型。
- **持久化使用显式 schema 和原子写入。** 第一版保存 Meta、Conversation 和审计数据；不保存 LLM client、queue、output channel、Context 或活动 Turn。旧格式若无法解析，返回明确错误并保留原文件。

## Risks / Trade-offs

- [单写者可能限制未来并行工具调用] → 先保证状态正确，工具 I/O 可在 Engine 控制的并行阶段执行，结果仍由单写者应用。
- [Snapshot 复制成本随上下文增长] → 只复制工具所需字段，并为大型消息提供只读视图或内容引用，避免无差别深拷贝。
- [持久化重构可能丢失旧 Session] → 提供 schema version 和显式迁移/拒绝策略，不静默覆盖无法解析的数据。
- [RuntimePort 抽象过度] → 第一版只实现确有工具需要的 mutation 类型，不提前覆盖所有 Runtime 字段。

## Migration Plan

1. 引入纯类型和状态结构，不改变现有入口。
2. 将 LLM/tool 调用迁移到 Engine 单写者路径。
3. 将 SimpleUI 输入改为队列提交。
4. 将 inproc 工具迁移到 Snapshot/Stage。
5. 切换持久化 schema 并删除旧直接写路径。

每个阶段都保留编译和测试。若状态迁移失败，可暂时维持旧 Session 读取但不允许旧组件继续写新状态。

## Open Questions

- 大型 Conversation Snapshot 是否需要共享只读引用而非复制，应在性能测试后决定。
- 活动 Turn 的崩溃恢复是否需要后续独立 change，本 change 默认不恢复。

