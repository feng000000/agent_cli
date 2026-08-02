package main

import "log"
import "os"
import "flag"

import "agentcli/internal/agent"
import "agentcli/internal/config"
import "agentcli/pkg/logger"


func main() {
	cfg, err := config.InitAgentConfig("./config.jsonc")
	if err != nil {
		log.Fatalf("load config failed: %v\n", err)
	}

	if err := logger.InitLogger("DEBUG", cfg.LogPath); err != nil {
		panic(err)
	}
	defer logger.Sync()

	var sessionID string

	flag.StringVar(&sessionID, "r", "", "session id")
	flag.Parse()

	logger.Debugf("got session param: %v\n", sessionID)
	if agent.StartSimpleUI(os.Stdin, os.Stdout, sessionID) != nil {
		log.Fatalf("%v", err)
	}
}
