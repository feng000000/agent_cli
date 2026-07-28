package response

import "myagent/pkg/llm"

type ResponseTypeEnum string

const (
	AgentRespTypeCmd       = "command"
	AgentRespTypeLLM       = "llm_response"
	AgentRespTypeError     = "error"
	AgentRespTypeMiddleMsg = "middle_message"
	AgentRespTypeConfirm   = "confirm"
)

type AgentResponse struct {
	RespType ResponseTypeEnum

	CmdResult     string
	LLMResponse   *llm.ChatResponse
	Err           error
	MiddleMessage string

	ConfirmMessage string
}
