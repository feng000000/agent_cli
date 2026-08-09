package handler

import "fmt"
import "sync"

import "agentcli/internal/tool"
import "agentcli/internal/session"
import "agentcli/internal/session/userinput"
import "agentcli/pkg/llm"
import "agentcli/pkg/logger"

func findToolCallID(id string, toolMsgs []tool.ToolResult) int {
	for i, msg := range toolMsgs {
		if msg.ToolCallID == id {
			return i
		}
	}
	return -1
}

func HandleToolCall(s *session.Session) error {
	// DEBUG:
	{
		logger.Debugf("HandleToolCall Start >>>>>>>>>>>>>>>>\n")
		logger.Debugf("[HandleToolCall] message (before):\n")
		for _, message := range s.Runtime.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
	}

	idChanMap := map[llm.ToolCall]chan tool.ToolResult{}
	runtimeMu := &sync.RWMutex{}


	for _, tc := range s.Runtime.Response.ToolCalls() {
		// 已经执行过
		if findToolCallID(tc.ID, s.ToolState.ToolResults) != -1 {
			continue
		}

		logger.Infof("exec tool: %v\n", tc.Function.Name)
		resCh := make(chan tool.ToolResult)
		idChanMap[tc] = resCh

		t, ok := s.ToolState.ToolMap[tc.Function.Name]
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) not exists", tc.Function.Name, tc.ID,
			)
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					resCh <- tool.ToolResult{
						Result: fmt.Sprintf(
							"exec tool %v failed: %v", tc.Function.Name, r,
						),
					}
				}
			}()
			res := tool.ExecInternalTool(s.Meta, s.Runtime, runtimeMu, tc.ID, t, tc.Function.Arguments)
			resCh <- res
		}()

	}

	llmToolMsgs := make([]llm.Message, 0, len(idChanMap))
	toolMsgs := make([]tool.ToolResult, 0, len(idChanMap))

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
		appendInput := s.Runtime.MessageQueue.GetTypedInput(userinput.InputTypePrompt)
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
	s.ToolState.ToolResults = append(s.ToolState.ToolResults, toolMsgs...)

	// DEBUG:
	{
		// fmt.Printf("[HandleToolCall] result: %v\n", ctx.Response.Content())
		logger.Debugf(
			"[HandleToolCall] HasToolCall: %v\n",
			s.Runtime.Response.HasToolCalls(),
		)
		logger.Debugf("[HandleToolCall] message (after):\n")
		for _, message := range s.Runtime.Messages {
			logger.Debugf("\tmessage: %+v\n", message)
		}
		logger.Debugf("HandleToolCall Done <<<<<<<<<<<<<<<<\n\n")
	}

	return nil
}
