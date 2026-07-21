package runtime

import "os"
import "sync"
import "context"
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

	ToolMap map[string]tool.Tool `json:"-"`
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
	mu   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	meta *Meta

	llmClient llm.LLMClient

	// SystemPrompt string

	// LoadedSkillsPath []string

	messages    []llm.Message
	usage       llm.Usage
	contextSize int
	response    *llm.ChatResponse
	localMemory string
}

func (s *Session) UpdateResponse(r *llm.ChatResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.response = r
}

func (s *Session) Messages() []llm.Message {

}

func (s *Session) AppendMessage(newMessage llm.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, newMessage)
}

func (s *Session) ToolList() []llm.Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	toolList := []llm.Tool{}
	for _, tool := range s.meta.ToolMap {
		toolList = append(toolList, *tool.Definition())
	}
	return toolList
}

// Cancel 取消 session 的Ctx
func (s *Session) Cancel() {
	s.cancel()
}

// Save 保存 Session 到磁盘
func (s *Session) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionDir := path.Join(SessionDirBase, s.Meta.SessionID)

	metaPath := path.Join(sessionDir, "metadata.json")
	serializeJson(metaPath, s.Meta)

	messagePath := path.Join(sessionDir, "history.jsonl")

	// OPTIMIZE: 改为jsonl 格式, 方便追加写入历史
	serializeJson(messagePath, s.Messages)

	return nil
}

// LoadSession 从磁盘 恢复 AgentSession
func LoadSession(sessionID string) (*Session, error) {
	s := &Session{}

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
	s.LLMClient = llmClient

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
	s.messages = messages

	logger.Debugf("loaded messages: %v", messages)

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	return s, nil
}

// NewSession 创建新 session
func NewSession() (*Session, error) {
	s := &Session{}

	s.meta = newMeta()

	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  s.meta.LLM.APIKey,
			BaseURL: s.meta.LLM.BaseURL,
			Model:   s.meta.LLM.Model,
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
	s.messages = []llm.Message{llm.SystemMessage(systemPrompt)}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

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
