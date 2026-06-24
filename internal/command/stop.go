package command

import "myagent/internal/runtime"

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
	state *runtime.ClientState, args []string,
) (string, error) {
	state.CancelFunc()
	return "", nil
}
