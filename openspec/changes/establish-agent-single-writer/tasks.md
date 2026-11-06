## 1. 纯类型与状态骨架

- [ ] 1.1 将 `pkg/llm` 的消息、工具、ToolCall、Usage 类型拆到纯类型包；验收：`go test ./pkg/llm/...` 且纯类型包不导入 HTTP client。
- [ ] 1.2 引入 `SessionMeta`、`Conversation`、`TurnState`、`RuntimeServices` 与工具执行记录类型；验收：单元测试覆盖默认值与状态转换。
- [ ] 1.3 为持久化数据定义 schema version、原子写入和解析失败处理；验收：保存/加载 round-trip 与损坏文件测试。

## 2. 单写者 Turn Engine

- [ ] 2.1 实现 Session 输入队列，并保证同一 Session 同时只有一个活动 Turn；验收：并发提交两条输入时执行顺序与结果稳定。
- [ ] 2.2 实现显式 Turn 状态机和事件输出，覆盖 LLM、工具、完成、失败与取消；验收：状态机单元测试。
- [ ] 2.3 将 Handler、Compressor 和 Agent 对 Session 的直接写入迁移到 Engine；验收：架构检查确认其他包不再赋值 Conversation/Runtime 字段。
- [ ] 2.4 实现 Turn 取消和进程依赖重新注入；验收：取消测试和重新加载 Session 后的 LLM/emit/queue 可用性测试。

## 3. Inproc RuntimePort

- [ ] 3.1 定义脱敏 `RuntimeSnapshot`、`RuntimeMutation` 和 `RuntimePort`，移除工具对完整 Runtime/Meta 的访问；验收：快照测试确认不含 API Key、LLM client 和事件通道。
- [ ] 3.2 实现 invocation 级 mutation buffer、Engine 校验与原子应用；验收：成功 mutation 顺序和失败回滚测试。
- [ ] 3.3 实现超时/取消后的迟到 mutation 拒绝；验收：慢工具超过超时后 Runtime 保持不变。
- [ ] 3.4 将现有 inproc 工具需要的 Runtime 操作迁移到窄接口；验收：`go test ./internal/tool/inproc/...`。

## 4. 集成与回归

- [ ] 4.1 将 SimpleUI 从每条输入启动 goroutine 改为提交 Session 队列；验收：连续输入集成测试。
- [ ] 4.2 增加并发运行测试并运行 `env GOCACHE=/tmp/agent_cli_go_cache go test -race ./...`；验收：无数据竞争。
- [ ] 4.3 更新架构文档，记录状态所有权、Turn 生命周期和 RuntimePort 边界；验收：文档与架构检查一致。

