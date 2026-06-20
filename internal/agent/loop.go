package agent

import (
	"bytes"
	"fmt"
	"io"
	"myagent/internal/config"
	"myagent/internal/handler"
	"myagent/internal/runtime"
	"myagent/internal/tools"
	"myagent/pkg/llm"
	"strings"
)

type Agent struct {
	Config config.AgentConfig
}

func (a *Agent) Run(input io.Reader, output io.Writer) error {
	ich := make(chan string)
	och := make(chan string)
	go readMessage(input, ich)
	go agentLoopCore(ich, och)

	for {
		fmt.Fprint(output, "User>")

		// TODO: och 中数据结构需要更复杂, 区分 思考/工具输出/中间输出/最终输出
		content := <-och

		fmt.Fprintf(output, "Agent>%s\n", content)
	}

}

// TODO: 终端输入\r分割
// TODO: api
// readMessage
func readMessage(r io.Reader, ch chan string) {
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
				ch <- string(data)
				data = data[:0]
			}
		}

		if err != nil {
			fmt.Printf("read input error: %v\n", err)
		}
	}
}

func RegisterTools() map[string]tools.Tool {
	list_dir := &tools.ListDirTool{}

	return map[string]tools.Tool{
		list_dir.Name(): list_dir,
	}
}

// TODO: channel 往返结构体需要更多信息
func agentLoopCore(input_ch chan string, output chan string) {
	llmClient, err := llm.NewDeepSeekClient(llm.DeepSeekConfig{})
	if err != nil {
		panic(err)
	}
	loopContext := runtime.LoopContext{
		OutputChan: output,
		LLMClient:  llmClient,
		ToolMap:    RegisterTools(),
	}

	loopRound := 1

	for {
		data := <-input_ch

		fmt.Printf("[agentLoopCore] round %v\n", loopRound)
		fmt.Printf("[agentLoopCore] input data: %v\n", string(data))

		loopContext.Query = string(data)
		resp, err := agentHandler(&loopContext)

		loopRound += 1

		if err != nil {
			output <- fmt.Sprintf("execute agent loop failed: %v", err)
		} else {
			output <- resp.Content()
		}
	}

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
func agentHandler(ctx *runtime.LoopContext) (*llm.ChatResponse, error) {
	var err error

	if strings.HasPrefix(ctx.Query, "/") { // command
		err = handler.HandleCommand(ctx)
	} else if ctx.Query != "" { // normal query
		err = handler.HandleQuery(ctx)
	} else if ctx.Response.HasToolCalls() { // tool call
		err = handler.HandleToolCall(ctx)
	} else {
		// fmt.Printf("invalid LoopContext: %v\n", ctx)
		fmt.Printf("invalid LoopContext\n")
		return nil, fmt.Errorf("invalid LoopContext")
	}

	// TODO: update history, memory

	if err != nil {
		return nil, err
	}


	if ctx.Response != nil &&
		!ctx.Response.HasToolCalls() &&
		ctx.Response.Content() != "" {
		return ctx.Response, nil
	}

	return agentHandler(ctx)
}
