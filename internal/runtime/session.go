package runtime

import "os"
import "context"
import "path"
import "encoding/json"

import "github.com/google/uuid"

import "myagent/pkg/llm"
import "myagent/pkg/logger"
import "myagent/internal/tool"

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
	listDir := &tool.ListDirTool{}
	readFile := &tool.ReadFileTool{}

	return &Meta{
		SessionID: uuid.NewString(),
		AgentMode: AgentModePlan,
		ToolMap: map[string]tool.Tool{
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
	Ctx    context.Context
	cancel context.CancelFunc

	Meta *Meta

	LLMClient llm.LLMClient

	// SystemPrompt string

	// LoadedSkillsPath []string

	Messages    []llm.Message
	Usage       llm.Usage
	ContextSize int
	Response    *llm.ChatResponse
	localMemory string
}

func (s *Session) ToolList() []llm.Tool {
	toolList := []llm.Tool{}
	for _, tool := range s.Meta.ToolMap {
		toolList = append(toolList, *tool.Definition())
	}
	return toolList
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

// Cancel 取消 session 的Ctx
func (s *Session) Cancel() {
	s.cancel()
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
func LoadSession(sessionID string) (*Session, error) {
	session := &Session{}

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
	session.Meta = meta

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
	session.LLMClient = llmClient

	// load messages
	systemPrompt, err := session.SystemPrompt()
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
	session.Messages = messages

	logger.Debugf("loaded messages: %v", messages)

	ctx, cancel := context.WithCancel(context.Background())
	session.Ctx = ctx
	session.cancel = cancel

	return session, nil
}

// NewSession 创建新 session
func NewSession() (*Session, error) {
	session := &Session{}

	session.Meta = newMeta()

	llmClient, err := llm.NewDeepSeekClient(
		llm.DeepSeekConfig{
			APIKey:  session.Meta.LLM.APIKey,
			BaseURL: session.Meta.LLM.BaseURL,
			Model:   session.Meta.LLM.Model,
		},
	)
	if err != nil {
		return nil, err
	}
	session.LLMClient = llmClient

	systemPrompt, err := session.SystemPrompt()
	if err != nil {
		return nil, err
	}
	session.Messages = []llm.Message{llm.SystemMessage(systemPrompt)}

	ctx, cancel := context.WithCancel(context.Background())
	session.Ctx = ctx
	session.cancel = cancel

	return session, nil
}
