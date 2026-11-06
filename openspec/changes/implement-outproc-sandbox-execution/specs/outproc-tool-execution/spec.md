## ADDED Requirements

### Requirement: 工具按 inproc 与 outproc 分类路由

系统 MUST 为每个工具提供明确执行分类和能力描述，并仅在 outproc 路径中启动独立 worker；outproc worker MUST NOT 导入或访问 Agent、Session 或 Runtime。

#### Scenario: 调用 outproc 工具

- **WHEN** LLM 调用标记为 outproc 的工具
- **THEN** Engine 通过 outproc 协议启动独立 sandbox worker 并等待结构化结果
- **AND** worker 无法通过内存对象修改 Session/runtime

#### Scenario: 未知分类

- **WHEN** 工具描述缺少执行分类或能力声明
- **THEN** 系统拒绝调用并返回配置错误

### Requirement: Outproc 默认文件策略

系统 MUST 默认仅允许 outproc worker 读取工作区、`/tmp` 和中央配置的只读运行时路径，并仅允许写入工作区和 `/tmp`；其他读取和写入 MUST 被拒绝，除非该请求获得显式能力授权。

#### Scenario: 读取工作区内文件

- **WHEN** outproc 工具请求读取工作区内的文件
- **THEN** worker 在 sandbox 中读取并返回内容或 artifact 引用

#### Scenario: 读取工作区和运行时路径外文件

- **WHEN** outproc 工具请求读取未授权的外部路径
- **THEN** sandbox 拒绝访问，且结果不包含文件内容

#### Scenario: 写入 `/tmp` 外非工作区路径

- **WHEN** outproc 工具尝试写入工作区和 `/tmp` 之外的路径
- **THEN** sandbox 拒绝写入并报告权限错误

### Requirement: Outproc 默认禁止网络

系统 MUST 默认禁止 outproc worker 访问外网；只有请求中包含已批准网络能力时才可启用受控出口。

#### Scenario: 无网络能力

- **WHEN** outproc 请求未声明网络能力
- **THEN** worker 无法建立 IPv4 或 IPv6 外网连接

#### Scenario: 网络能力被拒绝

- **WHEN** 请求声明网络能力但审批结果为拒绝
- **THEN** worker 不启动网络访问且返回拒绝状态

### Requirement: 沙箱必须 fail-closed

系统 MUST 在 Landlock、seccomp 或所需进程隔离无法强制时拒绝运行 outproc 工具，不得在降级 enforcement 下执行。

#### Scenario: 内核不支持 Landlock

- **WHEN** worker 无法启用要求的 Landlock 策略
- **THEN** 系统终止 worker 并返回 sandbox unavailable

#### Scenario: seccomp 策略加载失败

- **WHEN** seccomp 过滤器无法加载到 worker 进程
- **THEN** 工具不得执行，且错误被记录

### Requirement: Outproc 调用可控终止

系统 MUST 对每次 outproc 调用施加超时和资源限制，并在超时或取消时终止整个 worker 进程树。

#### Scenario: 工具超时

- **WHEN** outproc 工具超过声明超时
- **THEN** supervisor 终止 worker 及其子进程并返回 timeout 结果

#### Scenario: 输出超过限制

- **WHEN** 工具输出超过配置上限
- **THEN** 系统截断或保存为受限 artifact，并保持上下文安全边界

