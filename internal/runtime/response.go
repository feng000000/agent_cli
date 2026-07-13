package runtime

import "myagent/pkg/llm"

type ResponseTypeEnum string

const (
	AgentRespTypeCmd       = "command"
	AgentRespTypeLLM       = "llm_response"
	AgentRespTypeError     = "error"
	AgentRespTypeMiddleMsg = "middle_message"
)

type AgentResponse struct {
	RespType ResponseTypeEnum

	CmdResult     string
	LLMResponse   *llm.ChatResponse
	Err           error
	MiddleMessage string
}
