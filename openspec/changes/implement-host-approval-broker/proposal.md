## Why

outproc 工具在高风险场景下需要当前用户权限的宿主执行，但当前代码只有未使用的 `NeedConfirm`、`ApprovalPolicy` 和确认响应类型，没有审批状态机、精确 action manifest、一次性授权或审计。没有这些约束，模型生成的工具参数无法安全地转化为宿主操作。

## What Changes

- 新增 Approval Broker 和宿主 one-shot worker；宿主 worker 仅以当前非 root 用户执行，不附加 Landlock/seccomp，但必须启用 `NO_NEW_PRIVS`、清理敏感环境并使用资源限制。
- 宿主执行始终需要审批；支持手动用户确认和用户启用的 `approve-for-me` AI 确认。
- `approve-for-me` 可批准工作区外读取、敏感路径读取和 sandbox 网络开放；沙箱默认读取仍为工作区、`/tmp` 和运行时路径。
- 工作区外且非 `/tmp` 的所有写入类操作，包括 create、write、truncate、rename、delete、chmod、chown、symlink，永远要求用户明确批准，AI 不得单独批准。
- 拒绝 sudo/root 或任何提高 OS 权限的请求；宿主执行上限固定为当前用户权限。
- 定义规范化 action manifest、真实路径解析、symlink/TOCTOU 防护、一次性 token、TTL、tool call/session/turn 绑定和重放拒绝。
- 定义审批状态机：pending、requested、approved、denied、expired、executing、completed、failed；AI 失败或超时默认拒绝。
- 记录用户/AI 决策、manifest hash、argv、cwd、环境白名单、退出码和资源消耗。

## Capabilities

### New Capabilities

- `host-execution-approval`: 定义宿主执行边界、审批策略、AI 自动审批、权威限制、token 与审计行为。

### Modified Capabilities

无。仓库当前没有已归档的审批 spec。

## Impact

- 受影响模块：新增 `internal/approval`、`internal/protocol`、host worker 命令；修改 outproc 路由、CLI 输出、Session TurnState 和审计持久化。
- 引入宿主执行进程、审批 token 和审批审计格式；无数据库迁移。
- 安全影响：所有宿主执行均需批准；AI 自动审批可覆盖读取和网络，但不能覆盖工作区外写入类操作。
- 依赖 `implement-outproc-sandbox-execution` 提供执行等级和能力请求协议。

## Assumptions

- 当前用户不是 root 或 euid 为 0；运行时仍须拒绝 root。
- `approve-for-me` 是用户显式启用的策略，不是工具自行声明的模式。
- AI approver 使用独立、最小化上下文，只接收 action manifest 与确定性策略结果。

## Non-Goals

- 不分配 sudo、root、额外 capabilities 或 setuid 权限。
- 不实现远程 artifact 下载与执行策略。
- 不提供长期全局“永远允许任意命令”绕过。
