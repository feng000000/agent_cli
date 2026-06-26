package handler

import (
	"context"
	"myagent/internal/config"
	"myagent/internal/runtime"
	"myagent/pkg/llm"
	"myagent/pkg/logger"
	"time"
)

func HandleQuery(cfg config.ProjectConfig, state *runtime.AgentState) error {
	logger.Infof("Handle query")

	logger.Debugf("HandleQuery Start >>>>>>>>>>>>>>>>\n\n")
	logger.Debugf("[HandleQuery] query: %v\n", state.UserQuery)
	logger.Debugf("[HandleQuery] message (before):")
	for _, message := range state.MessageParams {
		logger.Debugf("\tmessage: %+v\n", message)
	}

	messages := []llm.Message{
		llm.SystemMessage(state.SystemPrompt),
		llm.UserMessage(state.UserQuery),
	}

	timeoutCtx, cancel := context.WithTimeout(
		context.Background(), 60*time.Second,
	)
	defer cancel()

	resp, err := state.LLMClient.Chat(
		timeoutCtx,
		llm.ChatRequest{
			Messages:   messages,
			Tools:      state.ToolList(),
			ToolChoice: llm.ToolChoiceAuto,
		},
	)

	// DEBUG:
	if !resp.HasToolCalls() {
		logger.Debugf("[CALL LLM] content:\n")
		logger.Debugf("%v\n", resp.Content())
	} else {
		logger.Debugf("[CALL LLM] tool calls:\n")
		for _, tc := range resp.ToolCalls() {
			logger.Debugf(
				"\tid: %s\ntype: %s\nfunction: %s\narguments: %s\n",
				tc.ID,
				tc.Type,
				tc.Function.Name,
				tc.Function.Arguments,
			)
		}
	}

	if err != nil {
		return err
	}

	state.Response = resp

	assistantMsg, ok := resp.Message()
	if ok {
		state.MessageParams = append(state.MessageParams, assistantMsg)
	}


	logger.Debugf("[HandleQuery] message (after):\n")
	for _, message := range state.MessageParams {
		logger.Debugf("\tmessage: %+v\n", message)
	}
	logger.Debugf("HandleQuery Done <<<<<<<<<<<<<<<<\n\n")

	lastMsg := state.MessageParams[len(state.MessageParams) - 1]
	if lastMsg.ReasoningContent != "" {
		state.OutputChan <- runtime.AgentResponse{
			RespType: runtime.AgentRespTypeMiddleMsg,
			MiddleMessage: lastMsg.ReasoningContent,
		}
	}

	return nil
}
