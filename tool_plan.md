# 文件系统工具扩展 + 工具执行架构安全加固

## Context

`internal/tool/filesystem.go` 目前只有 `ListDirTool`(可用)和 `ReadFileTool`(`execute` 为空桩,会让
`tool_call_handler.go:37` 的 `<-ch` **永久死锁**)。需要按既有结构补全更多文件系统工具(写入/修改/搜索/移动/
复制/删除),让 agent 具备真正的文件操作能力。

但这些是高危的破坏性操作,而当前工具执行架构存在三个安全缺口:
1. **无路径约束** —— 工具可读写删除磁盘任意位置。
2. **`ToolAskMode` 枚举已定义但从未被使用** —— 破坏性操作无审批。
3. **无 context/超时** —— 工具执行无法取消;空结果工具会死锁 handler。

目标:补全文件系统工具,并把执行架构改造为「目录约束 + 破坏性操作审批 + 超时取消 + 资源限制」。

用户决策:路径限制在工作目录内;三项架构改造全做;edit 同时支持字符串精确替换与正则替换。

---

## Part A — 工具执行架构改造

### A1. `Tool` 接口扩展(`internal/tool/interface.go`)
```go
type Tool interface {
    Name() string
    Desc() string
    Definition() *llm.Tool
    // 新签名:增加 ctx 用于取消/超时
    Execute(ctx context.Context, args string, res chan string)
    // 风险分级:true 表示会修改文件系统,需要审批
    Mutating() bool
}
```
- 同步更新 `ListDirTool` / `ReadFileTool` 及所有新工具的 `Execute` 签名。
- 只读工具(list/read/search)`Mutating()` 返回 `false`;写/改/移动/复制/删除返回 `true`。

### A2. 路径约束(`internal/tool/filesystem.go` 新增 `baseFsTool`)
所有文件系统工具内嵌一个基类,持有 workspace 根目录并提供安全解析:
```go
type baseFsTool struct{ root string } // 绝对路径

// resolve: 相对路径基于 root;Clean 后用 filepath.Rel/EvalSymlinks
// 校验解析结果(含父目录符号链接)仍位于 root 内,否则返回错误。
func (b baseFsTool) resolve(path string) (string, error)
```
- 拒绝 `../` 逃逸与符号链接逃逸。写/删除前对父目录做 `EvalSymlinks` 校验。
- 复用于全部工具的入口处。

### A3. 注入根目录(`internal/agent/core.go`)
- `registerTools()` → `registerTools(cfg config.ProjectConfig)`,根目录取 `cfg.Workspace.WorkspaceDir`
  (转绝对路径,`InitAgentState` 中已有 `a.Config`)。
- 每个工具用构造函数注入,如 `tool.NewWriteFileTool(root)`,注册全部新工具 + 修复后的 `ReadFileTool`。

### A4. handler 改造(`internal/handler/tool_call_handler.go`)
当前「先全部 `Execute` 再顺序阻塞读 channel」需重写为逐个处理:
1. **审批门控**:按 `state.AgentConfig.ToolAskMode` + `tool.Mutating()`:
   - `ToolAskModeNone` → 拒绝全部工具调用,回填工具结果 `"tool call rejected by policy"`。
   - `ToolAskModeAuto` → 全部自动批准。
   - `ToolAskModeAlways` → 仅对 `Mutating()==true` 的工具发起确认:经 `OutputChan` 发
     `AgentRespTypeMiddleMsg`(描述将执行的操作 + 参数),阻塞读 `state.InputChan` 取用户决定;
     非破坏性工具仍自动执行。被拒则回填拒绝结果,不执行。
2. **超时/取消**:为每次 `Execute` 派生 `context.WithTimeout(state.Ctx, 30s)`,用
   `select { case res := <-ch; case <-ctx.Done(): }` 读取,超时/取消回填错误结果 —— 彻底消除空结果死锁。
3. 工具结果统一通过 `llm.ToolResultMessage(tc.ID, res)` 回填(保持现有逻辑)。

### A5. 资源限制(`filesystem.go` 包级常量)
- `maxReadSize` / `maxWriteSize`(如 1 MiB):超限读取截断并标注,写入超限报错。
- `maxSearchResults`(如 200)、`maxWalkFiles`(如 5000):搜索结果/遍历上限。
- 遍历跳过 `.git`、`node_modules` 等;按 NUL 字节检测跳过二进制文件。

---

## Part B — 文件系统工具(全部写入 `internal/tool/filesystem.go`)

沿用既有模式:`struct{baseFsTool}` + args 结构体 + `Name/Desc/Definition/Mutating/Execute(goroutine+panic 恢复)/execute`。

| 工具 | Name | Mutating | 关键参数 | 说明 |
|---|---|---|---|---|
| 修复 ReadFileTool | `read_file` | false | `path`, 可选 `offset`/`limit` | 实现空 `execute`:按 `maxReadSize` 读取,可选行范围 |
| 写入 | `write_file` | true | `path`, `content` | 创建/覆盖;自动建父目录;`maxWriteSize` |
| 修改 | `edit_file` | true | `path`, `old`, `new`, `mode`("literal"\|"regex"), `replace_all` | literal 要求 `old` 唯一;regex 用 `regexp` 替换 |
| 文件名搜索 | `search_filename` | false | `path`, `pattern`(regex), 可选 `max` | `filepath.WalkDir` + `regexp` 匹配文件名;受 walk/result 上限约束 |
| 内容搜索 | `search_content` | false | `path`, `pattern`(regex), 可选 `max` | 遍历逐行匹配,返回 `file:line: 内容`;跳过二进制;受上限约束 |
| 移动/重命名 | `move_file` | true | `src`, `dst` | `os.Rename`;src/dst 均经 `resolve` 约束 |
| 复制 | `copy_file` | true | `src`, `dst`, 可选 `recursive` | 文件复制(可选目录递归);保留权限 |
| 删除 | `delete_file` | true | `path`, 可选 `recursive` | `os.Remove`/`RemoveAll`(目录需显式 `recursive`) |

实现要点:
- 每个 `execute` **第一步**调用 `b.resolve(args.Path)`,失败立即回 channel 报错。
- 保证每条路径 **恰好向 channel 发送一次**结果(避免死锁;A4 超时是兜底)。
- `edit_file` 的 `mode` 默认 `literal`;`literal` 下 `old` 命中 0 次或(非 `replace_all`)>1 次均报错。

---

## 关键改动文件
- `internal/tool/interface.go` —— 接口加 `ctx` 与 `Mutating()`。
- `internal/tool/filesystem.go` —— `baseFsTool` + 修复 ReadFile + 8 个工具 + 资源常量。
- `internal/agent/core.go` —— `registerTools(cfg)` 注入 root 并注册全部工具。
- `internal/handler/tool_call_handler.go` —— 审批门控 + 超时/取消读取重写。

## 验证
1. `go build ./...` 与 `go vet ./...` 通过。
2. 为 `baseFsTool.resolve` 写单测:`../` 逃逸、绝对路径越界、符号链接逃逸均被拒;正常相对路径通过。
3. `edit_file` 单测:literal 唯一命中/0 命中/多命中、regex 替换、`replace_all`。
4. 端到端:在 `ToolAskModeAlways` 下触发一次 `write_file`,确认走 `OutputChan` 审批、`InputChan` 拒绝时不落盘;
   `ToolAskModeNone` 下全部被拒;`ToolAskModeAuto` 下读/写/搜索/移动/复制/删除按序生效。
5. 回归:`read_file` 不再死锁(此前空桩问题)。
