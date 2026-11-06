## 1. 迁移工具包引用

- [ ] 1.1 在 `internal/tool/inproc/interface.go` 和 `toolimpl` 中统一 package/import 为 `agentcli/internal/tool/inproc`；验收：`rg -n 'agentcli/internal/tool(?!/(inproc|executor))'` 无旧工具引用。
- [ ] 1.2 迁移 `internal/session/toolstate` 与 `internal/handler` 的旧 import 和类型引用；验收：`go test ./internal/session/... ./internal/handler/...`。
- [ ] 1.3 迁移 `internal/tool/executor`、`cmd/executor` 和 `cmd/test` 的路径与类型名；验收：`go test ./internal/tool/executor ./cmd/...`。

## 2. 修复构建占位

- [ ] 2.1 将 `cmd/executor` 的未完成 `runTool` 改为明确返回“执行器待实现”错误的最小适配，不构造伪造 Runtime/Meta；验收：该包可编译且执行时返回可识别错误。
- [ ] 2.2 检查并安装 Linux `libseccomp` pkg-config 依赖；验收：`pkg-config --modversion libseccomp` 成功。
- [ ] 2.3 运行全仓测试并记录任何与本迁移无关的失败；验收：`env GOCACHE=/tmp/agent_cli_go_cache go test ./...`。

## 3. 防止回归

- [ ] 3.1 增加 CI/脚本检查旧 `agentcli/internal/tool` 路径不回流；验收：架构检查命令在旧 import 存在时失败。
- [ ] 3.2 更新开发说明中的包路径和 libseccomp 构建前置条件；验收：README 或开发文档可通过命令复现构建。

