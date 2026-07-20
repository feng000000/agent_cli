package runtime

import "context"
import _ "embed"
import "fmt"
import "path"
import "os"
import "errors"
import "strings"
import "text/template"
import "time"

import "myagent/pkg/llm"

//go:embed prompts/system_prompt.md
var systemPromptTemplate string

// LoadSkills TODO: 注入 skill 描述
func (s *Session) LoadSkills() error {
	entries, err := os.ReadDir()
	if err != nil {
		return err
	}

	// register skill map
	// TODO: the loading is really happened in the loop
	for _, skillDir := range entries {
		if !skillDir.IsDir() {
			continue
		}
		skill, err := os.ReadFile(path.Join(skillDir.Name(), "SKILL.md"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// B 下没有 SKILL.md，直接跳过。
				continue
			}
			return fmt.Errorf("read %q: %w", skillDir.Name(), err)
		}

		// TODO: parse skill Markdown meta data (yaml)

	}
}


// SystemPrompt 计算SystemPrompt
func (s *Session) SystemPrompt() (string, error) {
	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return "", err
	}

	workingDir, err:= os.Getwd()

	memoryData, err := os.ReadFile(s.Meta.Persistence.MemoryPath)
	userInfoData, err := os.ReadFile(s.Meta.Persistence.UserInfoPath)

	s.LoadSkills()

	sb := strings.Builder{}
	tmpl.Execute(
		&sb,
		map[string]string{
			"Date":      time.Now().Format(time.DateOnly),
			"WorkingDir": workingDir,
			"Memory":    string(memoryData),
			"UserInfo":  string(userInfoData),
			"SkillDir":  s.Meta.Persistence.SkillDir,
		},
	)


	return sb.String(), nil
}


type Compressor interface {
	Compress(ctx context.Context, state *Session) error
}

//go:embed prompts/compress_system_prompt.md
var compressSystemPrompt string

// LLMCompressor 大模型压缩总结
type LLMCompressor struct{}

// Compress 使用 大模型压缩总结
func (lc *LLMCompressor) Compress(
	ctx context.Context, state *Session,
) error {
	resp, err := state.LLMClient.Chat(
		ctx,
		llm.ChatRequest{
			Messages: append(
				state.Messages,
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
	state.Messages = []llm.Message{llm.AssistantMessage(summary)}

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
