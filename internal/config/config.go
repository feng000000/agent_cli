package config

import "os"
import "fmt"
import "encoding/json"


type ProjectConfig struct {
	LogPath string `json:"log_path"`

}

func defaultAgentConfig() ProjectConfig {
	return ProjectConfig{
		LogPath: "",
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
