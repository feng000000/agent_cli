package command

type CialloCommand struct {}

func (c CialloCommand) Name() string {
	return "ciallo"
}

// Desc command 作用描述, 以及用法说明
func (c CialloCommand) Desc() string {
	return "example command."
}

// Exec 输出mock数据
func (c CialloCommand) Exec(args []string) (string, error) {
	return "Ciallo~", nil
}
