package command

import "context"
import "time"

import "myagent/internal/runtime"
import "myagent/pkg/logger"

type CialloCommand struct{}

func (c CialloCommand) Name() string {
	return "ciallo"
}

// Desc command 作用描述, 以及用法说明
func (c CialloCommand) Desc() string {
	return "example command."
}

// Exec 输出mock数据
func (c CialloCommand) Exec(
	ctx context.Context,
	state *runtime.AgentState,
	args []string,
) (string, error) {
	logger.Debugf("Exec ciallo, prepare to sleep\n")
	time.Sleep(time.Second * 5)

	return "Ciallo~", nil
}
