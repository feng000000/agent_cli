package agent

import (
	"bytes"
	"context"
	"errors"
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

var EmptyQueryErr = fmt.Errorf("empty user query")

type Agent struct {
	Config config.AgentConfig
}

func (a *Agent) StartSimpleUI(input io.Reader, output io.Writer) error {
	ich := make(chan string)
	och := make(chan runtime.AgentResponse)
	go readMessage(input, ich)
	go outputMessage(output, och)

	for {
		logger.Debugf("test\n")
		fmt.Fprint(output, "User>")

		query := <-ich

		ctx, cancel := context.WithCancel(context.Background())
		clientState := runtime.ClientState{CancelFunc: cancel}

		if strings.HasPrefix(query, "/") {
			res, err := handler.HandleClientCommand(&clientState, query)
			if err == nil {
				logger.Debugf("exec client command: %v %v", res, err)
				fmt.Fprintf(output, "|🔧: %s\n", res)
				continue
			} else if !errors.Is(err, handler.SkipHandleCommand) {
				fmt.Fprintf(output, ">>>❗Error: %s\n", err.Error())
			}
			// skip
		}

		go a.Exec(ctx, query, ich, och)

		// content := <-och
		// switch content.RespType {
		// case runtime.AgentRespTypeLLM:
		// 	fmt.Fprintf(output, "Agent✨> %s\n", content.LLMResponse.Content())
		// case runtime.AgentRespTypeError:
		// 	fmt.Fprintf(output, ">>>❗Error: %s\n", content.Err.Error())
		// case runtime.AgentRespTypeMiddleMsg:
		// 	fmt.Fprintf(output, "|🤔> %s\n", content.MiddleMessage)
		// case runtime.AgentRespTypeCmd:
		// 	fmt.Fprintf(output, "|☁️🔧: %s\n", content.CmdResult)
		// }
	}

}

// TODO: 终端输入\r分割
// TODO: api
// readMessage 读取消息 r -> ch
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

// outputMessage 打印信息 ch -> o
func outputMessage(o io.Writer, ch chan runtime.AgentResponse) {
	for content := range ch {
		switch content.RespType {
		case runtime.AgentRespTypeLLM:
			fmt.Fprintf(o, "Agent✨> %s\n", content.LLMResponse.Content())
		case runtime.AgentRespTypeError:
			fmt.Fprintf(o, ">>>❗Error: %s\n", content.Err.Error())
		case runtime.AgentRespTypeMiddleMsg:
			fmt.Fprintf(o, "|🤔> %s\n", content.MiddleMessage)
		case runtime.AgentRespTypeCmd:
			fmt.Fprintf(o, "|☁️🔧: %s\n", content.CmdResult)
		}
	}
}

// Exec 执行一次 agent 流程
func (a *Agent) Exec(
	ctx context.Context,
	query string,
	input chan string,
	output chan runtime.AgentResponse,
) {
	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  a.Config.LLM.APIKey,
			BaseURL: a.Config.LLM.BaseURL,
			Model:   a.Config.LLM.Model,
		},
	)
	if err != nil {
		panic(err)
	}
	state := runtime.AgentState{
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

	resp, err := agentHandler(ctx, &state)

	if errors.Is(err, EmptyQueryErr) {
		return
	} else if err != nil {
		output <- runtime.AgentResponse{
			RespType: runtime.AgentRespTypeError,
			Err:      fmt.Errorf("execute agent loop failed: %v", err),
		}
	} else {
		logger.Debugf("agent response: %v\n", resp)
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
func agentHandler(
	ctx context.Context, state *runtime.AgentState,
) (*runtime.AgentResponse, error) {
	var err error

	if state.UserQuery == "" { // empty query
		return nil, EmptyQueryErr
	} else if strings.HasPrefix(state.UserQuery, "/") { // command
		res, err := handler.HandleAgentCommand(state)
		if err != nil {
			return nil, err
		}
		return &runtime.AgentResponse{
			RespType:  runtime.AgentRespTypeCmd,
			CmdResult: res,
		}, nil
	} else if state.UserQuery != "" { // normal query
		logger.Debugf("handle query: %v\n", state.UserQuery)
		err = handler.HandleQuery(state)
	} else if state.Response.HasToolCalls() { // tool call
		logger.Debugf("handle tool call: %v\n", state.UserQuery)
		err = handler.HandleToolCall(state)
	} else { // invalid query
		// fmt.Printf("invalid LoopContext: %v\n", ctx)
		logger.Errorf("invalid LoopContext\n")
		return nil, fmt.Errorf("invalid LoopContext")
	}

	if err != nil {
		return nil, err
	}

	// TODO: append query from ctx.InputChan

	if state.Response != nil &&
		!state.Response.HasToolCalls() &&
		state.Response.Content() != "" {
		return &runtime.AgentResponse{
			RespType:    runtime.AgentRespTypeLLM,
			LLMResponse: state.Response,
		}, nil
	}

	return agentHandler(ctx, state)
}
