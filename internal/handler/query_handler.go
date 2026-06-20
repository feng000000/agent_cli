package handler

import "context"
import "fmt"
import "time"

import "myagent/internal/runtime"
import "myagent/pkg/llm"

func HandleQuery(ctx *runtime.LoopContext) error {
	fmt.Println("Handle query")

	fmt.Printf("HandleQuery Start >>>>>>>>>>>>>>>>\n\n")
	fmt.Printf("[HandleQuery] query: %v\n", ctx.Query)
	fmt.Println("[HandleQuery] message (before):")
	for _, message := range ctx.MessageParams {
		fmt.Printf("\tmessage: %+v\n", message)
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
		fmt.Println("[CALL LLM] content:")
		fmt.Println(resp.Content())
	} else {
		fmt.Println("[CALL LLM] tool calls:")
		for _, tc := range resp.ToolCalls() {
			fmt.Printf(
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


	fmt.Println("[HandleQuery] message (after):")
	for _, message := range ctx.MessageParams {
		fmt.Printf("\tmessage: %+v\n", message)
	}
	fmt.Printf("HandleQuery Done <<<<<<<<<<<<<<<<\n\n")

	lastMsg := ctx.MessageParams[len(ctx.MessageParams) - 1]
	if lastMsg.ReasoningContent != "" {
		ctx.OutputChan <- lastMsg.ReasoningContent
	}

	return nil
}
