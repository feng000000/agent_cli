package handler

import "fmt"
import "time"
import "context"

import "myagent/internal/runtime"
import "myagent/pkg/llm"

func HandleToolCall(ctx *runtime.LoopContext) error {
	idChanMap := map[llm.ToolCall]chan string{}

	fmt.Printf("HandleToolCall Start >>>>>>>>>>>>>>>>\n")
	fmt.Println("[HandleToolCall] message (before):")
	for _, message := range ctx.MessageParams {
		fmt.Printf("\tmessage: %+v\n", message)
	}

	for _, tc := range ctx.Response.ToolCalls() {
		fmt.Printf("exec tool: %v\n", tc.Function.Name)
		resCh := make(chan string)
		idChanMap[tc] = resCh

		tool, ok := ctx.ToolMap[tc.Function.Name]
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) not exists", tc.Function.Name, tc.ID,
			)
		}

		go tool.Execute(tc.Function.Arguments, resCh)

	}

	for tc, ch := range idChanMap {
		res, ok := <-ch
		if !ok {
			return fmt.Errorf(
				"tool %v(%v) execute failed", tc.Function.Name, tc.ID,
			)
		}

		ctx.MessageParams = append(
			ctx.MessageParams,
			llm.ToolResultMessage(tc.ID, res),
		)
	}

	timeoutCtx, cancel := context.WithTimeout(
		context.Background(), 60*time.Second,
	)
	defer cancel()

	resp, err := ctx.LLMClient.Chat(
		timeoutCtx,
		llm.ChatRequest{
			Messages:   ctx.MessageParams,
			Tools:      ctx.ToolList(),
		},
	)
	if err != nil {
		return err
	}

	ctx.Response = resp
	assistantMsg, ok := resp.Message()
	if ok {
		ctx.MessageParams = append(ctx.MessageParams, assistantMsg)
	}

	// fmt.Printf("[HandleToolCall] result: %v\n", ctx.Response.Content())
	fmt.Printf("[HandleToolCall] HasToolCall: %v\n", ctx.Response.HasToolCalls())


	fmt.Println("[HandleToolCall] message (after):")
	for _, message := range ctx.MessageParams {
		fmt.Printf("\tmessage: %+v\n", message)
	}
	fmt.Printf("HandleToolCall Done <<<<<<<<<<<<<<<<\n\n")

	return nil
}
