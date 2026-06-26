package agent

import (
	"context"
	"errors"
	"fmt"
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

	a.State = runtime.AgentState{
		Ctx: ctx,
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

	resp, err := a.agentHandler()

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
// TODO: write-todo, write-memory
func registerTools() map[string]tool.Tool {
	list_dir := &tool.ListDirTool{}

	return map[string]tool.Tool{
		list_dir.Name(): list_dir,
	}
}

func (a *Agent) agentHandler() (*runtime.AgentResponse, error) {
	var err error

	if a.State.UserQuery == "" { // empty query
		return nil, EmptyQueryErr
	} else if strings.HasPrefix(a.State.UserQuery, "/") { // command
		res, err := handler.HandleAgentCommand(&a.State)
		if err != nil {
			return nil, err
		}
		return &runtime.AgentResponse{
			RespType:  runtime.AgentRespTypeCmd,
			CmdResult: res,
		}, nil
	} else if a.State.UserQuery != "" { // normal query
		logger.Debugf("handle query: %v\n", a.State.UserQuery)
		err = handler.HandleQuery(a.Config, &a.State)
	} else if a.State.Response.HasToolCalls() { // tool call
		logger.Debugf("handle tool call: %v\n", a.State.UserQuery)
		err = handler.HandleToolCall(&a.State)
	} else { // invalid query
		logger.Errorf("invalid LoopContext\n")
		return nil, fmt.Errorf("invalid LoopContext")
	}

	if err != nil {
		return nil, err
	}

	// TODO: append query from ctx.InputChan

	if a.State.Response != nil &&
		!a.State.Response.HasToolCalls() &&
		a.State.Response.Content() != "" {
		return &runtime.AgentResponse{
			RespType:    runtime.AgentRespTypeLLM,
			LLMResponse: a.State.Response,
		}, nil
	}

	return a.agentHandler()
}
