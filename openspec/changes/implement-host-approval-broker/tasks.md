## 1. Manifest 与确定性策略

- [ ] 1.1 定义 action manifest、规范化规则、真实路径解析和 hash；验收：路径、symlink、argv、环境与网络范围规范化测试。
- [ ] 1.2 实现硬拒绝、用户强制确认、AI eligible 和默认拒绝策略顺序；验收：策略表驱动测试。
- [ ] 1.3 将工作区外且非 `/tmp` 的写入类 action 标记为用户强制确认；验收：create/write/delete/rename/chmod/symlink 场景测试。

## 2. 审批 token 与宿主 worker

- [ ] 2.1 实现一次性 approval token、TTL、Session/Turn/ToolCall 绑定与重放拒绝；验收：过期、重放和跨调用 token 测试。
- [ ] 2.2 实现 host one-shot worker：非 root 检查、`NO_NEW_PRIVS`、最小环境、进程组和资源限制；验收：root、sudo 请求和敏感环境继承测试。
- [ ] 2.3 在执行前重新校验 manifest、真实路径与 token；验收：参数替换、symlink 替换和执行前修改测试。
- [ ] 2.4 实现审批审计追加写入；验收：成功、拒绝、超时、失败均有可关联记录。

## 3. 手动与 AI 审批

- [ ] 3.1 实现 CLI 手动审批展示和批准/拒绝流程；验收：宿主请求在未确认时不启动 worker。
- [ ] 3.2 实现独立最小上下文的 AI approver，支持工作区外读取与 sandbox 网络批准；验收：批准、拒绝、超时、无效 JSON 测试。
- [ ] 3.3 实现 `approve-for-me` 开关与默认关闭；验收：未启用时不会调用 AI approver。
- [ ] 3.4 验证 AI 不能批准工作区外写入类操作；验收：所有写操作场景均强制用户确认。

## 4. 回归与文档

- [ ] 4.1 增加审批状态机与并发 ToolCall 测试；验收：待审批、拒绝、批准、过期和取消状态转换稳定。
- [ ] 4.2 运行安全回归：`env GOCACHE=/tmp/agent_cli_go_cache go test ./internal/approval/... ./cmd/...`。
- [ ] 4.3 更新审批策略、环境白名单和审计格式文档；验收：策略变更可通过配置和测试复现。

