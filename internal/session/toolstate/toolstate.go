package toolstate

import "os"
import "encoding/json"

import "agentcli/internal/tool"
import "agentcli/internal/tool/toolimpl"


type ToolState struct {

	ToolMap  map[string]tool.InternalTool `json:"-"`
	// 已执行的 tool 结果
	ToolResults []tool.ToolResult `json:"tool_results"`
}


func defaultToolMap() map[string]tool.InternalTool {
	listDir := &toolimpl.ListDirTool{}
	readFile := &toolimpl.ReadFileTool{}

	return map[string]tool.InternalTool{
		listDir.Name():  listDir,
		readFile.Name(): readFile,
	}
}

// NewState 新建 ToolState
func NewToolState() *ToolState {
	toolState := &ToolState{
		ToolMap: defaultToolMap(),
	}
	return toolState
}

// LoadState 从给定 path 中加载 ToolState
func LoadToolState(path string) (*ToolState, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	toolState := &ToolState{}
	if err := decoder.Decode(toolState); err != nil {
		return nil, err
	}

	toolState.ToolMap = defaultToolMap()
	return toolState, nil
}
