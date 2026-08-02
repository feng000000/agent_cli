package handler

import "agentcli/internal/session"
import "agentcli/pkg/llm"
import "agentcli/pkg/logger"

func HandleQuery(query string, s *session.Session) error {
	logger.Infof("Handle query")

	return s.CallLLM(llm.UserMessage(query))
}
