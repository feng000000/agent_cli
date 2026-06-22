package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// levelEncoder 定义日志级别的显示格式
func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + l.CapitalString() + "]")
}

// timeEncoder 定义日志时间的显示格式
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05.000"))
}

// callerEncoder 定义日志调用方的显示格式
func callerEncoder(
	caller zapcore.EntryCaller,
	enc zapcore.PrimitiveArrayEncoder,
) {
	enc.AppendString("[" + caller.TrimmedPath() + "]:")
}
