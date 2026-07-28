package session

import "context"
import _ "embed"
import "fmt"
import "os"
import "strings"
import "text/template"
import "time"

import "myagent/pkg/llm"

//go:embed prompts/system_prompt.md
var systemPromptTemplate string

// LoadSkills TODO: 注入 skill 描述 (接入 Skill repository protocol)
func (s *Session) LoadSkillManifests() error {
	return fmt.Errorf("not implemented")
}

// TODO: 等后续接入 SRP 时实现
// 用户显式指定时加载, load-skill 工具调用时加载
func (s *Session) LoadSkills() error {
	return fmt.Errorf("not implemented")
}


// SystemPrompt 计算SystemPrompt
func (s *Session) SystemPrompt() (string, error) {
	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return "", err
	}

	workingDir, err:= os.Getwd()

	skillDir := s.Meta.Persistence.SkillDir
	memoryData, err := os.ReadFile(s.Meta.Persistence.MemoryPath)
	userInfoData, err := os.ReadFile(s.Meta.Persistence.UserInfoPath)

	// s.LoadSkillManifests()

	sb := strings.Builder{}
	tmpl.Execute(
		&sb,
		map[string]string{
			"Date":      time.Now().Format(time.DateOnly),
			"WorkingDir": workingDir,
			"Memory":    string(memoryData),
			"UserInfo":  string(userInfoData),
			"SkillDir":  skillDir,
		},
	)

	return sb.String(), nil
}


type Compressor interface {
	Compress(ctx context.Context, s *Session) error
}

//go:embed prompts/compress_system_prompt.md
var compressSystemPrompt string

// LLMCompressor 大模型压缩总结
type LLMCompressor struct{}

// Compress 使用 大模型压缩总结
func (lc *LLMCompressor) Compress(
	ctx context.Context, s *Session,
) error {
	resp, err := s.RawLLMClient().Chat(
		ctx,
		llm.ChatRequest{
			Messages: append(
				s.Messages,
				llm.UserMessage(compressSystemPrompt),
			),
		},
	)
	if err != nil {
		return err
	}

	summary := fmt.Sprintf(
		"<history_context_summary>%s</history_context_summary>",
		resp.Content(),
	)
	s.Messages = []llm.Message{llm.AssistantMessage(summary)}

	s.ForceUpdateContextInfo(
		[]llm.Message{llm.AssistantMessage(summary)},
		s.Usage.Append(&resp.Usage),
		resp.Usage.Completion,
	)

	return nil
}

// TruncateCompressor 直接截断, 只保留 topK 轮消息
type TruncateCompressor struct {
	TopK int
}


// Compress 直接截断, 只保留 topK 轮消息
func (tc *TruncateCompressor) Compress(
	ctx context.Context, state *Session,
) error {
	sysPrompt, messages := extractSystemPrompt(state.Messages)
	_, kept := splitCompressMessages(tc.TopK, messages)

	// system prompt + kept messages
	state.Messages = append([]llm.Message{*sysPrompt}, kept...)

	return nil
}

// HybridCompressor 混合压缩, 最近 <=k条保留(以用户消息划分),
// 之前的记录使用大模型压缩
type HybridCompressor struct {
	TopK int
}

// Compress 混合压缩, 最近 <=k条保留(以用户消息划分),
// 之前的记录使用大模型压缩
func (hc *HybridCompressor) Compress(
	ctx context.Context,
	state *Session,
) error {
	sysPrompt, messages := extractSystemPrompt(state.Messages)
	toCompress, kept := splitCompressMessages(hc.TopK, messages)

	state.Messages = toCompress

	(&LLMCompressor{}).Compress(ctx, state)

	// system prompt + compressed messages + kept messages
	state.Messages = append(
		append([]llm.Message{*sysPrompt}, state.Messages...),
		kept...,
	)

	return nil
}


// extractSystemPrompt 提取给定 []llm.Message 中的系统提示词
func extractSystemPrompt(
	messages []llm.Message,
) (sysPrompt *llm.Message, resMessages []llm.Message) {
	for i, msg := range messages {
		if msg.Role == llm.RoleSystem {
			sysPrompt = &messages[i]
			resMessages = messages[i+1:]
			return
		}
	}

	return nil, messages
}


// splitCompressMessages 提取给定 messages 中的 topK 条信息
// 返回待压缩信息和非压缩信息
func splitCompressMessages(
	topK int,
	messages []llm.Message,
) (toCompress []llm.Message, kept []llm.Message) {
	if len(messages) == 0 {
		return messages, []llm.Message{}
	}
	pos := len(messages) - 1
	for i := pos; i >= 0; i-- {
		if messages[i].Role == llm.RoleUser &&
			(pos == len(messages)-1 || len(messages)-i <= topK) {
			pos = i
		}
	}

	return messages[pos:], messages[pos+1:]
}
