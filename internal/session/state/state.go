package state

import "myagent/internal/session/tool"

import "myagent/pkg/llm"

type State struct {
	LocalMemory string

	// 历史记录/上下文
	Messages     []llm.Message

	// 已执行的 tool Message
	ToolMessages []tool.ToolMessage

	// 已加载的Skill
	LoadedSkill  map[string]bool

	// 当前会话用量
	Usage        *llm.Usage

	// 当前上下文大小
	ContextSize  int
}


// TODO: NewState
func NewState() *State {
	return nil
}

// TODO: LoadState
func LoadState() *State {
	return nil
}
