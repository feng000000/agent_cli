## Why

当前 Session、Runtime、Handler、Compressor 和 inproc Tool 都能直接修改同一份可变状态，而 SimpleUI 可以为每条输入并发启动 Agent 循环。这会造成消息、LLM 响应、Usage、上下文大小和工具状态竞争，也使“只有 inproc 能操作 runtime/context”无法形成可验证的安全边界。

## What Changes

- **BREAKING**：将 Session 状态拆分为 `SessionMeta`、`Conversation`、`TurnState`、`RuntimeServices`、`ToolRegistry` 和工具执行记录。
- Agent Engine 成为 Session 状态的唯一写入者；新用户输入进入队列，同一 Session 同时只执行一个 Turn。
- 引入显式 Turn 状态机，覆盖等待 LLM、等待工具、等待审批、执行工具、完成、失败和取消。
- **BREAKING**：inproc 工具不再持有完整 Runtime/Meta 指针，只接收脱敏 Snapshot，并通过 `RuntimePort` 暂存 mutation。
- Engine 仅在工具成功且调用未超时时原子应用 mutation；超时或失败的 mutation 必须丢弃。
- 拆分纯 LLM 协议类型与 DeepSeek 客户端，保证 `tool.contract` 可依赖无副作用的类型包。
- 统一持久化边界：Conversation 和审批/工具审计可持久化，LLM client、channel、Context 和队列不进入持久化模型。

## Capabilities

### New Capabilities

- `agent-session-lifecycle`: 定义单 Session/single-writer、输入排队、Turn 状态转换、取消、保存与加载行为。
- `inproc-runtime-mutation`: 定义 inproc 工具通过 Snapshot/Stage 修改 Runtime/context 的受控事务语义。

### Modified Capabilities

无。仓库当前没有已归档的行为 spec。

## Impact

- 受影响模块：`internal/agent`、`internal/session`、`internal/session/runtime`、`internal/session/toolstate`、`internal/handler`、`internal/tool/inproc`、`internal/tool/contract`、`pkg/llm`。
- 内部 API 和状态所有权发生破坏性变化；CLI 外部行为应保持兼容，但并发输入会改为排队执行。
- 增加协议类型包和架构依赖检查；不引入新的外部服务。
- 持久化格式需要版本化，旧格式的兼容或迁移策略在 design 中确定。
- 安全影响：不再向工具暴露 API Key、LLM client 或完整 Runtime；超时后的迟到 mutation 不再生效。

## Assumptions

- `repair-inproc-package-boundary` 已完成，仓库可编译。
- Turn 状态只保证单进程执行；活动 Turn 的跨重启恢复不是本 change 的必要能力。
- inproc 工具仍是同进程可信代码，但其 Runtime 权限必须通过窄接口收敛。

## Non-Goals

- 不实现 outproc sandbox、网络代理或宿主审批 broker。
- 不改变工具是否危险的分类规则。
- 不重写 DeepSeek API 行为，只拆分纯类型与客户端依赖。
