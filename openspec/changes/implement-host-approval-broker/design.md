## Context

宿主执行会移除 Landlock/seccomp，但工具参数由 LLM 决定。审批不能只看工具名，必须绑定规范化后的具体动作。现有代码中的 `NeedConfirm`、`ApprovalPolicy` 和 confirm 响应没有形成状态机，也没有防止参数替换、路径变化、token 重放或 AI 失败后的错误放行。

## Goals / Non-Goals

**Goals:**

- 所有宿主执行和扩展能力先审批后执行。
- 当前用户权限为硬上限。
- 工作区外写入类操作强制用户明确批准。
- 审批 token 与 action manifest 一次性绑定。

**Non-Goals:**

- 不授予 sudo/root 或额外 capabilities。
- 不实现远程 artifact fetch/execute。
- 不实现长期“全部允许”模式。

## Decisions

- **审批发生在 worker 启动前。** Broker 根据工具 descriptor 和参数生成 action manifest，先完成确定性策略、AI/用户审批和 token 签发，再启动 one-shot host worker。worker 不自行申请或提升权限。
- **manifest 采用规范化结构和 hash。** 至少包含 execution mode、argv/action type、cwd、真实读写路径、网络范围、环境变量名称、超时和资源限制。路径在审批与执行前各解析一次，基于真实路径判断工作区边界并检测 symlink/rename 变化。
- **策略按固定顺序执行：硬拒绝 → 用户强制确认 → AI eligible → 自动拒绝。** 硬拒绝不能被 AI 覆盖；工作区外且非 `/tmp` 的写入类操作进入用户强制确认；工作区外读取和 sandbox 网络对 `approve-for-me` eligible。
- **AI approver 是独立最小上下文组件。** 只接收 action manifest、确定性风险分类和必要策略，不接受工具原始输出或完整对话。超时、解析失败、模型不可用均 deny。
- **token 使用进程级随机 secret/HMAC 或内存不可伪造 handle。** token 绑定 manifest hash、Session/Turn/ToolCall ID 与 TTL，并使用 once 语义消费。具体密码学形式属于实现细节，但必须通过重放和替换测试。
- **宿主 worker 保持当前用户但不附加 Landlock/seccomp。** 必须拒绝 euid 0，设置 `NO_NEW_PRIVS`，清理敏感环境，使用新进程组、超时和 rlimit/cgroup，并向审计记录退出状态。
- **审计采用追加记录。** 审批请求、决策来源、manifest、执行参数和结果形成不可静默覆盖的事件链。

## Risks / Trade-offs

- [AI approver 仍可能误判或受注入影响] → 最小上下文、硬拒绝规则、默认 deny、独立测试集和可关闭自动审批共同缓解。
- [路径规范化存在 mount/bind 边界] → 使用真实路径和已配置 root 比较，禁止未识别 mount；后续可增加 openat2/landlock 辅助校验。
- [宿主执行无法回滚副作用] → 高风险写入用户强制确认，先做 preflight，并在执行后记录 side-effect summary。
- [NO_NEW_PRIVS 限制不等于完整攻击面隔离] → 宿主执行只允许当前用户操作，环境清理并禁止高风险提权入口；不宣称宿主 worker 是沙箱。

## Migration Plan

1. 定义 manifest、policy result、approval token 和审计 schema。
2. 实现手动确认流程并默认拒绝宿主执行。
3. 实现 host worker 与 token 消费。
4. 接入独立 AI approver 和 `approve-for-me`。
5. 增加重放、参数替换、symlink、root 与 AI 失败测试。

可先只支持手动审批；在 token 与 worker 验证前不得启用自动批准。

## Open Questions

- AI approver 使用本机模型还是远端模型由部署环境决定；接口必须先支持默认 deny 和超时。
- 环境变量白名单的默认集合需通过实际工具需求确定，不应继承完整 Shell 环境。

