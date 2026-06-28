package llm

import "bytes"
import "context"
import "encoding/json"
import "errors"
import "fmt"
import "io"
import "net/http"
import "os"
import "strings"
import "time"

const (
	// DefaultDeepSeekBaseURL 是 DeepSeek OpenAI-compatible API 的默认地址.
	DefaultDeepSeekBaseURL = "https://api.deepseek.com"

	// DefaultDeepSeekModel 是默认使用的 DeepSeek 模型.
	DefaultDeepSeekModel = "deepseek-v4-flash"
)

// DeepSeekConfig 表示 DeepSeek 客户端配置.
type DeepSeekConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// DeepSeekClient 实现 DeepSeek 的 Chat Completions 调用.
type DeepSeekClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// DeepSeekAPIError 表示 DeepSeek API 返回的错误.
type DeepSeekAPIError struct {
	StatusCode int
	Message    string
	Type       string
	Param      string
	Code       string
	Body       string
}

// Error 返回 DeepSeek API 错误的可读描述.
func (e *DeepSeekAPIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("deepseek api error: status=%d message=%s type=%s code=%s param=%s", e.StatusCode, e.Message, e.Type, e.Code, e.Param)
	}

	return fmt.Sprintf("deepseek api error: status=%d body=%s", e.StatusCode, e.Body)
}

// NewDeepSeekClient 创建 DeepSeek 客户端.
func NewDeepSeekClient(cfg DeepSeekConfig) (*DeepSeekClient, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	}
	if apiKey == "" {
		return nil, errors.New("deepseek api key is required")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultDeepSeekBaseURL
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = DefaultDeepSeekModel
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}

	return &DeepSeekClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		httpClient: httpClient,
	}, nil
}

// Chat 调用 DeepSeek Chat Completions API.
func (c *DeepSeekClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if c == nil {
		return nil, errors.New("deepseek client is nil")
	}
	if len(req.Messages) == 0 {
		return nil, errors.New("chat request requires at least one message")
	}

	if strings.TrimSpace(req.Model) == "" {
		req.Model = c.model
	}
	normalizeTools(req.Tools)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create chat request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send chat request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read chat response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseDeepSeekAPIError(resp.StatusCode, respBody)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("decode chat response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, errors.New("chat response has no choices")
	}

	return &chatResp, nil
}

// normalizeTools 填充工具定义中的默认类型.
func normalizeTools(tools []Tool) {
	for i := range tools {
		if tools[i].Type == "" {
			tools[i].Type = ToolTypeFunction
		}
	}
}

// parseDeepSeekAPIError 解析 DeepSeek API 错误响应.
func parseDeepSeekAPIError(statusCode int, body []byte) error {
	apiErr := &DeepSeekAPIError{StatusCode: statusCode, Body: string(body)}

	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   string `json:"param"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		apiErr.Message = payload.Error.Message
		apiErr.Type = payload.Error.Type
		apiErr.Param = payload.Error.Param
		apiErr.Code = payload.Error.Code
	}

	return apiErr
}
