package agent

import (
	"bytes"
	"fmt"
	"io"
	"myagent/config"
	"strings"
)

func debug(values ...any) {
	fmt.Printf("[DEBUG]: %v\n", values...)
}

type Agent struct {
	Config config.AgentConfig
}

// TODO: implement
type Tool struct {
}

type ToolCall struct {
	Params []string
	Tool   Tool
	Result string
}

type LoopContext struct {
	query    string
	history  []string
	toolCall []ToolCall
}

type LoopResponse struct {
	answer   string
	toolCall []ToolCall
}

// TODO: implement
// new query / tool use results:
// - check command
// - render system prompt template
// - inject context
//   - history
//   - tools, skills
//   - tool results
//
// - call LLM
//
// - parse response
//   - tool call: return agentHandler("", tool_call_list)
//   - response:
//
// - update history, memory
func agentHandler(ctx LoopContext) (LoopResponse, error) {
	response := LoopResponse{}

	if strings.HasPrefix(ctx.query, "/") { // command
		// TODO: handle command
	} else if ctx.query != "" { // normal query
		debug("mock query response")

		ctx.history = append(ctx.history, ctx.query)
		ctx.query = ""
		ctx.toolCall = []ToolCall{
			{
				Params: []string{"mock param1", "mock param2"},
				Tool:   Tool{},
			},
		}
	} else if len(ctx.toolCall) != 0 { // tool call
		debug("mock exec tool: %v", ctx.toolCall)

		for idx, tc := range ctx.toolCall {
			tc.Result = fmt.Sprintf("mock tool result %v", idx)
		}

		response.answer = fmt.Sprintf("exec tool: %v\n", ctx.toolCall)
	} else {
		debug("empty LoopMessage")
		return response, fmt.Errorf("empty LoopMessage: %v", ctx)
	}

	// TODO: update history, memory

	if response.answer != "" {
		return response, nil
	}

	return agentHandler(ctx)
}

// TODO: 终端输入 \r
// readMessage
func readMessage(r io.Reader) (string, error) {
	delimiter := []byte("\n")
	// fmt.Printf("| delimiter: (%v)\n", delimiter)

	var data []byte
	var buf [1]byte
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			data = append(data, buf[0])

			if bytes.HasSuffix(data, delimiter) {
				data = data[:len(data)-len(delimiter)]
				return string(data), nil
			}
		}

		if err != nil {
			return "", err
		}
	}
}

func (a *Agent) AgentLoop(input io.Reader, output io.Writer) error {
	for {
		output.Write([]byte("User>"))
		data, err := readMessage(input)
		if err != nil {
			fmt.Println("Error: read failed")
			fmt.Fprintf(output, "Read from input failed: %v", err)
			continue
		}

		resp, err := agentHandler(LoopContext{query: string(data)})
		if err != nil {
			return err
		}

		fmt.Fprintf(output, "Agent>%s\n", resp)
	}

}
