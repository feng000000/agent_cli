## 1. 工具契约与路由

- [ ] 1.1 定义 ToolDescriptor、执行分类、能力请求和版本字段；验收：无分类/无能力工具的路由拒绝测试。
- [ ] 1.2 建立 inproc/outproc registry 路由并挂接 Engine；验收：单元测试确认 inproc 不在 sandbox 启动、outproc 不获得 RuntimePort。
- [ ] 1.3 增加 import 架构检查，禁止 outproc/protocol/confine 引用 Agent/Session/Runtime；验收：违规引用时检查失败。

## 2. Worker 协议与生命周期

- [ ] 2.1 实现包含 request/session/turn/tool call ID、参数、能力、超时和资源限制的协议 DTO；验收：协议 round-trip 与缺失字段测试。
- [ ] 2.2 实现 supervisor 派生 one-shot worker、结果读取、错误分类与输出上限；验收：成功、非法请求、执行失败测试。
- [ ] 2.3 实现超时/取消时终止完整进程树；验收：启动子进程的 fixture 被全部终止。
- [ ] 2.4 完成工具列表和执行 handler，删除空响应路径；验收：端到端 executor 测试。

## 3. 沙箱策略

- [ ] 3.1 将默认 read 范围改为工作区、`/tmp`、中央运行时白名单，write 范围改为工作区与 `/tmp`；验收：路径矩阵测试。
- [ ] 3.2 移除不可信 outproc 的 BestEffort 降级，Landlock/seccomp 失败时 fail-closed；验收：mock 策略失败时工具未执行。
- [ ] 3.3 扩展 seccomp 防护，限制 ptrace、跨进程内存访问和跨进程组信号；验收：恶意 fixture 返回 EPERM。
- [ ] 3.4 增加进程组、rlimit/cgroup 和资源上限；验收：fork、内存、文件大小和超时 fixture 被限制。
- [ ] 3.5 将网络默认设为关闭，并只接受显式批准的能力；验收：无能力请求无法建立外网连接。

## 4. 迁移与验证

- [ ] 4.1 将现有高风险文件工具迁移到 outproc registry；验收：工具从 sandbox 读取工作区并拒绝工作区外路径。
- [ ] 4.2 增加恶意工具与逃逸回归测试；验收：`env GOCACHE=/tmp/agent_cli_go_cache go test ./internal/tool/... ./pkg/confine/...`。
- [ ] 4.3 更新 executor 运行说明和沙箱支持矩阵；验收：文档列明 Linux、Landlock ABI、libseccomp 和 fail-closed 行为。

