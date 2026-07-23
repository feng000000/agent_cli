package runtime

import "os"
import "context"
import "time"
import "path"
import "encoding/json"

import "github.com/google/uuid"

import "myagent/pkg/llm"
import "myagent/pkg/logger"

// import "myagent/internal/tool"

// TODO:  ClientState: client 分开实现
type ClientState struct {
	CancelFunc func()
}

type AgentModeEnum string

const (
	AgentModePlan     AgentModeEnum = "plan"
	AgentModeAutoEdit AgentModeEnum = "auto-edit"
	AgentModeAutoExec AgentModeEnum = "auto-exec"
)

type Persistence struct {
	PersistenceDir string `json:"persistence_dir"`
	MemoryPath     string `json:"memory_path"`
	UserInfoPath   string `json:"userinfo_path"`
	SkillDir       string `json:"skill_dir"`
}

type LLMConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type Meta struct {
	SessionID           string        `json:"session_id"`
	LLM                 LLMConfig     `json:"llm"`
	AgentMode           AgentModeEnum `json:"agent_mode"`
	MaxTokensToCompress int           `json:"max_tokens_to_compress"`
	Persistence         *Persistence  `json:"persistence"`

	ToolMap map[string]Tool `json:"-"`
}

func newMeta() *Meta {
	// TODO: default tools
	listDir := &ListDirTool{}
	readFile := &ReadFileTool{}

	return &Meta{
		SessionID: uuid.NewString(),
		AgentMode: AgentModePlan,
		ToolMap: map[string]Tool{
			listDir.Name():  listDir,
			readFile.Name(): readFile,
		},
		Persistence: &Persistence{
			PersistenceDir: "./.myagent/persistence",
			MemoryPath:     "./.myagent/persistence/memory.md",
			UserInfoPath:   "./.myagent/persistence/user_info.md",
			SkillDir:       "./.myagent/persistence/skills/",
		},
	}
}

const SessionDirBase = "./.myagent/sessions"

type Session struct {
	llmClient llm.LLMClient

	// MessageQueue 可以直接获取追加信息
	MessageQueue *MessageQueue
	// OutputChan emit ACP 事件
	OutputChan chan *AgentResponse

	Ctx    context.Context
	Cancel context.CancelFunc

	Meta        *Meta
	Messages    []llm.Message
	Usage       *llm.Usage
	ContextSize int
	Response    *llm.ChatResponse
	LocalMemory string
}

func (s *Session) RawLLMClient() llm.LLMClient {
	return s.llmClient
}

func (s *Session) ForceUpdateContextInfo(
	messages []llm.Message,
	usage *llm.Usage,
	contextSize int,
) {
	s.Messages = messages
	s.Usage = usage
	s.ContextSize = contextSize
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
					s.OutputChan <- &AgentResponse{
						RespType:      AgentRespTypeMiddleMsg,
						MiddleMessage: "got llm resp",
					}
				} else {
					s.OutputChan <- &AgentResponse{
						RespType:      AgentRespTypeMiddleMsg,
						MiddleMessage: "got llm resp failed",
					}
				}
				return
			case <-ticker.C:
				s.OutputChan <- &AgentResponse{
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
	msgs := append([]llm.Message{llm.SystemMessage(sysPrompt)}, s.Messages...)
	msgs = append(msgs, newMsgs...)
	resp, err := s.llmClient.Chat(
		s.Ctx,
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
	s.Response = resp
	if assistantMsg, ok := resp.Message(); ok {
		s.Messages = append(s.Messages, assistantMsg)
	}
	if s.Response != nil {
		s.Usage.Append(&s.Response.Usage)

		s.ContextSize = s.Response.Usage.Prompt +
			s.Response.Usage.Completion
	}

	// send reasoning event
	if lastMsg := s.Messages[len(s.Messages)-1]; lastMsg.ReasoningContent != "" {
		s.OutputChan <- &AgentResponse{
			RespType:      AgentRespTypeMiddleMsg,
			MiddleMessage: lastMsg.ReasoningContent,
		}
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
	serializeJson(messagePath, s.Messages)

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
	s.llmClient = llmClient

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

	ctx, cancel := context.WithCancel(context.Background())
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
