package runtime

import "os"
import "encoding/json"
import "context"

import "myagent/internal/session/userinput"
import "myagent/internal/session/response"
import "myagent/internal/session/meta"
import "myagent/pkg/llm"

// NewRuntime
func NewRuntime(
	meta *meta.Meta,
	mq *userinput.MessageQueue,
	output chan *response.AgentResponse,
) (*Runtime, error) {
	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  meta.LLM.APIKey,
			BaseURL: meta.LLM.BaseURL,
			Model:   meta.LLM.Model,
		},
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	r := &Runtime{
		Ctx:          ctx,
		Cancel:       cancel,
		MessageQueue: mq,
		outputChan:   output,
		llmClient:    llmClient,
	}

	return r, nil
}

// LoadRuntime
func LoadRuntime(
	path string,
	meta *meta.Meta,
	mq *userinput.MessageQueue,
	output chan *response.AgentResponse,
) (*Runtime, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	runtime := &Runtime{}
	if err := decoder.Decode(runtime); err != nil {
		return nil, err
	}


	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  meta.LLM.APIKey,
			BaseURL: meta.LLM.BaseURL,
			Model:   meta.LLM.Model,
		},
	)
	if err != nil {
		return nil, err
	}
	runtime.llmClient = llmClient

	ctx, cancel := context.WithCancel(context.Background())

	runtime.Ctx, runtime.Cancel = ctx, cancel


	return runtime, nil
}

type Runtime struct {
	llmClient llm.LLMClient `json:"-"`
	// MessageQueue 可以直接获取追加信息
	MessageQueue *userinput.MessageQueue `json:"-"`
	// outputChan emit ACP 事件
	outputChan chan *response.AgentResponse `json:"-"`

	Ctx    context.Context    `json:"-"`
	Cancel context.CancelFunc `json:"-"`

	Response     *llm.ChatResponse `json:"response"`
	SkillMap     map[string]string `json:"skill_map"`
	LoadedSkills map[string]bool   `json:"loaded_skills"`
	Messages     []llm.Message     `json:"messages"`
	Usage        *llm.Usage        `json:"usage"`
	ContextSize  int               `json:"context_size"`
}

// RawLLMClient 获取原始 LLMClient
func (r *Runtime) RawLLMClient() llm.LLMClient {
	return r.llmClient
}

// Emit 发送 AgentResponse 事件
func (r *Runtime) Emit(resp *response.AgentResponse) {
	r.outputChan <- resp
}
