package tool

import (
	"encoding/json"
	"fmt"
	"myagent/pkg/llm"
	"os"
	"strings"
)

type ListDirTool struct{}

type listDirArgs struct {
	Path string `json:"path"`
}

func (t ListDirTool) Name() string {
	return "list_directory"
}

func (t ListDirTool) Definition() *llm.Tool {
	return &llm.Tool{
		Type: llm.ToolTypeFunction,
		Function: llm.FunctionTool{
			Name:        t.Name(),
			Description: "list information about all files and folders within the directory",
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
	go t.execute(args, res)
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
