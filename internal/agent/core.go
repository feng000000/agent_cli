package agent

import "fmt"

import "myagent/internal/handler"
import "myagent/internal/runtime"
import "myagent/pkg/logger"

// Exec 调用 agent
func Exec(
	a *Agent,
	firstInput *runtime.UserInput,
) {
	if a.Session == nil {
		logger.Panicf("AgentState not initialized")
	}

	// run agent
	if res, err := a.runAgent(firstInput); err != nil {
		a.Session.OutputChan <- &runtime.AgentResponse{
			RespType: runtime.AgentRespTypeError,
			Err:      fmt.Errorf("execute agent loop failed: %v", err),
		}
	} else { // got a response
		logger.Debugf("agent response: %v\n", *res)
		a.Session.OutputChan <- res
	}

}

func NewAgent(
	sessionID string,
	mq *runtime.MessageQueue,
	output chan *runtime.AgentResponse,
) (*Agent, error) {
	if mq == nil {
		mq = runtime.NewMessageQueue()
	}
	if output == nil {
		output = make(chan *runtime.AgentResponse, 65536)
	}

	var session *runtime.Session
	var err error
	if sessionID != "" {
		session, err = runtime.LoadSession(sessionID, mq, output)
	} else {
		session, err = runtime.NewSession(mq, output)
	}

	if err != nil {
		return nil, err
	}

	return &Agent{Session: session}, nil

}

type Agent struct {
	Session *runtime.Session
}

// runAgent 调用 agentHandler, 同时监听 ctx 的取消信号
func (a *Agent) runAgent(firstInput *runtime.UserInput) (*runtime.AgentResponse, error) {
	agentHandleChan := make(
		chan struct {
			Resp *runtime.AgentResponse
			Err  error
		},
		1,
	)
	go func() {
		// empty query or command
		{
			if firstInput == nil || len(firstInput.Content) == 0 { // empty query
				return
			} else if firstInput.Type() == runtime.InputTypeCommand { // command
				res, err := handler.HandleAgentCommand(
					a.Session.Ctx,
					a.Session,
					string(firstInput.Content),
				)
				if err != nil {
					return
				}
				a.Session.OutputChan <- &runtime.AgentResponse{
					RespType:  runtime.AgentRespTypeCmd,
					CmdResult: res,
				}
				return
			}
		}

		data := struct {
			Resp *runtime.AgentResponse
			Err  error
		}{}
		data.Resp, data.Err = a.agentHandler(string(firstInput.Content))
		agentHandleChan <- data
	}()

	res := struct {
		Resp *runtime.AgentResponse
		Err  error
	}{}
	select {
	case <-a.Session.Ctx.Done():
		return nil, fmt.Errorf("agent handler canceled")
	case res = <-agentHandleChan:
	}

	return res.Resp, res.Err
}

// agentHandler 运行一次 agent 直到发生错误或产生最终结果
func (a *Agent) agentHandler(query string) (*runtime.AgentResponse, error) {
	defer func() {
		// compressor := runtime.LLMCompressor{}
		// compressor := runtime.TruncateCompressor{TopK: 10}
		compressor := runtime.HybridCompressor{TopK: 10}
		if a.Session.ContextSize >= a.Session.Meta.MaxTokensToCompress {
			logger.Infof(
				"Context size over limit (%v), exec Compress\n",
				a.Session.ContextSize,
			)
			compressor.Compress(a.Session.Ctx, a.Session)
		}
	}()

	first := true
	for {
		if a.Session == nil {
			return nil, fmt.Errorf("AgentState is nil")
		}

		// route
		{
			var err error

			if !first &&
				a.Session.Response != nil &&
				a.Session.Response.HasToolCalls() { // tool call

				logger.Debugf("handle tool call: %v\n", query)
				err = handler.HandleToolCall(a.Session)

			} else if query != "" { // normal query

				logger.Debugf("handle query: %v\n", query)
				err = handler.HandleQuery(query, a.Session)

			} else {
				err = fmt.Errorf("invalid context (empty tool call and query)")
			}

			if err != nil {
				return nil, err
			}
		}
		// // update usage (context size == )
		// if a.Session != nil && a.Session.Response != nil {
		// 	a.Session.Usage = a.Session.Usage.Append(&a.Session.Response.Usage)

		// 	// update context size
		// 	a.Session.ContextSize = a.Session.Response.Usage.PromptTokens +
		// 		a.Session.Response.Usage.CompletionTokens
		// }

		a.Session.Save()

		if a.Session.Response != nil && !a.Session.Response.HasToolCalls() {
			if a.Session.Response.Content() != "" {
				return &runtime.AgentResponse{
					RespType:    runtime.AgentRespTypeLLM,
					LLMResponse: a.Session.Response,
				}, nil
			}
			return nil, fmt.Errorf("LLM empty response")
		}

		first = false
	}
}
