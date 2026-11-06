## ADDED Requirements

### Requirement: 宿主执行始终需要审批

系统 MUST 在任何移除 Landlock/seccomp 的宿主执行前获得审批；未经批准的宿主执行请求 MUST 不启动宿主 worker。

#### Scenario: 用户批准宿主执行

- **WHEN** 宿主 action manifest 已展示且用户选择批准
- **THEN** Broker 签发绑定该 manifest 的一次性 token 并启动宿主 one-shot worker

#### Scenario: 用户拒绝宿主执行

- **WHEN** 用户拒绝 action manifest
- **THEN** 系统不启动宿主 worker并返回 denied

### Requirement: 宿主执行权限上限为当前用户

系统 MUST 拒绝 root、sudo、setuid 提权和额外 OS capabilities；宿主 worker MUST 以当前非 root 用户运行并启用 `NO_NEW_PRIVS`。

#### Scenario: 请求 sudo

- **WHEN** action manifest 请求 sudo、root 或 setuid 提权
- **THEN** Broker 在任何审批前硬性拒绝该请求

#### Scenario: 当前进程为 root

- **WHEN** Agent 或 Broker 以 root 身份运行
- **THEN** 系统拒绝宿主执行并报告不支持的权限状态

### Requirement: Approve-for-me 的 AI 审批边界

系统 SHALL 在用户启用 `approve-for-me` 时允许独立 AI approver 批准工作区外读取和 sandbox 网络开放；AI 失败、超时或返回无效结果时 MUST 默认拒绝。

#### Scenario: AI 批准外部读取

- **WHEN** 请求读取工作区外路径且策略允许 AI 裁决
- **THEN** AI approver 基于完整 action manifest 返回批准，系统签发一次性授权

#### Scenario: AI 批准 sandbox 网络

- **WHEN** outproc 工具请求网络能力并启用 `approve-for-me`
- **THEN** AI approver 可批准该次调用的网络范围

#### Scenario: AI 服务不可用

- **WHEN** AI approver 超时或返回无法解析的结果
- **THEN** 系统拒绝该请求且不启动 worker

### Requirement: 工作区外写入类操作必须用户明确批准

系统 MUST 对工作区外且非 `/tmp` 的 create、write、truncate、rename、delete、chmod、chown 和 symlink 操作要求用户明确批准；`approve-for-me` MUST NOT 单独批准这些操作。

#### Scenario: 删除工作区外文件

- **WHEN** action manifest 请求删除工作区外且非 `/tmp` 的文件
- **THEN** 系统强制进入用户确认流程，AI 自动批准不可替代

#### Scenario: 通过 symlink 修改工作区外目标

- **WHEN** 请求路径经过 symlink 或路径变化最终指向工作区外且非 `/tmp`
- **THEN** Broker 基于规范化真实路径要求用户明确批准

### Requirement: 审批 token 与 manifest 强绑定

系统 MUST 将审批 token 绑定到规范化 manifest hash、Session ID、Turn ID、ToolCall ID 和有效期；token MUST 一次性使用，参数或路径变化后 MUST 重新审批。

#### Scenario: 重放 approval token

- **WHEN** 同一 token 被第二次用于宿主执行
- **THEN** Broker 拒绝执行并记录重放事件

#### Scenario: 参数在批准后变化

- **WHEN** argv、cwd、写入目标、读取目标或网络范围与已批准 manifest 不一致
- **THEN** Broker 拒绝执行并要求重新审批

### Requirement: 宿主执行审计

系统 MUST 记录审批请求、用户或 AI 决策、manifest hash、argv、cwd、环境变量名称白名单、退出码和资源消耗。

#### Scenario: 宿主执行完成

- **WHEN** 宿主 worker 正常或异常退出
- **THEN** 审计记录包含决策来源、执行结果和关联 ToolCall ID

