## 1. Artifact Store

- [ ] 1.1 实现受限 quarantine 与 artifact metadata，包括来源、重定向、content-type、大小、digest 和权限；验收：写入、读取、权限和损坏元数据测试。
- [ ] 1.2 实现原子 artifact 写入、执行前 digest 校验和 symlink/rename 防护；验收：替换与 TOCTOU 测试。
- [ ] 1.3 实现保留期、大小上限和失败清理；验收：超限下载、取消和执行结束后的清理测试。

## 2. Fetch 与网络出口

- [ ] 2.1 实现 fetch worker，只访问批准网络范围并写 quarantine，不读工作区；验收：隔离文件访问测试。
- [ ] 2.2 实现 egress proxy 的域名/地址 allowlist 和重定向校验；验收：allowlist、越界目标、越界重定向测试。
- [ ] 2.3 禁止 curl/wget 到 shell/解释器的管道及等价命令；验收：策略拒绝测试。
- [ ] 2.4 记录 fetch 审计事件；验收：来源、目标、重定向和响应摘要可查询。

## 3. Run Artifact

- [ ] 3.1 实现 run-artifact 只接受 artifact ID + digest，并在无网络 sandbox 中执行；验收：未固定 digest、无网络和执行结果测试。
- [ ] 3.2 接入审批策略：未知来源强制用户确认，可信固定 digest 可由 `approve-for-me` 批准；验收：未知、可信、签名失效与超时场景测试。
- [ ] 3.3 将大输出保存为受限 artifact 并限制返回上下文大小；验收：大输出 fixture 不导致无界上下文增长。

## 4. 回归与文档

- [ ] 4.1 增加端到端 fetch→digest→approval→run 流程测试；验收：每个阶段独立审计且不可跳过。
- [ ] 4.2 运行安全回归：`env GOCACHE=/tmp/agent_cli_go_cache go test ./internal/tool/outproc/... ./internal/approval/...`。
- [ ] 4.3 更新远程 artifact 使用说明、可信 digest 策略和失败处理文档；验收：文档明确禁止直接下载执行。

