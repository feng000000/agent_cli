package handler

import "fmt"
import "context"

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleToolCall(ctx context.Context, state *runtime.AgentState) error {
	idChanMap := map[llm.ToolCall]chan string{}

	logger.Debugf("HandleToolCall Start >>>>>>>>>>>>>>>>\n")
	logger.Debugf("[HandleToolCall] message (before):\n")
	for _, message := range state.MessageParams {
		logger.Debugf("\tmessage: %+v\n", message)
	}

	for _, tc := range state.Response.ToolCalls() {
		logger.Infof("exec tool: %v\n", tc.Function.Name)
		resCh := make(chan string)
		idChanMap[tc] = resCh

		tool, ok := state.ToolMap[tc.Function.Name]
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) not exists", tc.Function.Name, tc.ID,
			)
		}

		tool.Execute(tc.Function.Arguments, resCh)

	}

	for tc, ch := range idChanMap {
		res, ok := <-ch
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) execute failed", tc.Function.Name, tc.ID,
			)
		}

		state.MessageParams = append(
			state.MessageParams,
			llm.ToolResultMessage(tc.ID, res),
		)
	}

	resp, err := state.LLMClient.Chat(
		ctx,
		llm.ChatRequest{
			Messages:   state.MessageParams,
			Tools:      state.ToolList(),
		},
	)
	if err != nil {
		return err
	}

	state.Response = resp
	assistantMsg, ok := resp.Message()
	if ok {
		state.MessageParams = append(state.MessageParams, assistantMsg)
	}

	// fmt.Printf("[HandleToolCall] result: %v\n", ctx.Response.Content())
	logger.Debugf("[HandleToolCall] HasToolCall: %v\n", state.Response.HasToolCalls())


	logger.Debugf("[HandleToolCall] message (after):\n")
	for _, message := range state.MessageParams {
		logger.Debugf("\tmessage: %+v\n", message)
	}
	logger.Debugf("HandleToolCall Done <<<<<<<<<<<<<<<<\n\n")

	return nil
}
