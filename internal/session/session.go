package session

import "os"
import "context"
import "time"
import "path"
import "encoding/json"


import "myagent/pkg/llm"
import "myagent/pkg/logger"

import "myagent/internal/session/state"
import "myagent/internal/session/runtime"
import "myagent/internal/session/meta"

// TODO:  ClientState: client 分开实现
type ClientState struct {
	CancelFunc func()
}



const SessionDirBase = "./.myagent/sessions"

type Session struct {
	Meta *meta.Meta

	State *state.State

	Runtime *runtime.Runtime
}

func (s *Session) RawLLMClient() llm.LLMClient {
	return s.Runtime.llmClient
}

func (s *Session) ForceUpdateContextInfo(
	messages []llm.Message,
	usage *llm.Usage,
	contextSize int,
) {
	s.State.Messages = messages
	s.State.Usage = usage
	s.State.ContextSize = contextSize
}

func (s *Session) Emit(resp *AgentResponse) {
	s.Runtime.outputChan <- resp
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
					s.Emit(
						&AgentResponse{
							RespType:      AgentRespTypeMiddleMsg,
							MiddleMessage: "got llm resp",
						},
					)
				} else {
					s.Emit(
						&AgentResponse{
							RespType:      AgentRespTypeMiddleMsg,
							MiddleMessage: "got llm resp failed",
						},
					)
				}
				return
			case <-ticker.C:
				s.Runtime.outputChan <- &AgentResponse{
					RespType:      AgentRespTypeMiddleMsg,
					MiddleMessage: "waiting",
				}
			}
		}
	}()

	sysPrompt, err := s.SystemPrompt()
	if err != nil {
		return err
	}
	msgs := append([]llm.Message{llm.SystemMessage(sysPrompt)}, s.State.Messages...)
	msgs = append(msgs, newMsgs...)
	resp, err := s.Runtime.llmClient.Chat(
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
		s.State.Messages = append(s.State.Messages, assistantMsg)
	}
	if s.Runtime.Response != nil {
		s.State.Usage.Append(&s.Runtime.Response.Usage)

		s.State.ContextSize = s.Runtime.Response.Usage.Prompt +
			s.Runtime.Response.Usage.Completion
	}

	// send reasoning event
	lastMsg := s.State.Messages[len(s.State.Messages)-1]
	if lastMsg.ReasoningContent != "" {
		s.Emit(
			&AgentResponse{
				RespType:      AgentRespTypeMiddleMsg,
				MiddleMessage: lastMsg.ReasoningContent,
			},
		)
	}

	return nil
}

func (s *Session) ToolList() []llm.Tool {

	toolList := []llm.Tool{}
	for _, tool := range s.Meta.ToolMap {
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

	// OPTIMIZE: 改为jsonl 格式, 方便追加写入历史
	serializeJson(messagePath, s.State.Messages)

	return nil
}

// LoadSession 从磁盘 恢复 AgentSession
func LoadSession(
	sessionID string,
	mq *MessageQueue,
	output chan *AgentResponse,
) (*Session, error) {
	s := &Session{
		MessageQueue: mq,
		OutputChan:   output,
	}

	// TODO: 可从 内存registry -> 磁盘 分级读

	sessionDir := path.Join(SessionDirBase, sessionID)
	logger.Debugf("load state from %v", sessionDir)

	// load meta
	meta := newMeta()
	meta.SessionID = sessionID
	{
		metaPath := path.Join(sessionDir, "metadata.json")
		file, err := os.Open(metaPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)

		if err := decoder.Decode(&meta); err != nil {
			logger.Warnf("load meta from %v failed, New meta", metaPath)
			meta = newMeta()
		}
	}
	s.Meta = meta

	runtime := &Runtime{MessageQueue: mq, outputChan: output}
	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  meta.LLM.APIKey,
			BaseURL: meta.LLM.BaseURL,
			Model:   meta.LLM.Model,
		},
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	runtime.llmClient = llmClient
	runtime.Ctx = ctx
	runtime.Cancel = cancel

	s.Runtime = runtime

	state := &State{}
	// load messages
	systemPrompt, err := s.SystemPrompt()
	if err != nil {
		return nil, err
	}
	messages := []llm.Message{llm.SystemMessage(systemPrompt)}
	func() {
		messagePath := path.Join(sessionDir, "history.jsonl")

		file, err := os.Open(messagePath)
		if err != nil {
			return
		}
		defer file.Close()

		decoder := json.NewDecoder(file)

		if err := decoder.Decode(&messages); err != nil {
			logger.Warnf("load history message failed: %v", err)
		}
	}()
	s.Messages = messages

	logger.Debugf("loaded messages: %v", messages)

	s.Ctx = ctx
	s.Cancel = cancel

	return s, nil
}

// NewSession 创建新 session
func NewSession(
	mq *MessageQueue,
	output chan *AgentResponse,
) (*Session, error) {
	s := &Session{
		MessageQueue: mq,
		OutputChan:   output,
	}

	s.Meta = newMeta()

	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  s.Meta.LLM.APIKey,
			BaseURL: s.Meta.LLM.BaseURL,
			Model:   s.Meta.LLM.Model,
		},
	)
	if err != nil {
		return nil, err
	}
	s.llmClient = llmClient

	systemPrompt, err := s.SystemPrompt()
	if err != nil {
		return nil, err
	}
	s.Messages = []llm.Message{llm.SystemMessage(systemPrompt)}

	ctx, cancel := context.WithCancel(context.Background())
	s.Ctx = ctx
	s.Cancel = cancel

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
