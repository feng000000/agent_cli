package tool

import "fmt"
import "os"
import "sync"
import "time"
import "context"

import "myagent/pkg/llm"
import "myagent/pkg/logger"

type ToolMessage struct {
	// Type        string `json:"type"`
	ToolCallID  string `json:"tool_call_id"`
	ToolName    string `json:"tool_name"`
	Result      string `json:"result"`
	NeedConfirm bool   `json:"need_confirm"`
}

func ExecTool(
	s *Session,
	sessionMu *sync.RWMutex,
	toolCallID string,
	tool Tool,
	arg string,
) ToolMessage {
	ret := make(chan ToolMessage)

	go func() {
		toolMsg := ToolMessage{ToolCallID: toolCallID, ToolName: tool.Name()}
		defer func() {
			if r := recover(); r != nil {
				toolMsg.Result = fmt.Sprintf(
					"tool %v exec panic: %v", tool.Name(), r,
				)
				ret <- toolMsg
			}
		}()

		// TODO: check if tool need Confirm
		NeedConfirm := false
		if NeedConfirm {
			toolMsg.NeedConfirm = true
			ret <- toolMsg
			return
		}

		toolRes, err := tool.execute(s, sessionMu, arg)
		if err != nil {
			toolMsg.Result = fmt.Sprintf(
				"tool %v exec error: %v", tool.Name(), err.Error(),
			)
			ret <- toolMsg
			return
		}

		if len(toolRes) >= 512 {
			// 长工具输出落盘, 仅保留路径
			path, err := saveToolResult(toolRes)
			if err == nil {
				toolMsg.Result = fmt.Sprintf(
					"tool result too long, save to %v", path,
				)
				ret <- toolMsg
				return
			}

			// 保存失败, 回退到 直接加载到上下文
			s.OutputChan <- &AgentResponse{
				RespType: AgentRespTypeMiddleMsg,
				MiddleMessage: fmt.Sprintf(
					"save tool result failed, load to context directly: %v",
					err,
				),
			}
		}

		toolMsg.Result = toolRes
		ret <- toolMsg
		return
	}()


	// 超时
	ctx, cancel := context.WithTimeout(s.Ctx, tool.Timeout())
	defer cancel()

	select {
	case result:= <- ret:
		return result
	case <-ctx.Done():
		return ToolMessage{
			ToolCallID: toolCallID,
			ToolName: tool.Name(),
			Result: "tool exec timeout",
		}
	}

}

type Tool interface {
	// Name 标识工具名称
	Name() string

	// Desc 返回工具说明如何使用/何时使用
	Desc() string

	// Definition 为工具定义, 用于作为请求LLM 时的参数
	Definition() *llm.Tool

	Timeout() time.Duration

	// Execute 执行工具;
	// s 为当前会话的 session 对象
	// arg 为json字符串, 解析结果通过 res channel 传递
	execute(s *Session, sessionMu *sync.RWMutex, arg string) (string, error)
}


// saveToolResult 保存 工具输出到临时文件, 并返回文件路径
func saveToolResult(toolRes string) (string, error) {
	tmp, err := os.CreateTemp("", "tool-result-*.txt")
	if err != nil {
		return "", err
	}

	logger.Debugf("write tool result to %v", tmp.Name())

	_, err = tmp.Write([]byte(toolRes))

	if err != nil {
		return "", nil
	}
	return tmp.Name(), nil
}
