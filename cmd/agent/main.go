package main

import "os"
import "log"

import "myagent/internal/agent"
import "myagent/internal/config"
import "myagent/pkg/logger"


func main() {
	cfg, err := config.InitAgentConfig("./config.jsonc")
	if err != nil {
		log.Fatalf("load config failed: %v\n", err)
	}

	if err := logger.InitLogger("DEBUG", cfg.LogPath); err != nil {
		panic(err)
	}
	defer logger.Sync()


	a := agent.Agent{
		Config: cfg,
	}

	if a.Run(os.Stdin, os.Stdout) != nil {
		log.Fatalf("%v", err)
	}
}
