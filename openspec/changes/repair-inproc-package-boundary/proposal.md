## Why

仓库正在把工具层从 `internal/tool` 迁移到 `internal/tool/inproc`，但 package 名、import 路径和调用签名只完成了一部分，导致当前工作区无法通过 `go test ./...`。在开展 Runtime、沙箱和审批重构前，必须先建立一个可编译、可验证的基线。

## What Changes

- 将工具接口、工具实现、ToolState、Handler、Executor 和测试入口统一迁移到 `internal/tool/inproc`。
- 修正 `package tool` 与 import path 不一致的问题，补齐 executor 中未完成的工具调用适配。
- 只做路径、package、类型和调用签名的机械调整，不改变工具的运行时行为、沙箱策略或审批语义。
- 增加架构检查，禁止新增对旧 `agentcli/internal/tool` 与 `agentcli/internal/tool/toolimpl` 的引用。
- 确认当前 Linux 环境下的 `libseccomp` 构建依赖；若环境暂时缺少该依赖，应在任务与 CI 中明确安装要求，而不是绕过类型检查。

## Capabilities

### New Capabilities

无。本 change 是纯包迁移和编译修复，不引入可观察的行为变化。

### Modified Capabilities

无。

## Impact

- 受影响模块：`internal/tool/inproc`、`internal/tool/inproc/toolimpl`、`internal/session/toolstate`、`internal/handler`、`internal/tool/executor`、`cmd/executor`、`cmd/test`。
- 不修改 Session/Runtime 数据结构、工具路由规则、用户接口、配置格式或持久化数据。
- 不增加新依赖；继续使用现有 Landlock/seccomp 依赖。
- 迁移完成后，整个仓库应能进入后续架构 change 的可编译起点。

## Assumptions

- `internal/tool/inproc` 是后续进程内工具的唯一目标路径。
- 本 change 不重新设计 `InternalToolContext`、Runtime 锁或工具审批流程。
- 保留现有工具定义与执行逻辑，只修复组织边界和构建错误。

## Non-Goals

- 不引入 RuntimePort、Snapshot、Mutation 或单写者 Agent Engine。
- 不实现 outproc worker、沙箱 fail-closed 策略或宿主审批。
- 不修复与包迁移无关的持久化、并发或 LLM 逻辑缺陷。
