## Context

工具接口已移动到 `internal/tool/inproc`，接口 package 名已改为 `inproc`，但 ToolState、Handler、Executor 和测试入口仍引用旧路径。`cmd/executor` 中还有一段未完成的 `runTool`，即使 import 修复后仍会因未定义变量导致编译失败。由于后续所有架构 change 都依赖稳定构建，必须先完成这一层机械迁移。

## Goals / Non-Goals

**Goals:**

- 统一使用 `agentcli/internal/tool/inproc` 和 `agentcli/internal/tool/inproc/toolimpl`。
- 恢复 `go test ./...` 的代码编译阶段。
- 用架构检查阻止旧路径回流。

**Non-Goals:**

- 不重构 `InternalToolContext` 或其 Runtime 权限。
- 不实现 executor 的沙箱、协议或工具执行能力。
- 不修复 Session 持久化、并发输入或 LLM 逻辑。

## Decisions

- **目标路径固定为 `internal/tool/inproc`。** 不再保留旧 `internal/tool` 兼容包，避免同一工具存在两个 import 路径。替代方案是增加兼容 shim，但会让后续边界检查更难执行。
- **保持接口 package 名为 `inproc`。** import path 与 package 名保持一致，工具实现直接引用 `inproc.InternalToolContext`。
- **`cmd/executor` 暂时返回明确的未实现错误。** 当前 `runTool` 没有有效的 Runtime/Meta 上下文，直接伪造上下文会隐藏设计问题。本 change 只保证代码可编译，真实执行由 `implement-outproc-sandbox-execution` 负责。
- **`libseccomp` 作为构建前置条件。** 包迁移完成后，若 `pkg/confine` 仍因系统依赖失败，任务中记录安装与 CI 验证命令，不用 build tag 绕过安全代码。

## Risks / Trade-offs

- [临时未实现错误看起来像功能回退] → 当前 executor 本来返回空响应且不可用，明确错误比空成功更安全。
- [全仓机械替换可能误改字符串或注释] → 只替换 import path 和 package 引用，执行 `go test` 与静态检索验证。
- [Linux 构建依赖在开发机缺失] → 明确记录 libseccomp 环境要求，并在 CI 容器中固定安装。

## Migration Plan

先迁移接口和 toolimpl，再迁移调用方，最后修复 executor 编译占位并运行全仓测试。若迁移导致无法编译，可回滚单个文件的 import 调整，但不得恢复旧包长期共存。

