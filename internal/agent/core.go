package agent

import "context"
import "errors"
import "fmt"
import "strings"

import "myagent/internal/config"
import agentctx "myagent/internal/context"
import "myagent/internal/handler"
import "myagent/internal/runtime"
import "myagent/internal/tool"
import "myagent/pkg/llm"
import "myagent/pkg/logger"

var EmptyQueryErr = fmt.Errorf("empty user query")

type Agent struct {
	Config config.ProjectConfig
	State  runtime.AgentState
}

func (a *Agent) InitAgentState(
	ctx context.Context,
	query string,
	input chan string,
	output chan runtime.AgentResponse,
) error {
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
	systemPrompt, err := agentctx.GetSystemPrompt(a.Config)
	if err != nil {
		panic(err)
	}

	a.State = runtime.AgentState{
		Ctx: ctx,
		AgentConfig: runtime.AgentConfig{
			AgentMode:   runtime.AgentModePlan,
			ToolAskMode: runtime.ToolAskModeAuto,
		},
		SystemPrompt: systemPrompt,
		UserQuery:    query,
		InputChan:    input,
		OutputChan:   output,
		LLMClient:    llmClient,
		ToolMap:      registerTools(),
	}

	return nil
}

// Exec 执行一次 agent 流程
func (a *Agent) Exec(
	ctx context.Context,
	query string,
	input chan string,
	output chan runtime.AgentResponse,
) {
	a.InitAgentState(ctx, query, input, output)

	logger.Debugf("[agentLoopCore] query: %v\n", query)

	agentHandleChan := make(
		chan struct {
			Resp *runtime.AgentResponse
			Err  error
		},
		1,
	)
	go func() {
		data := struct {
			Resp *runtime.AgentResponse
			Err  error
		}{}
		data.Resp, data.Err = a.agentHandler(ctx)
		agentHandleChan <- data
	}()

	res := struct {
		Resp *runtime.AgentResponse
		Err  error
	}{}
	select {
	case <-ctx.Done():
		return
	case res = <-agentHandleChan:
	}

	if errors.Is(res.Err, EmptyQueryErr) {
		return
	} else if res.Err != nil {
		output <- runtime.AgentResponse{
			RespType: runtime.AgentRespTypeError,
			Err:      fmt.Errorf("execute agent loop failed: %v", res.Err),
		}
	} else {
		logger.Debugf("agent response: %v\n", *res.Resp)
		output <- *res.Resp
	}
}

// registerTools 返回所有工具的注册表
func registerTools() map[string]tool.Tool {
	// TODO: write-todo, write-memory
	list_dir := &tool.ListDirTool{}

	return map[string]tool.Tool{
		list_dir.Name(): list_dir,
	}
}

func (a *Agent) agentHandler(ctx context.Context) (*runtime.AgentResponse, error) {
	var err error

	if err = ctx.Err(); err != nil {
		return nil, err
	}

	if a.State.UserQuery == "" { // empty query
		return nil, EmptyQueryErr
	} else if strings.HasPrefix(a.State.UserQuery, "/") { // command
		res, err := handler.HandleAgentCommand(ctx, &a.State)
		if err != nil {
			return nil, err
		}
		return &runtime.AgentResponse{
			RespType:  runtime.AgentRespTypeCmd,
			CmdResult: res,
		}, nil
	}

	if a.State.Response.HasToolCalls() { // tool call
		logger.Debugf("handle tool call: %v\n", a.State.UserQuery)
		err = handler.HandleToolCall(ctx, &a.State)

	} else if a.State.UserQuery != "" { // normal query
		logger.Debugf("handle query: %v\n", a.State.UserQuery)
		err = handler.HandleQuery(ctx, a.Config, &a.State)

	} else { // invalid query
		logger.Errorf("invalid LoopContext\n")
		return nil, fmt.Errorf("invalid LoopContext")
	}

	if err != nil {
		return nil, err
	}

	// TODO: append query from ctx.InputChan

	// TODO: update usage && context size

	// TODO: async update storage

	if a.State.Response != nil &&
		!a.State.Response.HasToolCalls() &&
		a.State.Response.Content() != "" {
		return &runtime.AgentResponse{
			RespType:    runtime.AgentRespTypeLLM,
			LLMResponse: a.State.Response,
		}, nil
	}

	return a.agentHandler(ctx)
}
