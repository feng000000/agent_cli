package config

import "os"
import "fmt"
import "encoding/json"

type WorkspaceConfig struct {
	WorkspaceDir string `json:"workspace_dir"`
	MemoryPath   string `json:"memory_path"`
	UserInfoPath string `json:"userinfo_path"`
	SkillDir     string `json:"skill_dir"`
}

type AgentConfig struct {
	LogPath string `json:"log_path"`

	Workspace WorkspaceConfig `json:"workspace"`
}

func defaultAgentConfig() AgentConfig {
	return AgentConfig{
		LogPath: "",
		Workspace: WorkspaceConfig{
			WorkspaceDir: "./workspace",
			MemoryPath:   "./workspace/.memory.md",
			UserInfoPath: "./workspace/.user_info.md",
			SkillDir:     "./workspace/skills/",
		},
	}
}

func InitAgentConfig(path string) (AgentConfig, error) {
	cfg := defaultAgentConfig()

	file, err := os.Open(path)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return AgentConfig{}, fmt.Errorf("decode config: %w", err)
	}

	return cfg, nil
}
