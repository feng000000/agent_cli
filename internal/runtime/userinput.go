package runtime

import "strings"

type InputTypeEnum string

const (
	InputTypePrompt  InputTypeEnum = "prompt"
	InputTypeCommand InputTypeEnum = "command"
)

type UserInput struct {
	Content []byte
}

// Type 用户输入的类型
func (ui *UserInput) Type() InputTypeEnum {
	if strings.HasPrefix(string(ui.Content), "/") {
		return InputTypeCommand
	}
	return InputTypePrompt
}
