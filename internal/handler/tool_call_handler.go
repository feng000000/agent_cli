package handler

import "fmt"

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func HandleToolCall(s *runtime.Session) error {
	// DEBUG:
	{
		logger.Debugf("HandleToolCall Start >>>>>>>>>>>>>>>>\n")
		logger.Debugf("[HandleToolCall] message (before):\n")
		for _, message := range s.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
	}

	idChanMap := map[llm.ToolCall]chan string{}
	for _, tc := range s.Response.ToolCalls() {
		logger.Infof("exec tool: %v\n", tc.Function.Name)
		resCh := make(chan string)
		idChanMap[tc] = resCh

		tool, ok := s.Meta.ToolMap[tc.Function.Name]
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) not exists", tc.Function.Name, tc.ID,
			)
		}

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

	toolMsgs := []llm.Message{}
	for tc, ch := range idChanMap {
		res, ok := <-ch
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) execute failed", tc.Function.Name, tc.ID,
			)
		}

		logger.Debugf("tool %v result: %v\n", tc.Function.Name, res)
		toolMsgs = append(
			toolMsgs,
			llm.ToolResultMessage(tc.ID, res),
		)
	}

	// append query from state.InputChan
	{
		appendInput := s.MessageQueue.GetTypedInput(runtime.InputTypePrompt)
		if appendInput != nil && len(appendInput.Content) != 0 {
			toolMsgs= append(
				toolMsgs,
				llm.UserMessage(string(appendInput.Content)),
			)
		}
	}

	// request LLM
	// TODO: 长工具输出落盘, 仅返回路径
	err := s.CallLLM(toolMsgs...)
	if err != nil {
		return err
	}

	// DEBUG:
	{
		// fmt.Printf("[HandleToolCall] result: %v\n", ctx.Response.Content())
		logger.Debugf(
			"[HandleToolCall] HasToolCall: %v\n",
			s.Response.HasToolCalls(),
		)
		logger.Debugf("[HandleToolCall] message (after):\n")
		for _, message := range s.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
		logger.Debugf("HandleToolCall Done <<<<<<<<<<<<<<<<\n\n")
	}

	return nil
}
