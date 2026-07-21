package handler

import "context"
import "time"

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleQuery(
	ctx context.Context,
	query string,
	s *runtime.Session,
	output chan *runtime.AgentResponse,
) error {
	logger.Infof("Handle query")

	// DEBUG:
	{
		logger.Debugf("HandleQuery Start >>>>>>>>>>>>>>>>\n\n")
		logger.Debugf("[HandleQuery] query: %v\n", query)
		logger.Debugf("[HandleQuery] message (before):")
		for _, message := range s.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
	}
	s.AppendMessage(llm.UserMessage(query))

	var gotRespCh chan bool = make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gotRespCh:
				output <- &runtime.AgentResponse{
					RespType:      runtime.AgentRespTypeMiddleMsg,
					MiddleMessage: "got llm resp",
				}
				return
			case <-ticker.C:
				output <- &runtime.AgentResponse{
					RespType:      runtime.AgentRespTypeMiddleMsg,
					MiddleMessage: "waiting",
				}
			}
		}
	}()

	resp, err := s.LLMClient.Chat(
		ctx,
		llm.ChatRequest{
			Messages:   s.Messages,
			Tools:      s.ToolList(),
			ToolChoice: llm.ToolChoiceAuto,
		},
	)
	gotRespCh <- true

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

	s.UpdateResponse(resp)


	assistantMsg, ok := resp.Message()
	if ok {
		s.AppendMessage(assistantMsg)
	}

	logger.Debugf("[HandleQuery] message (after):\n")
	for _, message := range s.Messages {
		logger.Debugf("\tmessage: %+v\n", message)
	}
	logger.Debugf("HandleQuery Done <<<<<<<<<<<<<<<<\n\n")

	lastMsg := s.Messages[len(s.Messages)-1]
	if lastMsg.ReasoningContent != "" {
		output <- &runtime.AgentResponse{
			RespType:      runtime.AgentRespTypeMiddleMsg,
			MiddleMessage: lastMsg.ReasoningContent,
		}
	}

	return nil
}
