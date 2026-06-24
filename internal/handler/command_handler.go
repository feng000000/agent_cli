package handler

import (
	"fmt"
	"myagent/internal/command"
	"myagent/internal/runtime"
	"myagent/pkg/logger"
	"strings"
)


func HandleCommand(ctx *runtime.LoopContext) (string, error) {
	fields := strings.Fields(ctx.UserQuery)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := command.GetCommand(strings.TrimPrefix(fields[0], "/"))
	if cmd == nil {
		return "", fmt.Errorf("command %v not found", fields[0])
	}

	logger.Debugf("call command %v\n", cmd.Name())
	res, err := cmd.Exec(fields[1:])
	if err != nil {
		return "", err
	}

	return res, nil
}
