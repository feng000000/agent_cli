package config

import "os"
import "fmt"
import "encoding/json"

type LLMConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type WorkspaceConfig struct {
	WorkspaceDir string `json:"workspace_dir"`
	MemoryPath   string `json:"memory_path"`
	UserInfoPath string `json:"userinfo_path"`
	SkillDir     string `json:"skill_dir"`
}

type ProjectConfig struct {
	LogPath string `json:"log_path"`

	LLM       LLMConfig       `json:"llm"`
	Workspace WorkspaceConfig `json:"workspace"`
}

func defaultAgentConfig() ProjectConfig {
	return ProjectConfig{
		LogPath: "",
		Workspace: WorkspaceConfig{
			WorkspaceDir: "./workspace",
			MemoryPath:   "./workspace/.memory.md",
			UserInfoPath: "./workspace/.user_info.md",
			SkillDir:     "./workspace/skills/",
		},
	}
}

func InitAgentConfig(path string) (ProjectConfig, error) {
	cfg := defaultAgentConfig()

	file, err := os.Open(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return ProjectConfig{}, fmt.Errorf("decode config: %w", err)
	}

	return cfg, nil
}
