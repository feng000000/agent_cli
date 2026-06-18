package myagent

import (
	"log"
	"myagent/agent"
	"myagent/config"
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

	err = a.AgentLoop(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
