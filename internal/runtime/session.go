package runtime

import "os"
import "bufio"
import "context"
import "fmt"
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
	SessionID string `json:"session_id"`

	Persistence *Persistence         `json:"persistence"`
	LLM         LLMConfig            `json:"llm"`
	AgentMode   AgentModeEnum        `json:"agent_mode"`
	ToolMap     map[string]tool.Tool `json:"tool_map"`
}

func newMeta() *Meta {
	list_dir := &tool.ListDirTool{}
	return &Meta{
		SessionID: uuid.NewString(),
		AgentMode: AgentModePlan,
		ToolMap: map[string]tool.Tool{
			list_dir.Name(): list_dir,
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

	LLMClient    llm.LLMClient
	SystemPrompt string

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

func serializeJson(path string, data ...any) error {
	fullData := []byte{}
	for d := range data {
		jsonBytes, err := json.MarshalIndent(d, "", "    ")
		if err != nil {
			return err
		}
		fullData = append(fullData, jsonBytes...)
	}
	err := os.WriteFile(path, fullData, 0644)
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
	messages := make([]any, len(s.Messages))
	serializeJson(messagePath, messages...)

	return fmt.Errorf("saveState not implemented")
}

// LoadSession 从磁盘 恢复 AgentSession
func LoadSession(sessionID string) (*Session, error) {
	// TODO: implement LoadState, 可从 内存registry -> 磁盘 分级读

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

	systemPrompt, err := getSystemPrompt(meta.Persistence)
	if err != nil {
		return nil, err
	}
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

	// load messages
	messages := []llm.Message{llm.SystemMessage(systemPrompt)}
	func() {
		messagePath := path.Join(sessionDir, "history.jsonl")
		// 打开 JSONL 文件
		file, err := os.Open(messagePath)
		if err != nil {
			return
		}
		defer file.Close()

		// 使用 bufio 按行扫描
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Bytes()

			// 跳过空行或注释
			if len(line) == 0 {
				continue
			}

			// 解析 JSON 数据
			var message llm.Message
			if err := json.Unmarshal(line, &message); err != nil {
				logger.Warnf("load message err: %v\n", err)
				continue
			}

			messages = append(messages, message)
		}

		if err := scanner.Err(); err != nil {
			logger.Warnf("load history message failed: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	return &Session{
		Ctx:          ctx,
		cancel:       cancel,
		Meta:         meta,
		LLMClient:    llmClient,
		SystemPrompt: systemPrompt,
		Messages:     []llm.Message{llm.SystemMessage(systemPrompt)},
	}, nil
}

// NewSession 创建新 session
func NewSession() (*Session, error) {
	meta := newMeta()

	systemPrompt, err := getSystemPrompt(meta.Persistence)
	if err != nil {
		return nil, err
	}
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
	return &Session{
		Ctx:          ctx,
		cancel:       cancel,
		Meta:         meta,
		LLMClient:    llmClient,
		SystemPrompt: systemPrompt,
		Messages:     []llm.Message{llm.SystemMessage(systemPrompt)},
	}, nil
}
