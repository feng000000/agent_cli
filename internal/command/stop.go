package command

import "context"
import "myagent/internal/session"

type StopCommand struct{}

func (c StopCommand) Name() string {
	return "stop"
}

// Desc command 作用描述, 以及用法说明
func (c StopCommand) Desc() string {
	return "stop agent"
}

// Exec 停止Agent 运行
func (c StopCommand) Exec(
	ctx context.Context,
	s *session.Session,
	args []string,
) (string, error) {
	// state.Cancel()
	// return fmt.Sprintf("stop session %v", state.Meta.SessionID), nil
	// TODO: 使用 ACP 停止
	return "not implemented", nil
}
