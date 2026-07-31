package meta

import "os"
import "encoding/json"

import "github.com/google/uuid"

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

}


// LoadMeta 从给定 path 中加载Meta
func LoadMeta(path string) (*Meta, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	meta := &Meta{}
	if err := decoder.Decode(meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// NewMeta
func NewMeta() *Meta {
	return &Meta{
		SessionID: uuid.NewString(),
		AgentMode: AgentModePlan,
		Persistence: &Persistence{
			PersistenceDir: "./.myagent/persistence",
			MemoryPath:     "./.myagent/persistence/memory.md",
			UserInfoPath:   "./.myagent/persistence/user_info.md",
			SkillDir:       "./.myagent/persistence/skills/",
		},
	}
}
