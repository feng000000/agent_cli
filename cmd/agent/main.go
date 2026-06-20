package main

import (
	"log"
	"myagent/internal/agent"
	"myagent/internal/config"
	"os"
)


func main() {
	cfg, err := config.InitAgentConfig("./config.json")
	if err != nil {
		log.Fatalf("load config failed: %v\n", err)
	}

	a := agent.Agent{
		Config: cfg,
	}

	if a.Run(os.Stdin, os.Stdout) != nil {
		log.Fatalf("%v", err)
	}
}
