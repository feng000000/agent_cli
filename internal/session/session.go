package session

import "os"
import "path"
import "time"
import "encoding/json"

import "agentcli/internal/session/userinput"
import "agentcli/internal/session/meta"
import "agentcli/internal/session/response"
import "agentcli/internal/session/runtime"
import "agentcli/internal/session/toolstate"
import "agentcli/pkg/llm"


// TODO:  ClientState: client 分开实现
type ClientState struct {
	CancelFunc func()
}



const SessionDirBase = "./.myagent/sessions"

type Session struct {
	Meta *meta.Meta

	ToolState *toolstate.ToolState

	Runtime *runtime.Runtime
}

func (s *Session) ForceUpdateContextInfo(
	messages []llm.Message,
	usage *llm.Usage,
	contextSize int,
) {
	s.Runtime.Messages = messages
	s.Runtime.Usage = usage
	s.Runtime.ContextSize = contextSize
}


// CallLLM 封装请求 LLM 和更新上下文的操作
func (s *Session) CallLLM(newMsgs ...llm.Message) error {
	// send event
	var gotRespCh chan bool = make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case res := <-gotRespCh:
				if res {
					s.Runtime.Emit(
						&response.AgentResponse{
							RespType:      response.AgentRespTypeMiddleMsg,
							MiddleMessage: "got llm resp",
						},
					)
				} else {
					s.Runtime.Emit(
						&response.AgentResponse{
							RespType:      response.AgentRespTypeMiddleMsg,
							MiddleMessage: "got llm resp failed",
						},
					)
				}
				return
			case <-ticker.C:
				s.Runtime.Emit(
					&response.AgentResponse{
						RespType:      response.AgentRespTypeMiddleMsg,
						MiddleMessage: "waiting",
					},
				)
			}
		}
	}()

	sysPrompt, err := s.SystemPrompt()
	if err != nil {
		return err
	}
	msgs := append(
		[]llm.Message{llm.SystemMessage(sysPrompt)},
		s.Runtime.Messages...,
	)
	msgs = append(msgs, newMsgs...)
	resp, err := s.Runtime.RawLLMClient().Chat(
		s.Runtime.Ctx,
		llm.ChatRequest{
			Messages: msgs,
			Tools:    s.ToolList(),
		},
	)
	// got resp signal
	if err != nil {
		gotRespCh <- false
		return err
	}
	gotRespCh <- true

	// update response, message, usage, context size
	s.Runtime.Response = resp
	if assistantMsg, ok := resp.Message(); ok {
		s.Runtime.Messages = append(s.Runtime.Messages, assistantMsg)
	}
	if s.Runtime.Response != nil {
		s.Runtime.Usage.Append(&s.Runtime.Response.Usage)

		s.Runtime.ContextSize = s.Runtime.Response.Usage.Prompt +
			s.Runtime.Response.Usage.Completion
	}

	// send reasoning event
	lastMsg := s.Runtime.Messages[len(s.Runtime.Messages)-1]
	if lastMsg.ReasoningContent != "" {
		s.Runtime.Emit(
			&response.AgentResponse{
				RespType:      response.AgentRespTypeMiddleMsg,
				MiddleMessage: lastMsg.ReasoningContent,
			},
		)
	}

	return nil
}

func (s *Session) ToolList() []llm.Tool {

	toolList := []llm.Tool{}
	for _, tool := range s.ToolState.ToolMap {
		toolList = append(toolList, *tool.Definition())
	}
	return toolList
}

// Save 保存 Session 到磁盘
func (s *Session) Save() error {
	sessionDir := path.Join(SessionDirBase, s.Meta.SessionID)

	metaPath := path.Join(sessionDir, "metadata.json")
	serializeJson(metaPath, s.Meta)

	messagePath := path.Join(sessionDir, "history.jsonl")

	serializeJson(messagePath, s.Runtime.Messages)

	return nil
}

// LoadSession 从磁盘 恢复 AgentSession
func LoadSession(
	sessionID string,
	mq *userinput.MessageQueue,
	output chan *response.AgentResponse,
) (*Session, error) {
	// TODO: 可从 内存registry -> 磁盘 分级读

	metaPath := path.Join(SessionDirBase, sessionID, "meta.json")
	toolStatePath := path.Join(SessionDirBase, sessionID, "tool_state.json")
	runtimePath := path.Join(SessionDirBase, sessionID, "runtime.json")

	meta, err := meta.LoadMeta(metaPath)
	if err != nil {return nil, err}
	toolState, err := toolstate.LoadToolState(toolStatePath)
	if err != nil {return nil, err}
	runtime, err := runtime.LoadRuntime(runtimePath, meta, mq, output)
	if err != nil {return nil, err}

	s := &Session{
		Meta: meta,
		ToolState: toolState,
		Runtime: runtime,
	}
	return s, nil
}

// NewSession 创建新 session
func NewSession(
	mq *userinput.MessageQueue,
	output chan *response.AgentResponse,
) (*Session, error) {
	meta := meta.NewMeta()
	toolState := toolstate.NewToolState()
	runtime, err := runtime.NewRuntime(meta, mq, output)
	if err != nil {return nil, err}

	s := &Session{
		Meta: meta,
		ToolState: toolState,
		Runtime: runtime,
	}

	return s, nil
}

func serializeJson(path string, data any) error {
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	if jsonBytes != nil {
		err = os.WriteFile(path, jsonBytes, 0644)
	}
	if err != nil {
		return err
	}
	return nil
}
