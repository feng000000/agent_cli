package executor

import "encoding/json"

type RespCodeEnum int

const (
	RespCodeSuccess        RespCodeEnum = iota
	RespCodeInvalidContent RespCodeEnum = iota
	RespCodeInvalidParam   RespCodeEnum = iota
	RespCodeExecuteFailed  RespCodeEnum = iota
)

type Request struct {
	ToolName  string          `json:"tool_name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type Response struct {
	Code    RespCodeEnum `json:"code"`
	Content any          `json:"content,omitempty"`
}
