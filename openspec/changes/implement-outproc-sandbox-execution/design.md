## Context

当前 executor 同时扮演 HTTP server、工具注册表和未完成的执行器，接口只传工具名和参数，且没有区分 inproc/outproc。现有 seccomp 仅拦截 IPv4/IPv6 socket，Landlock 在空可读列表时允许全盘只读，并在 ABI 不支持时使用 BestEffort，不能支持不可信工具的安全承诺。

## Goals / Non-Goals

**Goals:**

- 提供明确的 outproc 工具分类、能力声明、协议和独立 sandbox worker。
- 默认最小文件读取、工作区/`tmp` 写入和无网络。
- enforcement 失败时拒绝执行。
- 终止超时/取消的完整进程树。

**Non-Goals:**

- 不实现宿主执行或审批 token。
- 不实现远程 artifact fetch/execute。
- 不保证静态工具参数能够预测任意脚本的所有系统调用。

## Decisions

- **主进程 supervisor 只派生 one-shot worker。** 每次调用创建新进程以应用不可逆的单次策略。替代方案是长期沙箱 server，但不同请求需要不同能力时无法安全放宽 Landlock/seccomp。
- **ToolDescriptor 携带分类与能力，而非由图路径推断。** 缺少分类或能力的工具拒绝路由；模型只能调用系统注册的 descriptor。
- **协议使用强类型 DTO 和 request ID。** 请求包含 Session/Turn/ToolCall ID、工具版本、规范化参数、执行等级、能力、超时和资源限制；结果包含状态、内容、artifact、错误码和时长。
- **默认文件策略是只读工作区、`tmp` 和运行时白名单，只写工作区与 `tmp`。** 运行时白名单由宿主配置，不接受工具自行扩展。临时结果写入受限目录并校验权限。
- **fail-closed。** Landlock/seccomp ABI 不足、过滤器加载失败或策略无法验证时，worker 在工具执行前退出。只读全盘和 BestEffort 不再用于不可信 outproc。
- **基础 seccomp 防护与脚本执行能力分离。** 默认限制 ptrace、跨进程内存访问和跨进程组信号；允许 fork/exec 的 worker 仍受进程组、cgroup、rlimit 和超时控制。更完整的 syscall allowlist 在实现前用威胁模型测试补齐。
- **网络策略由能力注入，proxy 实现属于远程 artifact change。** 本 change 可以拒绝网络或接受已批准网络能力，但不接受工具自行抬高能力。
- **架构测试强制 import 边界。** `outproc`、`protocol` 和 `confine` 不得引用 Agent/Session/Runtime 具体类型。

## Risks / Trade-offs

- [one-shot worker 启动成本高] → 优先保证隔离与策略正确，后续基于基准测试评估 warm pool；warm pool 不得跨策略复用。
- [seccomp 默认允许全部 syscall 仍暴露攻击面] → 先禁止高风险进程操作，建立恶意 fixture 测试，再逐步收紧 allowlist。
- [Landlock 只控制路径访问，不控制 Unix Socket 对端] → 同时限制可访问 socket 路径、进程信号与 ptrace。
- [脚本需要网络时能力边界复杂] → 本 change 保持网络默认拒绝，远程脚本由后续 artifact change 通过 proxy 处理。

## Migration Plan

1. 引入 contract、registry 和协议 DTO，保留 inproc 路径。
2. 实现 worker launcher、fail-closed policy 和超时终止。
3. 将现有文件工具迁移到 outproc sandbox。
4. 删除 executor 的空执行 handler 和旧 socket API。
5. 增加架构检查与恶意工具夹具。

若新 worker 尚不稳定，可不让高风险工具进入 registry，但不能回退到 main process 直接执行。

## Open Questions

- 第一版是否支持 fork/exec 脚本工具，需要在恶意 fixture 和资源限制实现后确认。
- cgroup v2 是否作为硬前置条件，还是先用 rlimit/process group 并记录能力缺口。

