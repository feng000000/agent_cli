## Why

当前 executor 只有未完成的 Unix Socket/HTTP 骨架，执行接口返回空结果，工具注册表也没有统一区分 inproc 与 outproc。高风险工具和未来可插拔工具缺少可信的进程隔离、能力声明和 fail-closed 沙箱策略。

## What Changes

- 引入统一 ToolDescriptor，显式区分 `inproc` 与 `outproc`，并声明文件、网络、子进程和宿主执行能力。
- 实现 outproc 协议，包含请求/会话/Turn/ToolCall ID、工具版本、规范化参数、执行等级、能力请求、超时和资源限制。
- 宿主 supervisor 为每次 outproc sandbox 调用启动独立 worker，不执行工具代码。
- outproc 默认只读工作区、`/tmp` 和受控运行时路径；只写工作区与 `tmp`，默认禁止外网。
- Landlock/seccomp 缺失、策略应用失败或 enforcement 降级时拒绝执行，不运行降级后的高风险工具。
- 完成 executor 的列表、执行、错误、超时、结果大小限制和进程树终止流程。
- 为不可信工具增加基础 syscall 防护，限制 ptrace、跨进程内存访问和超出所属进程组的信号；需要子进程的工具通过明确能力启用。
- 建立架构检查，outproc 代码不得导入 Agent、Session 或 Runtime。

## Capabilities

### New Capabilities

- `outproc-tool-execution`: 定义 outproc 工具的分类、能力声明、协议、sandbox worker 生命周期和 fail-closed 执行行为。

### Modified Capabilities

无。仓库当前没有已归档的工具执行 spec。

## Impact

- 受影响模块：`internal/tool/contract`、`internal/tool/registry`、`internal/tool/outproc`、`internal/tool/executor`、`cmd/executor`、`pkg/confine`、Handler/ToolState 路由。
- 引入独立 worker 协议和进程生命周期；现有 Unix Socket executor 骨架会被替代或重构。
- 沙箱依赖 Linux Landlock/seccomp 与 libseccomp；不支持或无法强制时 outproc 工具不可执行。
- 无数据迁移；新增能力清单、worker 审计和 artifact 引用格式。
- 安全影响：高风险工具不再与 Agent 同进程，不能访问 Runtime，默认不能读取工作区外文件或访问外网。

## Assumptions

- `repair-inproc-package-boundary` 与 `establish-agent-single-writer` 已提供稳定的工具路由和状态接口。
- 宿主执行审批属于后续 `implement-host-approval-broker`，本 change 只允许运行 sandbox outproc。
- 允许子进程的工具仍处于同一 sandbox 策略和资源限制内。

## Non-Goals

- 不实现宿主审批、`approve-for-me` 或提权 token。
- 不实现远程脚本下载、quarantine 或 digest 审批。
- 不提供 Windows/macOS 沙箱降级实现。
