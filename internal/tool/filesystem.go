package tool

import "encoding/json"
import "fmt"
import "os"
import "strings"

import "myagent/pkg/llm"

type ListDirTool struct{}

type listDirArgs struct {
	Path string `json:"path"`
}

func (t ListDirTool) Name() string {
	return "list_directory"
}

func (t ListDirTool) Desc() string {
	return "list information about all files and folders within the directory"
}

func (t ListDirTool) Definition() *llm.Tool {
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

func (t ListDirTool) Execute(args string, res chan string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				res <- fmt.Sprintf("exec tool panic: %v", r)
			}
		}()
		t.execute(args, res)
	}()
}

func (t ListDirTool) execute(arg string, res chan string) {
	var args listDirArgs
	if err := json.Unmarshal([]byte(arg), &args); err != nil {
		res <- fmt.Sprintf("invalid arguments: %v", err)
		return
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		res <- "path is required"
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		res <- fmt.Sprintf("read directory failed: %v", err)
		return
	}

	names := make([]string, 0, len(entries))
	limit := 200
	for _, entry := range entries {
		name := entry.Name()
		info, err := entry.Info()
		if err != nil {
			res <- fmt.Sprintf("read entry info failed: %v", err)
			return
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
		res <- "directory is empty"
		return
	}

	res <- strings.Join(names, "\n")
}

type ReadFileTool struct{}

type ReadFileArgs struct {
	Path string `json:"path"`
}

func (t ReadFileTool) Name() string {
	return "read_file"
}

func (t ReadFileTool) Desc() string {
	return "read the given path file, return the file content (string)"
}

func (t ReadFileTool) Definition() *llm.Tool {
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

func (t ReadFileTool) Execute(args string, res chan string) {
	go t.execute(args, res)
}

func (t ReadFileTool) execute(arg string, res chan string) {
	// TODO: implement
}
