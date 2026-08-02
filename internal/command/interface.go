package command

import "context"
import "agentcli/internal/session"

// AgentCommand 由Agent处理的命令
type AgentCommand interface {
	// Name 命令名称, 可通过 /Name 调用
	Name() string

	// Desc command 作用描述, 以及用法说明
	Desc() string

	// Exec 执行命令
	Exec(
		ctx context.Context,
		state *session.Session,
		args []string,
	) (string, error)
}

// ClientCommand 由客户端处理的命令
type ClientCommand interface {
	// Name 命令名称, 可通过 /Name 调用
	Name() string

	// Desc command 作用描述, 以及用法说明
	Desc() string

	// Exec 执行命令
	Exec(
		ctx context.Context,
		state *session.ClientState,
		args []string,
	) (string, error)
}

// TODO: more command
var agentCmdRegistry = map[string]AgentCommand{
	"ciallo": CialloCommand{},
	"compress": CompressCommand{},
	"stop": StopCommand{},
}
var clientCmdRegistry = map[string]ClientCommand{
}

func GetAgentCommand(name string) AgentCommand {
	cmd, ok := agentCmdRegistry[name]
	if !ok {
		return nil
	}

	return cmd
}

func GetClientCommand(name string) ClientCommand {
	cmd, ok := clientCmdRegistry[name]
	if !ok {
		return nil
	}

	return cmd
}


// TODO: command decorator: include
// log, sandbox, panic recover,
// emit tool call event, empty response, long response(save to file system)
