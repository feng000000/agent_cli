package runtime

import "myagent/internal/tools"
import "myagent/pkg/llm"

type LoopContext struct {
	// InputChan 可以直接追加信息
	InputChan chan string
	// OutputChan 可以直接输出临时信息
	OutputChan chan string
	LLMClient  llm.LLMClient

	Query string

	MessageParams []llm.Message
	ToolMap		map[string]tools.Tool

	Response *llm.ChatResponse
}

func (lc *LoopContext) ToolList() []llm.Tool {
	toolList := []llm.Tool{}
	for _, tool := range lc.ToolMap {
		toolList = append(toolList, *tool.Definition())
	}
	return toolList
}

// type LoopResponse struct {
// 	Answer   string
// 	ToolCall []llm.ToolCall
// }
