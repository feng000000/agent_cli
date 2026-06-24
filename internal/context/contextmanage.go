package context

import "myagent/pkg/llm"



func GetSystemPrompt() string {
	// TODO: Role, datetime, workspace, user.md, memory.md
	// tool [desc], when to use
	return "You are a helpful assistant."
}


func Compact(messages []llm.Message) []llm.Message {
	// TODO: implement
	return messages
}
