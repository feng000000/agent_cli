## Why

让模型从网络获取脚本并直接执行，会把可变远程内容、网络能力和脚本执行合并成一次难以审批的操作。`curl | bash` 无法向用户展示实际执行内容，也无法把批准绑定到固定 digest，存在供应链、提示注入、数据外传和 TOCTOU 风险。

## What Changes

- 禁止 `curl | bash`、`wget -O- | sh` 及等价下载后直接执行的管道。
- 新增独立的 fetch 能力：通过受控网络读取远程 artifact，只写 quarantine，不读取工作区，不继承敏感环境。
- fetch 结果记录来源 URL、重定向、内容类型、大小上限和 SHA-256 digest。
- 新增 run-artifact 能力：只执行本地固定 artifact，默认无网络，执行前重新校验 digest，并使用只读文件描述符或等价机制防止替换。
- 网络访问通过受控 egress proxy 或等价目标范围控制实现，而不是仅允许任意 IPv4/IPv6 socket。
- `approve-for-me` 可批准网络和外部读取，但未知来源的“下载后执行”默认要求用户确认；可信签名、固定 digest 和明确 allowlist 可作为自动批准条件。
- 设置 artifact 大小、超时、资源、输出和保留期限；清理 quarantine 与长期 tool result。
- 增加针对管道执行、重定向变化、digest 不匹配、symlink 替换、超限下载和越权网络的测试。

## Capabilities

### New Capabilities

- `remote-artifact-execution`: 定义远程内容下载、隔离存储、digest 审批、无网络执行、网络出口控制和清理行为。

### Modified Capabilities

无。仓库当前没有已归档的远程执行 spec。

## Impact

- 受影响模块：outproc fetch/run-artifact 工具、网络 proxy、approval manifest、artifact store、executor 协议和审计。
- 新增 quarantine 目录与 digest/来源元数据；不涉及数据库迁移。
- 安全影响：远程代码执行被拆分为两次可审计授权；脚本执行默认无网络，网络访问目标可限制。
- 依赖 `implement-outproc-sandbox-execution` 与 `implement-host-approval-broker`。

## Assumptions

- 宿主具备可靠的文件 digest、原子写入和可取消子进程能力。
- AI 审批可以读取脚本内容，但其结果仍是策略输入，不是安全证明。
- 可信来源和签名策略可由后续配置扩展，第一版优先采用手动确认。

## Non-Goals

- 不保证静态分析能证明任意脚本安全。
- 不允许 shell 工具绕过 fetch/run-artifact 规则直接下载并执行内容。
- 不实现操作系统包管理器或软件签名体系的通用替代方案。
