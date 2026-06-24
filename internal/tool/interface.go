package tool

import "myagent/pkg/llm"


type Tool interface {
	// Name 标识工具名称
	Name() string

	// Definition 为工具定义, 用于作为请求LLM 时的参数
	Definition() *llm.Tool

	// Execute 执行工具;
	// arg 为json字符串, 解析结果通过 res channel 传递
	Execute(args string, res chan string)
}
