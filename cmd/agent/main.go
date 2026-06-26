package main

import (
	"log"
	"os"

	"myagent/internal/agent"
	"myagent/internal/config"
	"myagent/pkg/logger"
)


func main() {
	cfg, err := config.InitAgentConfig("./config.jsonc")
	if err != nil {
		log.Fatalf("load config failed: %v\n", err)
	}

	if err := logger.InitLogger("DEBUG", cfg.LogPath); err != nil {
		panic(err)
	}
	defer logger.Sync()


	a := agent.Agent{Config: cfg}

	if agent.StartSimpleUI(a, os.Stdin, os.Stdout) != nil {
		log.Fatalf("%v", err)
	}
}
