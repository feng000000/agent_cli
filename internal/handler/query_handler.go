package handler

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleQuery(query string, s *runtime.Session) error {
	logger.Infof("Handle query")

	return s.CallLLM(llm.UserMessage(query))
}
