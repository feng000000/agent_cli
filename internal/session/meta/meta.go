package meta


import "github.com/google/uuid"

import "myagent/internal/session/tool"

type AgentModeEnum string

const (
	AgentModePlan     AgentModeEnum = "plan"
	AgentModeAutoEdit AgentModeEnum = "auto-edit"
	AgentModeAutoExec AgentModeEnum = "auto-exec"
)

type Persistence struct {
	PersistenceDir string `json:"persistence_dir"`
	MemoryPath     string `json:"memory_path"`
	UserInfoPath   string `json:"userinfo_path"`
	SkillDir       string `json:"skill_dir"`
}

type LLMConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type Meta struct {
	SessionID           string        `json:"session_id"`
	LLM                 LLMConfig     `json:"llm"`
	AgentMode           AgentModeEnum `json:"agent_mode"`
	MaxTokensToCompress int           `json:"max_tokens_to_compress"`
	Persistence         *Persistence  `json:"persistence"`

	ToolMap  map[string]tool.Tool `json:"-"`
	SkillMap map[string]string
}


// TODO: LoadMeta
func LoadMeta() *Meta {
	return nil
}

// TODO: NewMeta
func NewMeta() *Meta {
	// TODO: default tools
	listDir := &tool.ListDirTool{}
	readFile := &tool.ReadFileTool{}

	return &Meta{
		SessionID: uuid.NewString(),
		AgentMode: AgentModePlan,
		ToolMap: map[string]tool.Tool{
			listDir.Name():  listDir,
			readFile.Name(): readFile,
		},
		Persistence: &Persistence{
			PersistenceDir: "./.myagent/persistence",
			MemoryPath:     "./.myagent/persistence/memory.md",
			UserInfoPath:   "./.myagent/persistence/user_info.md",
			SkillDir:       "./.myagent/persistence/skills/",
		},
	}
}
