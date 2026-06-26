package context

import (
	_ "embed"
	"os"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"myagent/internal/runtime"
	"myagent/internal/config"
	"myagent/pkg/llm"
)

//go:embed prompts/system_prompt.md
var systemPromptTemplate string

type systemPromptParams struct {

}


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
			"Date": time.Now().Format(time.DateOnly),
			"Workspace": cfg.Workspace.WorkspaceDir,
			"Memory": string(memoryData),
			"UserInfo": string(userInfoData),
			"SkillDir": cfg.Workspace.SkillDir,
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
type LLMCompressor struct {}

// Compress 使用大模型总结历史信息
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

// HybridCompressor 混合压缩, 最近k条保留, 之前的大模型压缩
type HybridCompressor struct {
	topK int
}

// extractSystemPrompt 提取
func (hc *HybridCompressor) extractSystemPrompt(
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

// splitCompressMessages 拆分 待压缩的消息
func (hc *HybridCompressor) splitCompressMessages(
	messages []llm.Message,
) (toCompress []llm.Message, kept []llm.Message) {
	if len(messages) == 0 {
		return messages, []llm.Message{}
	}
	pos := len(messages) - 1
	for i := pos; i >= 0; i-- {
		if messages[i].Role == llm.RoleUser &&
			(pos == len(messages)-1 || len(messages)-i <= hc.topK) {
			pos = i
		}
	}

	return messages[pos:], messages[pos+1:]
}

// Compress 混合压缩, 最近 <=k条保留(以用户消息划分), 之前的记录使用大模型压缩
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
