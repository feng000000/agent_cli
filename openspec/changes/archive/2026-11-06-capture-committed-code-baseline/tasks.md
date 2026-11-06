## 1. 建立可复现基线

- [x] 1.1 确认基线提交为 `29f9faf`，并使用 `git archive HEAD` 解包到 `/tmp/agent_cli_head_29f9faf`；验收：`git rev-parse HEAD` 输出 `29f9faf`。
- [x] 1.2 通过 `git status --short` 记录排除范围：7 个已修改跟踪文件及未跟踪的 `.codex/`、`openspec/`；验收：proposal 的假设与影响范围准确列出该边界。
- [x] 1.3 只从临时快照读取已提交源码，核对入口、Agent、会话、命令、工具、LLM、MCP 与沙箱模块；验收：proposal 的 Current Baseline 覆盖上述模块且不引用未提交实现。

## 2. 固化实现状态与风险

- [x] 2.1 在临时快照执行 `env GOCACHE=/tmp/agent_cli_go_cache go test ./...`；验收：记录旧工具导入路径和 `libseccomp` pkg-config 两类失败，不宣称 `HEAD` 可完整构建。
- [x] 2.2 区分已实现骨架、占位实现和未来路线图，包括工具迁移、技能加载、Web Search、MCP Client/Server、ACP 与 TUI；验收：proposal 与 design 对边界描述一致。
- [x] 2.3 确认本 change 无运行行为变化，使用纯文档归档策略且不修改主 specs；验收：`.openspec.yaml` 设置 `skip_specs: true`，design 记录取舍。

## 3. 校验与归档

- [x] 3.1 执行 `openspec validate capture-committed-code-baseline --strict --type change --json`；验收：确认纯文档 change 因无 delta specs 返回预期校验错误，并决定使用 `--skip-specs` 归档而不伪造 capability。
- [x] 3.2 执行 `openspec archive capture-committed-code-baseline --yes --skip-specs --json`；验收：change 进入 `openspec/changes/archive/`，`openspec list --json` 不再显示活动 change。
