package handler

import "myagent/internal/session"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleQuery(query string, s *session.Session) error {
	logger.Infof("Handle query")

	return s.CallLLM(llm.UserMessage(query))
}
