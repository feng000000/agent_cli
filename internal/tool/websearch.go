package tool


import "myagent/pkg/llm"

// TODO: implement WebSearchTool
type WebSearchTool struct {}

func (wt *WebSearchTool) Name() string {
	return "web-search"
}

// Desc 返回工具说明如何使用/何时使用
func (wt *WebSearchTool) Desc() string {
	return "not implemented"
}

// Definition 为工具定义, 用于作为请求LLM 时的参数
func (wt *WebSearchTool) Definition() *llm.Tool {
	return &llm.Tool{}
}

// Execute 执行工具;
// arg 为json字符串, 解析结果通过 res channel 传递
func (wt *WebSearchTool) Execute(args string) {

}
