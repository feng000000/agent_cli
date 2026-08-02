package toolimpl

import "encoding/json"
import "fmt"
import "os"
import "strings"
import "time"

import "agentcli/internal/tool"
import "agentcli/pkg/llm"

type ListDirTool struct{}

type listDirArgs struct {
	Path string `json:"path"`
}

func (t *ListDirTool) Name() string {
	return "list_directory"
}

func (t *ListDirTool) Desc() string {
	return "list information about all files and folders within the directory"
}


func (t *ListDirTool) Timeout() time.Duration {
	return time.Second * 3
}

func (t *ListDirTool) Definition() *llm.Tool {
	return &llm.Tool{
		Type: llm.ToolTypeFunction,
		Function: llm.FunctionTool{
			Name:        t.Name(),
			Description: t.Desc(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Folder directory to view",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

func (t *ListDirTool) ExecuteImpl(
	tc *tool.ToolContext,
	arg string,
) (string, error) {
	var args listDirArgs
	if err := json.Unmarshal([]byte(arg), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("read directory failed: %v", err)

	}

	// list dir
	names := make([]string, 0, len(entries))
	limit := 200
	for _, entry := range entries {
		name := entry.Name()
		info, err := entry.Info()
		if err != nil {
			return "", fmt.Errorf("read entry info failed: %v", err)

		}

		filemode := info.Mode()
		filesize := info.Size()

		if entry.IsDir() {
			name += "/"
		}
		// -rw-rw-r--	330	file.go
		names = append(names, fmt.Sprintf("%v\t%v\t%v", filemode, filesize, name))
		if len(names) >= limit {
			names = append(names, "[truncated]: too many files/folders")
			break
		}
	}

	if len(names) == 0 {
		return "", fmt.Errorf("directory is empty")

	}

	return strings.Join(names, "\n"), nil
}

type ReadFileTool struct{}

type ReadFileArgs struct {
	Path string `json:"path"`
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Timeout() time.Duration {
	return time.Second * 5
}

func (t *ReadFileTool) Desc() string {
	return "read the given path file, return the file content (string)"
}

func (t *ReadFileTool) Definition() *llm.Tool {
	return &llm.Tool{
		Type: llm.ToolTypeFunction,
		Function: llm.FunctionTool{
			Name:        t.Name(),
			Description: "read the given path file, return the file content (string)",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Folder directory to view",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

// TODO: 封装解析参数操作
// TODO: 空返回在外部统一处理, 空返回导致 deepseek api 报错
func (t *ReadFileTool) ExecuteImpl(
	tc *tool.ToolContext,
	arg string,
) (string, error) {
	var args ReadFileArgs
	if err := json.Unmarshal([]byte(arg), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	res := string(data)
	if res == "" {
		res = "<empty file>"
	}
	return res, nil
}


type ReadMemoryTool struct{}

func (t *ReadMemoryTool) Name() string {
	return "read_memory"
}

func (t *ReadMemoryTool) Desc() string {
	return "Read the memory from disk to update context"
}

func (t *ReadMemoryTool) Timeout() time.Duration {
	return time.Second * 5
}

func (t *ReadMemoryTool) Definition() *llm.Tool {
	return &llm.Tool{
		Type: llm.ToolTypeFunction,
		Function: llm.FunctionTool{
			Name:        t.Name(),
			Description: t.Desc(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{},
			},
		},
	}
}

func (t *ReadMemoryTool) ExecuteImpl(
	tc *tool.ToolContext,
	arg string,
) string {
	memory_bytes, err := os.ReadFile(tc.Meta.Persistence.MemoryPath)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err.Error())
	}

	return string(memory_bytes)
}
