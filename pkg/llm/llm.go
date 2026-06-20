package llm

import "context"

const (
	// RoleSystem 表示系统消息角色.
	RoleSystem = "system"

	// RoleUser 表示用户消息角色.
	RoleUser = "user"

	// RoleAssistant 表示模型消息角色.
	RoleAssistant = "assistant"

	// RoleTool 表示工具结果消息角色.
	RoleTool = "tool"

	// ToolTypeFunction 表示函数工具类型.
	ToolTypeFunction = "function"

	// ToolChoiceNone 表示禁用工具调用.
	ToolChoiceNone = "none"

	// ToolChoiceAuto 表示由模型自行决定是否调用工具.
	ToolChoiceAuto = "auto"

	// ToolChoiceRequired 表示要求模型必须调用工具.
	ToolChoiceRequired = "required"
)

// LLMClient 定义通用 LLM 对话接口.
type LLMClient interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ChatRequest 表示一次对话补全请求.
type ChatRequest struct {
	Model           string          `json:"model,omitempty"`
	Messages        []Message       `json:"messages"`
	Tools           []Tool          `json:"tools,omitempty"`
	ToolChoice      any             `json:"tool_choice,omitempty"`
	MaxTokens       *int            `json:"max_tokens,omitempty"`
	Temperature     *float64        `json:"temperature,omitempty"`
	TopP            *float64        `json:"top_p,omitempty"`
	ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
	Thinking        *Thinking       `json:"thinking,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	UserID          string          `json:"user_id,omitempty"`
}

// Message 表示一条对话消息.
type Message struct {
	Role             string     `json:"role"`
	Content          string     `json:"content,omitempty"`
	Name             string     `json:"name,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
}

// Tool 表示模型可调用的工具定义.
type Tool struct {
	Type     string       `json:"type"`
	Function FunctionTool `json:"function"`
}

// FunctionTool 表示函数工具的元数据和参数 schema.
type FunctionTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Strict      *bool          `json:"strict,omitempty"`
}

// ToolCall 表示模型返回的一次工具调用.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 表示模型要调用的函数名和 JSON 参数.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ResponseFormat 表示模型输出格式约束.
type ResponseFormat struct {
	Type string `json:"type"`
}

// Thinking 表示模型思考模式配置.
type Thinking struct {
	Type string `json:"type"`
}

// ChatResponse 表示一次对话补全响应.
type ChatResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

// Choice 表示模型返回的一个候选结果.
type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

// Usage 表示本次请求的 token 用量.
type Usage struct {
	PromptTokens          int                    `json:"prompt_tokens"`
	CompletionTokens      int                    `json:"completion_tokens"`
	TotalTokens           int                    `json:"total_tokens"`
	PromptCacheHitTokens  int                    `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int                    `json:"prompt_cache_miss_tokens"`
	CompletionDetails     CompletionTokenDetails `json:"completion_tokens_details"`
}

// CompletionTokenDetails 表示生成 token 的细分用量.
type CompletionTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// SystemMessage 创建系统消息.
func SystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

// UserMessage 创建用户消息.
func UserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// AssistantMessage 创建模型消息.
func AssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// ToolResultMessage 创建工具执行结果消息.
func ToolResultMessage(toolCallID, content string) Message {
	return Message{Role: RoleTool, ToolCallID: toolCallID, Content: content}
}

// FunctionToolDefinition 创建函数工具定义.
func FunctionToolDefinition(name, description string, parameters map[string]any) Tool {
	return Tool{
		Type: ToolTypeFunction,
		Function: FunctionTool{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}

// NamedToolChoice 创建指定函数工具的 tool_choice 参数.
func NamedToolChoice(functionName string) map[string]any {
	return map[string]any{
		"type": ToolTypeFunction,
		"function": map[string]string{
			"name": functionName,
		},
	}
}

// Message 返回第一个候选结果中的消息.
func (r *ChatResponse) Message() (Message, bool) {
	if r == nil || len(r.Choices) == 0 {
		return Message{}, false
	}

	return r.Choices[0].Message, true
}

// Content 返回第一个候选结果中的文本内容.
func (r *ChatResponse) Content() string {
	msg, ok := r.Message()
	if !ok {
		return ""
	}

	return msg.Content
}

// ToolCalls 返回第一个候选结果中的工具调用列表.
func (r *ChatResponse) ToolCalls() []ToolCall {
	msg, ok := r.Message()
	if !ok {
		return nil
	}

	return msg.ToolCalls
}

// HasToolCalls 判断第一个候选结果是否包含工具调用.
func (r *ChatResponse) HasToolCalls() bool {
	return len(r.ToolCalls()) > 0
}
