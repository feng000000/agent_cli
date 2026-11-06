## Context

远程脚本的内容可在审批后变化，下载管道又隐藏实际执行字节。即使允许 `approve-for-me` 批准网络和工作区外读取，把网络、任意脚本执行和工作区读取合并为一次动作仍会产生不可接受的数据外传与供应链风险。

## Goals / Non-Goals

**Goals:**

- 将远程获取和执行拆成两个固定、可审计的动作。
- 执行绑定固定 digest 的 immutable artifact。
- 网络通过受控出口并限制目标范围。
- 未知来源默认人工确认。

**Non-Goals:**

- 不证明任意脚本无恶意。
- 不替代系统包管理器的签名机制。
- 不允许 shell 绕过 fetch/run-artifact 规则。

## Decisions

- **禁止网络到解释器的直接管道。** 命令规划层识别并拒绝 `curl`/`wget` 到 shell 或解释器的管道与等价组合。
- **fetch 只写 quarantine。** fetch worker 可以访问已批准网络范围，只写隔离目录，不读取工作区，不继承敏感环境；保存 URL、重定向链、content-type、大小和 digest。
- **run-artifact 无网络。** 只接受 artifact ID + digest，执行前校验内容并防止 symlink/rename 替换。执行使用 sandbox worker，默认不开放网络。
- **网络使用 egress proxy。** proxy 接受目标范围，校验每一跳重定向，记录目标与响应摘要。任意 socket 许可不等价于受控网络。
- **未知来源手动审批。** `approve-for-me` 仅在可信来源、签名有效或固定 digest 命中策略时自动批准。AI 可总结脚本，但不被视为安全证明。
- **artifact store 使用受限目录和生命周期策略。** 所有文件具有不可猜测 ID、0600/0700 权限、大小上限、保留期限和清理任务；失败与取消立即清理不完整文件。
- **执行结果与 script 内容分离。** 大输出保存为受限 artifact，模型中只放入必要摘要或引用。

## Risks / Trade-offs

- [两阶段执行降低便利性] → 提供组合命令但内部仍拆成两个审批/执行阶段，用户能看到真实 digest。
- [动态下载的二级脚本绕过首次检查] → 默认执行无网络；若脚本需要网络，后续 fetch 仍走同一 policy。
- [proxy 实现复杂] → 第一版只支持 HTTPS 和受限域名列表，不支持任意协议。
- [可信 digest 策略可能过期] → digest 策略带 TTL、来源和作用域，过期后回退人工确认。

## Migration Plan

1. 实现 artifact store、digest 与清理。
2. 实现 fetch 与 proxy 目标范围控制。
3. 实现 run-artifact 与无网络 sandbox。
4. 接入审批和 `approve-for-me` 的固定 digest 策略。
5. 禁止旧 shell 下载管道并增加回归测试。

若 proxy 尚未可用，只允许人工审批的 fetch，不允许改为直接开放任意网络。

## Open Questions

- 第一版 egress proxy 是使用独立本地进程还是 HTTP 客户端封装，需要结合 sandbox 网络实现评估。
- 可信来源签名格式需要后续确定；当前先以来源 allowlist + digest + 用户确认实现。

