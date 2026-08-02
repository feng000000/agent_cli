package main

import "fmt"
import "log"
import "context"
import "agentcli/pkg/llm"
import "agentcli/internal/tool"
import "agentcli/internal/tool/toolimpl"

type jsonO map[string]any

func main() {
	client, err := llm.NewDeepSeekClient(llm.DeepSeekConfig{})
	if err != nil {
		log.Fatal(err)
	}

	messages := []llm.Message{
		llm.UserMessage("./ 目录下有哪些文件"),
	}

	listDirTool := toolimpl.ListDirTool{}

	tools := []llm.Tool{
		*listDirTool.Definition(),
		// llm.FunctionToolDefinition(
		// 	"get_weather",
		// 	"Get weather for a location.",
		// 	jsonO{
		// 		"type": "object",
		// 		"properties": jsonO{
		// 			"location": jsonO{
		// 				"type":        "string",
		// 				"description": "City name, e.g. Hangzhou",
		// 			},
		// 		},
		// 		"required": []string{"location"},
		// 	},
		// ),
	}

	resp, err := client.Chat(context.Background(), llm.ChatRequest{
		Messages:   messages,
		Tools:      tools,
		ToolChoice: llm.ToolChoiceAuto,
	})
	if err != nil {
		log.Fatal(err)
	}

	if !resp.HasToolCalls() {
		fmt.Println(resp.Content())
		return
	}

	fmt.Println("tool calls:")
	for _, tc := range resp.ToolCalls() {
		fmt.Printf("id: %s\n", tc.ID)
		fmt.Printf("type: %s\n", tc.Type)
		fmt.Printf("function: %s\n", tc.Function.Name)
		fmt.Printf("arguments: %s\n", tc.Function.Arguments)
	}

	assistantMsg, _ := resp.Message()
	messages = append(messages, assistantMsg)
	for _, tc := range resp.ToolCalls() {
		fmt.Printf(
			"tool call: %s(%s)\n",
			tc.Function.Name,
			tc.Function.Arguments,
		)
		if tc.Function.Name == listDirTool.Name() {

			// 这里替换成真实工具执行逻辑
			// toolResult := "杭州当前气温 28°C，多云。"
			toolResult, err := listDirTool.ExecuteImpl(&tool.ToolContext{}, tc.Function.Arguments)
			if err != nil {
				fmt.Printf("tool exec failed: %v\n", err)
				messages = append(
					messages,
					llm.ToolResultMessage(tc.ID, fmt.Sprintf("tool exec failed: %v", err)),
				)
			}

			fmt.Printf("tool result: %v\n", toolResult)

			messages = append(
				messages,
				llm.ToolResultMessage(tc.ID, toolResult),
			)
		}
	}

	finalResp, err := client.Chat(context.Background(), llm.ChatRequest{
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(finalResp.Content())
	fmt.Println("tool calls:")
	for _, tc := range resp.ToolCalls() {
		fmt.Printf("id: %s\n", tc.ID)
		fmt.Printf("type: %s\n", tc.Type)
		fmt.Printf("function: %s\n", tc.Function.Name)
		fmt.Printf("arguments: %s\n", tc.Function.Arguments)
	}
}
