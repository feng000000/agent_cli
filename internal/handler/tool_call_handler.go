package handler

import "fmt"
import "context"

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleToolCall(
	ctx context.Context,
	state *runtime.Session,
	messageQueue *runtime.MessageQueue,
) error {
	// DEBUG:
	{
		logger.Debugf("HandleToolCall Start >>>>>>>>>>>>>>>>\n")
		logger.Debugf("[HandleToolCall] message (before):\n")
		for _, message := range state.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
	}

	idChanMap := map[llm.ToolCall]chan string{}
	for _, tc := range state.Response.ToolCalls() {
		logger.Infof("exec tool: %v\n", tc.Function.Name)
		resCh := make(chan string)
		idChanMap[tc] = resCh

		tool, ok := state.Meta.ToolMap[tc.Function.Name]
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) not exists", tc.Function.Name, tc.ID,
			)
		}

		// TODO: 检查 permission, emit tool call event
		go func() {
			defer func() {
				if r := recover(); r != nil {
					resCh <- fmt.Sprintf(
						"exec tool %v failed: %v", tc.Function.Name, r,
					)
				}
			}()
			res := tool.Execute(tc.Function.Arguments)
			resCh <- res
		}()

	}

	for tc, ch := range idChanMap {
		res, ok := <-ch
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) execute failed", tc.Function.Name, tc.ID,
			)
		}

		logger.Debugf("tool %v result: %v\n", tc.Function.Name, res)
		state.Messages = append(
			state.Messages,
			llm.ToolResultMessage(tc.ID, res),
		)
	}

	// append query from state.InputChan
	{
		appendInput := messageQueue.GetTypedInput(runtime.InputTypePrompt)
		if appendInput != nil && len(appendInput.Content) != 0 {
			state.Messages = append(
				state.Messages,
				llm.UserMessage(string(appendInput.Content)),
			)
		}
	}

	// request LLM
	// TODO: call LLM, update state.Response, update state.MessageParam 封装到一起
	// 还有 update usage && context size
	{
		req := llm.ChatRequest{
			Messages: state.Messages,
			Tools:    state.ToolList(),
		}
		resp, err := state.LLMClient.Chat(ctx, req)
		if err != nil {
			return err
		}

		state.Response = resp
		if assistantMsg, ok := resp.Message(); ok {
			state.Messages = append(state.Messages, assistantMsg)
		}
	}

	// DEBUG:
	{
		// fmt.Printf("[HandleToolCall] result: %v\n", ctx.Response.Content())
		logger.Debugf(
			"[HandleToolCall] HasToolCall: %v\n",
			state.Response.HasToolCalls(),
		)
		logger.Debugf("[HandleToolCall] message (after):\n")
		for _, message := range state.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
		logger.Debugf("HandleToolCall Done <<<<<<<<<<<<<<<<\n\n")
	}

	return nil
}
