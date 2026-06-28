package context

import "context"
import _ "embed"
import "fmt"
import "os"
import "strings"
import "text/template"
import "time"

import "myagent/internal/config"
import "myagent/internal/runtime"
import "myagent/pkg/llm"

//go:embed prompts/system_prompt.md
var systemPromptTemplate string

// TODO: 确定 skill 结构, 注入 skill 描述
func GetSystemPrompt(cfg config.ProjectConfig) (string, error) {
	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return "", err
	}

	memoryData, err := os.ReadFile(cfg.Workspace.MemoryPath)
	userInfoData, err := os.ReadFile(cfg.Workspace.UserInfoPath)

	sb := strings.Builder{}
	tmpl.Execute(
		&sb,
		map[string]string{
			"Date":      time.Now().Format(time.DateOnly),
			"Workspace": cfg.Workspace.WorkspaceDir,
			"Memory":    string(memoryData),
			"UserInfo":  string(userInfoData),
			"SkillDir":  cfg.Workspace.SkillDir,
		},
	)

	return sb.String(), nil
}

type Compressor interface {
	Compress(ctx context.Context, state *runtime.AgentState) error
}

//go:embed prompts/compress_system_prompt.md
var compressSystemPrompt string

// LLMCompressor 大模型压缩总结
type LLMCompressor struct{}

func (lc *LLMCompressor) Compress(
	ctx context.Context, state *runtime.AgentState,
) error {
	resp, err := state.LLMClient.Chat(
		ctx,
		llm.ChatRequest{
			Messages: append(
				state.MessageParams,
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
	state.MessageParams = []llm.Message{llm.AssistantMessage(summary)}

	return nil
}

// TruncateCompressor 直接截断, 只保留 topK 轮消息
type TruncateCompressor struct {
	topK int
}

func (tc *TruncateCompressor) extractSystemPrompt(
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

func (tc *TruncateCompressor) splitCompressMessages(
	messages []llm.Message,
) (toCompress []llm.Message, kept []llm.Message) {
	if len(messages) == 0 {
		return messages, []llm.Message{}
	}
	pos := len(messages) - 1
	for i := pos; i >= 0; i-- {
		if messages[i].Role == llm.RoleUser &&
			(pos == len(messages)-1 || len(messages)-i <= tc.topK) {
			pos = i
		}
	}

	return messages[pos:], messages[pos+1:]
}

func (tc *TruncateCompressor) Compress(
	ctx context.Context, state *runtime.AgentState,
) error {
	sysPrompt, messages := tc.extractSystemPrompt(state.MessageParams)
	_, kept := tc.splitCompressMessages(messages)

	// system prompt + kept messages
	state.MessageParams = append([]llm.Message{*sysPrompt}, kept...)

	return nil
}

// HybridCompressor 混合压缩, 最近 <=k条保留(以用户消息划分),
// 之前的记录使用大模型压缩
type HybridCompressor struct {
	TruncateCompressor
}

func (hc *HybridCompressor) Compress(
	ctx context.Context,
	state *runtime.AgentState,
) error {
	sysPrompt, messages := hc.extractSystemPrompt(state.MessageParams)
	toCompress, kept := hc.splitCompressMessages(messages)

	state.MessageParams = toCompress

	(&LLMCompressor{}).Compress(ctx, state)

	// system prompt + compressed messages + kept messages
	state.MessageParams = append(
		append([]llm.Message{*sysPrompt}, state.MessageParams...),
		kept...,
	)

	return nil
}
