## Why

当前仓库 `HEAD`（`29f9faf`，`WIP: distinguish inproc tool and outproc tool`）包含大量仍在演进中的 Agent CLI 骨架，但缺少一份只基于已提交代码的现状基线。工作区同时存在未提交修复和未跟踪的 OpenSpec 配置，若直接依据工作区总结，会把尚未提交、尚未验证的实现状态混入事实基线。

## Current Baseline

- 入口：`cmd/agent` 加载 `config.jsonc`、初始化日志并启动简单命令行 UI；`cmd/executor`、`cmd/test` 是工具执行器和 LLM 调用的开发入口。
- Agent 循环：`internal/agent` 将用户输入分流到普通查询、Agent 命令或工具调用，并持续请求 LLM 直到得到最终文本；支持等待事件、取消信号和上下文压缩。
- 会话与上下文：`internal/session` 保存元数据、运行态、消息、工具结果和用量；提供磁盘加载/保存骨架、消息队列，以及截断、LLM 总结和混合三种压缩策略。
- LLM 接入：`pkg/llm` 定义通用 Chat Completions 数据模型，`DeepSeekClient` 支持环境变量/配置注入 API Key、默认 Base URL/模型、工具定义和结构化 API 错误。
- 工具层：已提交目录只保留 `internal/tool/inproc` 及 `list_directory`、`read_file`、`read_memory`、`load_skill`、`release_skill`、`web-search` 骨架；内部工具包含 JSON 参数定义、执行超时、panic 恢复以及长结果落盘逻辑。
- 命令系统：`internal/command` 注册 Agent 命令 `ciallo`、`compress`、`stop`；客户端命令注册表为空，`stop` 和技能工具仍为占位实现。
- 安全边界：`pkg/confine` 组合 Landlock 文件系统限制与 seccomp 网络限制；设计目标是全盘只读、仅白名单目录可写、默认禁止 IPv4/IPv6 socket，并继承到子进程。
- 服务端扩展：`internal/server/mcp` 包含一个 Streamable HTTP MCP server 原型；`pkg/mcp` 与 ACP Agent server 仍只有占位说明。
- 可构建性：在独立解包的 `HEAD` 上执行 `env GOCACHE=/tmp/agent_cli_go_cache go test ./...` 失败。已提交代码将工具包移动到 `internal/tool/inproc`，但 `internal/session/toolstate` 等多处仍引用不存在的 `agentcli/internal/tool` 与 `agentcli/internal/tool/toolimpl`；执行当前环境还缺少 `libseccomp` 的 pkg-config 元数据。因此本基线只描述代码结构和实现状态，不宣称 `HEAD` 已可完整构建或运行。

## What Changes

- 新增一份仅基于已提交代码的 OpenSpec 基线文档，记录模块职责、主要数据流、已实现范围、占位能力与已知构建缺口。
- 明确分析边界：使用 `HEAD` `29f9faf` 的 Git 对象；排除 7 个已修改的已跟踪文件，以及未跟踪的 `.codex/` 和 `openspec/`。
- 将后续修复拆分为独立 change 候选，不在本 change 中修改 Go 代码、依赖、配置或接口。
- 完成文档校验后归档本 change，不更新主 specs，避免把尚不可构建的 WIP 状态声明为正式系统能力。

## Capabilities

### New Capabilities

无。本 change 仅归档现状分析和实施建议，不引入或改变系统行为。

### Modified Capabilities

无。

## Impact

- 受影响交付物：`openspec/changes/capture-committed-code-baseline/` 下的 proposal、design、tasks；归档后位于 `openspec/changes/archive/`。
- 分析覆盖范围：`cmd/`、`internal/agent/`、`internal/session/`、`internal/handler/`、`internal/command/`、`internal/tool/`、`internal/server/`、`pkg/confine/`、`pkg/llm/`、`pkg/logger/` 与 `pkg/mcp/`。
- 运行行为、API、配置格式、依赖、数据存储和安全策略均不变。
- 无数据迁移、兼容性切换或回滚操作。

## Assumptions

- “当前代码”指 `HEAD` `29f9faf` 的已提交跟踪文件，不包含任何工作区改动或未跟踪文件。
- 本次要求是固化代码现状与分析结论，而不是依据现状实现功能修复。
- 已提交的 WIP 代码先按“结构存在、完整度不一”记录；构建失败与占位实现不升级为正式 capability。

## Non-Goals

- 不修复工具包导入路径、`cmd/executor` 调用链或 `libseccomp` 环境问题。
- 不为现有 WIP 功能补测试，也不改变 Agent、会话、工具或 MCP 的运行语义。
- 不把未提交代码中的修复纳入基线。
