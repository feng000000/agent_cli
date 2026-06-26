package runtime

import (
	"context"
	"myagent/internal/tool"
	"myagent/pkg/llm"
)

type AgentMode string

const (
	AgentModePlan     AgentMode = "plan"
	AgentModeAutoEdit AgentMode = "auto-edit"
)

type ToolAskMode string

const (
	ToolAskModeAuto   ToolAskMode = "auto"   // 自动批准，不询问
	ToolAskModeAlways ToolAskMode = "always" // 每次都询问
	ToolAskModeNone   ToolAskMode = "none"   // 拒绝所有工具调用
)

type AgentConfig struct {
	AgentMode   AgentMode
	ToolAskMode ToolAskMode
}

// TODO: Model Usage
type AgentState struct {
	Ctx context.Context

	// InputChan 可以直接追加信息
	InputChan chan string
	// OutputChan 可以直接输出临时信息
	OutputChan chan AgentResponse

	AgentConfig AgentConfig
	LLMClient  llm.LLMClient

	UserQuery string

	MessageParams []llm.Message
	ToolMap       map[string]tool.Tool

	Response *llm.ChatResponse
}

func (lc *AgentState) ToolList() []llm.Tool {
	toolList := []llm.Tool{}
	for _, tool := range lc.ToolMap {
		toolList = append(toolList, *tool.Definition())
	}
	return toolList
}

type ResponseType string

const (
	AgentRespTypeCmd       = "command"
	AgentRespTypeLLM       = "llm_response"
	AgentRespTypeError     = "error"
	AgentRespTypeMiddleMsg = "middle_message"
)

type AgentResponse struct {
	RespType ResponseType

	CmdResult     string
	LLMResponse   *llm.ChatResponse
	Err           error
	MiddleMessage string
}



type ClientState struct {
	CancelFunc func()

}
