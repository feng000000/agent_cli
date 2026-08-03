package main

import "fmt"
import "os"
import "flag"
import "strings"
import "encoding/json"

import "agentcli/internal/tool"
import "agentcli/internal/tool/toolimpl"
import "agentcli/pkg/confine"

type stringListParam []string

func (slp *stringListParam) String() string {
	return strings.Join(*slp, ",")
}

func (slp *stringListParam) Set(value string) error {
	*slp = append(*slp, value)
	return nil
}

type RetCodeEnum int

const (
	RetCodeSuccess        RetCodeEnum = iota
	RetCodeInvalidContent RetCodeEnum = iota
	RetCodeInvalidParam   RetCodeEnum = iota
	RetCodeExecuteFailed  RetCodeEnum = iota
)


// Output 打印执行结果并退出
func Output(code RetCodeEnum, content any) {
	err := json.NewEncoder(os.Stdout).Encode(
		struct {
			Code    RetCodeEnum `json:"code"`
			Content any         `json:"content"`
		}{
			Code:    code,
			Content: content,
		},
	)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(
			struct {
				Code    RetCodeEnum `json:"code"`
				Content any         `json:"content"`
			}{
				Code:    RetCodeInvalidContent,
				Content: fmt.Sprintf("Invalid content: %v", content),
			},
		)
	}
	os.Exit(0)
}

func main() {
	tools := []tool.Tool{
		&toolimpl.ListDirTool{},
		&toolimpl.ReadFileTool{},
		&toolimpl.LoadSkillTool{},
		// TODO: register more tools
	}

	listTool := flag.Bool("list", false, "list the registry tool")
	toolName := flag.String("name", "", "the tool name")
	workDir := flag.String("workdir", "", "the work dir")
	allowNet := flag.Bool("allow-net", true, "if allow network")

	var extraDirs *stringListParam
	var readableDirs *stringListParam
	flag.Var(extraDirs, "extra-dir", "extra writable directory; may be called repeatedly")
	flag.Var(readableDirs, "readable-dir", "the readable dir; may be called repeatedly")

	flag.Parse()

	if *listTool {
		type toolInfo struct {
			Name string `json:"name"`
			Desc string `json:"desc"`
		}

		infoList := []toolInfo{}
		for _, tool := range tools {
			infoList = append(
				infoList,
				toolInfo{Name: tool.Name(), Desc: tool.Desc()},
			)
		}
		Output(RetCodeInvalidParam, infoList)
		return
	}

	if toolName == nil || *toolName == "" || workDir == nil || *workDir == "" {
		Output(RetCodeInvalidParam, "tool name or workdir cannot be empty")
	}

	var targetTool tool.Tool
	for _, tool := range tools {
		if *toolName == tool.Name() {
			targetTool = tool
			break
		}
	}
	if targetTool == nil {
		Output(
			RetCodeInvalidParam,
			fmt.Sprintf("tool %v not exists", toolName),
		)
		return
	}

	policy := confine.Policy{
		WritableDirs: append([]string{*workDir}, *extraDirs...),
		ReadableDirs: *readableDirs,
		AllowNetwork: *allowNet,
	}

	if err := confine.Apply(policy); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	res, err := runTool(targetTool)
	if err != nil {
		Output(
			RetCodeExecuteFailed,
			fmt.Sprintf("exec tool failed: %v", err.Error()),
		)
		return
	}

	Output(RetCodeSuccess, res)
}


// TODO: Executor cmd 一次性执行, 并附加限制, 外层以 mcp server 暴露

func runTool(t tool.Tool) (string, error) {
	// 可以读取大部分文件。
	// 只能写入 WritableDirs。
	// 默认无法创建外联 socket。

	res := tool.ExecTool(s.Meta, s.Runtime, runtimeMu, tc.ID, t, tc.Function.Arguments)
}
