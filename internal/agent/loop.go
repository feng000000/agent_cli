package agent

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"myagent/internal/config"
	"myagent/internal/handler"
	"myagent/internal/runtime"
	"myagent/internal/tool"
	"myagent/pkg/llm"
	"myagent/pkg/logger"
)

type Agent struct {
	Config config.AgentConfig
}

func (a *Agent) Exec(input io.Reader, output io.Writer) error {
	ich := make(chan string)
	och := make(chan runtime.AgentResponse)
	go readMessage(input, ich)

	for {
		fmt.Fprint(output, "User>")

		query := <-ich
		go a.runAgent(query, ich, och)

		content := <-och

		switch content.RespType {
		case runtime.AgentRespTypeLLM:
			fmt.Fprintf(output, "Agent>%s\n", content.LLMResponse.Content())
		case runtime.AgentRespTypeError:
			fmt.Fprintf(output, ">>>!Error: %s\n", content.Err.Error())
		case runtime.AgentRespTypeMiddleMsg:
			fmt.Fprintf(output, "| > %s\n", content.MiddleMessage)
		case runtime.AgentRespTypeCMD:
			fmt.Fprintf(output, "|| tools exec result: %s\n", content.CMDResult)
		}
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
			logger.Errorf("read input error: %v\n", err)
		}
	}
}

// runAgent 执行一次 agent 流程
func (a *Agent) runAgent(query string, input chan string, output chan runtime.AgentResponse) {
	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey: a.Config.LLM.APIKey,
			BaseURL: a.Config.LLM.BaseURL,
			Model: a.Config.LLM.Model,
		},
	)
	if err != nil {
		panic(err)
	}
	loopContext := runtime.LoopContext{
		AgentConfig: runtime.AgentConfig{
			AgentMode:   runtime.AgentModePlan,
			ToolAskMode: runtime.ToolAskModeAuto,
		},

		UserQuery:  query,
		InputChan:  input,
		OutputChan: output,
		LLMClient:  llmClient,
		ToolMap:    registerTools(),
	}

	logger.Debugf("[agentLoopCore] query: %v\n", query)

	resp, err := agentHandler(&loopContext)

	if err != nil {
		output <- runtime.AgentResponse{
			RespType: runtime.AgentRespTypeMiddleMsg,
			Err: fmt.Errorf("execute agent loop failed: %v", err),
		}
	} else {
		output <- *resp
	}
}

// registerTools 返回所有工具的注册表
func registerTools() map[string]tool.Tool {
	list_dir := &tool.ListDirTool{}

	return map[string]tool.Tool{
		list_dir.Name(): list_dir,
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
func agentHandler(ctx *runtime.LoopContext) (*runtime.AgentResponse, error) {
	var err error

	if strings.HasPrefix(ctx.UserQuery, "/") { // command
		res, err := handler.HandleCommand(ctx)
		if err != nil {
			return nil, err
		}
		return &runtime.AgentResponse{
			RespType:    runtime.AgentRespTypeCMD,
			CMDResult: res,
		}, nil

	} else if ctx.UserQuery != "" { // normal query
		err = handler.HandleQuery(ctx)
	} else if ctx.Response.HasToolCalls() { // tool call
		err = handler.HandleToolCall(ctx)
	} else {
		// fmt.Printf("invalid LoopContext: %v\n", ctx)
		logger.Errorf("invalid LoopContext\n")
		return nil, fmt.Errorf("invalid LoopContext")
	}

	// TODO: update history, memory

	if err != nil {
		return nil, err
	}

	// TODO: append query from ctx.InputChan

	if ctx.Response != nil &&
		!ctx.Response.HasToolCalls() &&
		ctx.Response.Content() != "" {
		return &runtime.AgentResponse{
			RespType:    runtime.AgentRespTypeLLM,
			LLMResponse: ctx.Response,
		}, nil
	}

	return agentHandler(ctx)
}
