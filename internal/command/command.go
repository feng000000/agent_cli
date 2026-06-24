package command


type Command interface {
	// Name 命令名称, 可通过 /Name 调用
	Name() string

	// Desc command 作用描述, 以及用法说明
	Desc() string

	// Exec 执行命令
	Exec(args []string) (string, error)
}


// TODO: more command (core)
var cmdRegistry map[string]Command = map[string]Command{
	"ciallo": CialloCommand{},
}


func GetCommand(name string) Command {
	cmd, ok := cmdRegistry[name]
	if !ok {
		return nil
	}

	return cmd
}
