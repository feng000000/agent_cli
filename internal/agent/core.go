package agent

import "fmt"

import "agentcli/internal/handler"
import "agentcli/internal/session"
import "agentcli/internal/session/userinput"
import "agentcli/internal/session/response"
import "agentcli/pkg/logger"

// Exec 调用 agent
func Exec(
	a *Agent,
	firstInput *userinput.UserInput,
) {
	if a.Session == nil {
		logger.Panicf("AgentState not initialized")
	}

	// run agent
	if res, err := a.runAgent(firstInput); err != nil {
		a.Session.Runtime.Emit(
			&response.AgentResponse{
				RespType: response.AgentRespTypeError,
				Err:      fmt.Errorf("execute agent loop failed: %v", err),
			},
		)
	} else { // got a response
		logger.Debugf("agent response: %v\n", *res)
		a.Session.Runtime.Emit(res)
	}

}

func NewAgent(
	sessionID string,
	mq *userinput.MessageQueue,
	output chan *response.AgentResponse,
) (*Agent, error) {
	if mq == nil {
		mq = userinput.NewMessageQueue()
	}
	if output == nil {
		output = make(chan *response.AgentResponse, 65536)
	}

	var s *session.Session
	var err error
	if sessionID != "" {
		s, err = session.LoadSession(sessionID, mq, output)
	} else {
		s, err = session.NewSession(mq, output)
	}

	if err != nil {
		return nil, err
	}

	return &Agent{Session: s}, nil

}

type Agent struct {
	Session *session.Session
}

// runAgent 调用 agentHandler, 同时监听 ctx 的取消信号
func (a *Agent) runAgent(firstInput *userinput.UserInput) (*response.AgentResponse, error) {
	agentHandleChan := make(
		chan struct {
			Resp *response.AgentResponse
			Err  error
		},
		1,
	)
	go func() {
		// empty query or command
		{
			if firstInput == nil || len(firstInput.Content) == 0 { // empty query
				return
			} else if firstInput.Type() == userinput.InputTypeCommand { // command
				res, err := handler.HandleAgentCommand(
					a.Session.Runtime.Ctx,
					a.Session,
					string(firstInput.Content),
				)
				if err != nil {
					return
				}
				a.Session.Runtime.Emit(
					&response.AgentResponse{
						RespType:  response.AgentRespTypeCmd,
						CmdResult: res,
					},
				)
				return
			}
		}

		data := struct {
			Resp *response.AgentResponse
			Err  error
		}{}
		data.Resp, data.Err = a.agentHandler(string(firstInput.Content))
		agentHandleChan <- data
	}()

	res := struct {
		Resp *response.AgentResponse
		Err  error
	}{}
	select {
	case <-a.Session.Runtime.Ctx.Done():
		return nil, fmt.Errorf("agent handler canceled")
	case res = <-agentHandleChan:
	}

	return res.Resp, res.Err
}

// agentHandler 运行一次 agent 直到发生错误或产生最终结果
func (a *Agent) agentHandler(query string) (*response.AgentResponse, error) {
	defer func() {
		// compressor := runtime.LLMCompressor{}
		// compressor := runtime.TruncateCompressor{TopK: 10}
		compressor := session.HybridCompressor{TopK: 10}
		if a.Session.Runtime.ContextSize >= a.Session.Meta.MaxTokensToCompress {
			logger.Infof(
				"Context size over limit (%v), exec Compress\n",
				a.Session.Runtime.ContextSize,
			)
			compressor.Compress(a.Session.Runtime.Ctx, a.Session)
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
				a.Session.Runtime.Response != nil &&
				a.Session.Runtime.Response.HasToolCalls() { // tool call

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

		if a.Session.Runtime.Response != nil && !a.Session.Runtime.Response.HasToolCalls() {
			if a.Session.Runtime.Response.Content() != "" {
				return &response.AgentResponse{
					RespType:    response.AgentRespTypeLLM,
					LLMResponse: a.Session.Runtime.Response,
				}, nil
			}
			return nil, fmt.Errorf("LLM empty response")
		}

		first = false
	}
}
