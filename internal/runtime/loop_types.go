package runtime

import "myagent/internal/tool"
import "myagent/pkg/llm"

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
	AgentMode AgentMode
	ToolAskMode ToolAskMode
}

// TODO: Usage
type LoopContext struct {
	AgentConfig AgentConfig

	// InputChan 可以直接追加信息
	InputChan chan string
	// OutputChan 可以直接输出临时信息
	OutputChan chan AgentResponse
	LLMClient  llm.LLMClient

	UserQuery string

	MessageParams []llm.Message
	ToolMap       map[string]tool.Tool

	Response *llm.ChatResponse
}

func (lc *LoopContext) ToolList() []llm.Tool {
	toolList := []llm.Tool{}
	for _, tool := range lc.ToolMap {
		toolList = append(toolList, *tool.Definition())
	}
	return toolList
}


type ResponseType string
const (
	AgentRespTypeCMD = "command"
	AgentRespTypeLLM = "llm_response"
	AgentRespTypeError = "error"
	AgentRespTypeMiddleMsg = "middle_message"
)

type AgentResponse struct {
	RespType ResponseType

	CMDResult string
	LLMResponse *llm.ChatResponse
	Err error
	MiddleMessage string
}
