package runtime

import "context"

import "myagent/internal/session/userinput"
import "myagent/internal/session/response"
import "myagent/pkg/llm"

type Runtime struct {
	llmClient llm.LLMClient
	// MessageQueue 可以直接获取追加信息
	MessageQueue *userinput.MessageQueue
	// outputChan emit ACP 事件
	outputChan chan *response.AgentResponse

	Ctx      context.Context
	Cancel   context.CancelFunc
	Response *llm.ChatResponse
}



// TODO: NewRuntime
func NewRuntime() *Runtime {
	return nil
}

// TODO: LoadRuntime
func LoadRuntime() *Runtime {
	return nil
}
