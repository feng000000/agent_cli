package runtime

import "fmt"
import "sync"

import "myagent/pkg/llm"


// TODO: implement ExecuteWrapper()
// TODO: 长工具输出落盘, 仅返回路径
// TODO: 超时
func ExecTool(
	s *Session,
	sessionMu *sync.RWMutex,
	tool Tool,
	arg string,
) string {
	ret := make(chan string)
	go func() {
		defer func(){
			if r := recover(); r != nil {
				ret <- fmt.Sprintf("tool %v exec panic: %v", tool.Name(), r)
			}
		}()

		toolRes, err := tool.execute(s, sessionMu, arg)
		if err != nil {
			ret <- fmt.Sprintf(
				"tool %v exec error: %v",
				tool.Name(),
				err.Error(),
			)
			return
		}
		ret <- toolRes
		return
	}()

	return <- ret
}

type Tool interface {
	// Name 标识工具名称
	Name() string

	// Desc 返回工具说明如何使用/何时使用
	Desc() string

	// Definition 为工具定义, 用于作为请求LLM 时的参数
	Definition() *llm.Tool

	// Execute 执行工具;
	// s 为当前会话的 session 对象
	// arg 为json字符串, 解析结果通过 res channel 传递
	execute(s *Session, sessionMu *sync.RWMutex, arg string) (string, error)
}
