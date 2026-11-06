## ADDED Requirements

### Requirement: 禁止下载后直接执行

系统 MUST 拒绝 `curl | bash`、`wget -O- | sh` 以及将网络响应直接交给 shell 或解释器的等价操作。

#### Scenario: 下载管道执行

- **WHEN** 工具请求通过 shell 管道下载并执行远程内容
- **THEN** 策略层拒绝执行并返回明确的禁止原因

### Requirement: 远程内容先隔离为 artifact

系统 MUST 将远程内容先写入 quarantine，记录来源 URL、重定向、内容类型、大小和 SHA-256 digest；fetch 阶段 MUST NOT 执行内容。

#### Scenario: 成功获取脚本

- **WHEN** fetch 请求通过受控网络获得脚本
- **THEN** 系统保存 immutable artifact、计算 digest，并将执行权限留给后续独立审批

#### Scenario: 下载超过大小限制

- **WHEN** 响应内容超过配置上限
- **THEN** 系统中止下载、删除不完整 artifact，并返回限制错误

### Requirement: 执行固定 digest 的本地 artifact

系统 MUST 仅执行已经固定来源和 digest 的本地 artifact，并在执行前重新校验 digest；脚本执行 worker MUST 默认无网络。

#### Scenario: digest 一致

- **WHEN** 本地 artifact 的 digest 与已批准 manifest 一致且调用未超时
- **THEN** 系统在无网络 sandbox 中执行该 artifact

#### Scenario: digest 不一致

- **WHEN** 执行前发现 artifact 内容或 digest 发生变化
- **THEN** 系统拒绝执行并要求重新获取和审批

#### Scenario: artifact 被 symlink 替换

- **WHEN** 审批后的路径被 symlink 或重命名替换
- **THEN** 系统拒绝执行，不跟随替换后的目标

### Requirement: 网络访问经受控出口

系统 MUST 通过受控 egress proxy 或等价目标范围机制限制网络访问；仅有“允许任意 IPv4/IPv6 socket”不视为满足网络范围控制。

#### Scenario: 目标在 allowlist 内

- **WHEN** fetch 请求目标匹配已批准域名或地址范围
- **THEN** proxy 允许请求并记录目标、重定向和响应摘要

#### Scenario: 目标超出网络授权

- **WHEN** 请求或后续重定向超出已批准网络范围
- **THEN** proxy 阻断请求且不返回远程内容

### Requirement: 未知远程代码默认人工批准

系统 MUST 对未知来源的下载后执行默认要求用户确认；`approve-for-me` 仅在可信来源、签名有效或策略明确允许固定 digest 时自动批准。

#### Scenario: 未知来源脚本

- **WHEN** artifact 来源未在可信列表且没有有效签名
- **THEN** 系统要求用户确认具体 digest 和执行参数

#### Scenario: 可信固定 digest

- **WHEN** artifact 来自可信来源、digest 与策略记录一致且用户已启用 `approve-for-me`
- **THEN** AI approver 可在策略允许范围内批准执行

### Requirement: Artifact 生命周期受控

系统 MUST 为 quarantine、tool result 和下载 artifact 设置访问权限、保留期限和清理规则，并在失败、超时或取消后清理临时文件。

#### Scenario: 执行结束后清理

- **WHEN** artifact 不再被审计或显式保留
- **THEN** 系统在保留期结束后删除该 artifact 及临时元数据

