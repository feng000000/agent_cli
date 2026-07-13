package handler

import "context"
import "fmt"
import "strings"

import "myagent/internal/command"
import "myagent/internal/runtime"
import "myagent/pkg/logger"

var SkipHandleCommand = fmt.Errorf("skip command")

type cmd struct {
	Cmd  string
	Args []string
}

func splitCommand(query string) (*cmd, error) {
	parts := strings.Fields(query)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return &cmd{Cmd: strings.TrimPrefix(parts[0], "/"), Args: parts[1:]}, nil

}

func HandleClientCommand(
	ctx context.Context,
	state *runtime.ClientState,
	query string,
) (string, error) {
	cmdInput, err := splitCommand(query)
	if err != nil {
		return "", err
	}

	cmd := command.GetClientCommand(cmdInput.Cmd)
	if cmd == nil {
		return "", SkipHandleCommand
	}

	res, err := cmd.Exec(ctx, state, cmdInput.Args)
	return res, nil
}

// HandleAgentCommand 处理Agent运行时的命令
func HandleAgentCommand(
	ctx context.Context,
	state *runtime.Session,
	query string,
) (string, error) {
	cmdInput, err := splitCommand(query)
	if err != nil {
		return "", err
	}

	cmd := command.GetAgentCommand(cmdInput.Cmd)
	if cmd == nil {
		return "", fmt.Errorf("command %v not found", cmdInput.Cmd)
	}

	logger.Debugf("call command %v\n", cmd.Name())
	res, err := cmd.Exec(ctx, state, cmdInput.Args)
	if err != nil {
		return "", err
	}

	return res, nil
}
