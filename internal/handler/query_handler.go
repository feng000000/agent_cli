package handler

import "context"
import "time"

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleQuery(ctx *runtime.LoopContext) error {
	logger.Infof("Handle query")

	logger.Debugf("HandleQuery Start >>>>>>>>>>>>>>>>\n\n")
	logger.Debugf("[HandleQuery] query: %v\n", ctx.Query)
	logger.Debugf("[HandleQuery] message (before):")
	for _, message := range ctx.MessageParams {
		logger.Debugf("\tmessage: %+v\n", message)
	}

	messages := []llm.Message{llm.UserMessage(ctx.Query)}

	timeoutCtx, cancel := context.WithTimeout(
		context.Background(), 60*time.Second,
	)
	defer cancel()

	resp, err := ctx.LLMClient.Chat(
		timeoutCtx,
		llm.ChatRequest{
			Messages:   messages,
			Tools:      ctx.ToolList(),
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

	ctx.Query = ""
	ctx.Response = resp

	assistantMsg, ok := resp.Message()
	if ok {
		ctx.MessageParams = append(ctx.MessageParams, assistantMsg)
	}


	logger.Debugf("[HandleQuery] message (after):\n")
	for _, message := range ctx.MessageParams {
		logger.Debugf("\tmessage: %+v\n", message)
	}
	logger.Debugf("HandleQuery Done <<<<<<<<<<<<<<<<\n\n")

	lastMsg := ctx.MessageParams[len(ctx.MessageParams) - 1]
	if lastMsg.ReasoningContent != "" {
		ctx.OutputChan <- lastMsg.ReasoningContent
	}

	return nil
}
