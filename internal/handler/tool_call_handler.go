package handler

import "fmt"
import "sync"

import "myagent/internal/runtime"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

func findToolCallID(id string, toolMsgs []runtime.ToolMessage) int {
	for i, msg := range toolMsgs {
		if msg.ToolCallID == id {
			return i
		}
	}
	return -1
}

func HandleToolCall(s *runtime.Session) error {
	// DEBUG:
	{
		logger.Debugf("HandleToolCall Start >>>>>>>>>>>>>>>>\n")
		logger.Debugf("[HandleToolCall] message (before):\n")
		for _, message := range s.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
	}

	idChanMap := map[llm.ToolCall]chan runtime.ToolMessage{}
	sessionRWMu := &sync.RWMutex{}


	for _, tc := range s.Response.ToolCalls() {
		// 已经执行过
		if findToolCallID(tc.ID, s.ToolMessages) != -1 {
			continue
		}

		logger.Infof("exec tool: %v\n", tc.Function.Name)
		resCh := make(chan runtime.ToolMessage)
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
					resCh <- runtime.ToolMessage{
						Result: fmt.Sprintf(
							"exec tool %v failed: %v", tc.Function.Name, r,
						),
					}
				}
			}()
			res := runtime.ExecTool(s, sessionRWMu, tc.ID, tool, tc.Function.Arguments)
			resCh <- res
		}()

	}

	llmToolMsgs := make([]llm.Message, 0, len(idChanMap))
	toolMsgs := make([]runtime.ToolMessage, 0, len(idChanMap))

	// wait tool exec results
	for tc, ch := range idChanMap {
		toolMsg, ok := <-ch
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) execute failed", tc.Function.Name, tc.ID,
			)
		}

		logger.Debugf("tool %v result: %v\n", tc.Function.Name, toolMsg)
		toolMsgs = append(toolMsgs, toolMsg)
		llmToolMsgs = append(
			llmToolMsgs,
			llm.ToolResultMessage(tc.ID, toolMsg.Result),
		)
	}

	// append query from state.InputChan
	{
		appendInput := s.MessageQueue.GetTypedInput(runtime.InputTypePrompt)
		if appendInput != nil && len(appendInput.Content) != 0 {
			llmToolMsgs= append(
				llmToolMsgs,
				llm.UserMessage(string(appendInput.Content)),
			)
		}
	}

	// request LLM
	err := s.CallLLM(llmToolMsgs...)
	if err != nil {
		return err
	}

	// update session.ToolMessages
	s.ToolMessages = append(s.ToolMessages, toolMsgs...)

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
