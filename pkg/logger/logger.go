package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger      *zap.Logger = nil
	LogLevelMap             = map[string]zapcore.Level{
		"DEBUG":   zapcore.DebugLevel,
		"INFO":    zapcore.InfoLevel,
		"WARNING": zapcore.WarnLevel,
		"ERROR":   zapcore.ErrorLevel,
	}
)

func Debugf(msg string, v ...any) {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Debug(fmt.Sprintf(msg, v...))
}

func Infof(msg string, v ...any) {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Info(fmt.Sprintf(msg, v...))
}

func Warnf(msg string, v ...any) {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Warn(fmt.Sprintf(msg, v...))
}

func Errorf(msg string, v ...any) {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Error(fmt.Sprintf(msg, v...))
}

func Panicf(msg string, v ...any) {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Panic(fmt.Sprintf(msg, v...))
}

func Fatalf(msg string, v ...any) {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Fatal(fmt.Sprintf(msg, v...))
}

func Sync() {
	if logger == nil {
		panic("logger without InitLogger")
	}
	logger.Sync()
}

// InitLogger 用于初始化日志输出格式
//
//	Args:
//		- logLevel (zapcore.Level): 可选值为
//			logger.DebugLevel
//			logger.InfoLevel
//			logger.WarnLevel
//			logger.ErrorLevel
//
//		- logFilePath (string): JSON格式时, 日志文件的路径
//
//	Return:
//		- (error)
func InitLogger(level string, logFilePath string) error {
	logLevel, ok := LogLevelMap[level]
	if !ok {
		fmt.Printf("invalid log level: %v, set to DEBUG", level)
		logLevel = zap.DebugLevel
	}

	var stackTraceKey string
	if logLevel == zap.DebugLevel {
		stackTraceKey = "stacktrace"
	}
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "lv",
		TimeKey:       "time",
		CallerKey:     "caller",
		StacktraceKey: stackTraceKey,
	}

	var core zapcore.Core

	console_core := createConsoleCore(logLevel, encoderConfig)
	if logFilePath != "" {
		file_core, err := createFileCore(logLevel, encoderConfig, logFilePath)
		if err != nil {
			return err
		}

		core = zapcore.NewTee(*console_core, *file_core)
	} else {
		core = *console_core
	}

	logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(logLevel),
		zap.AddCallerSkip(1),
	)

	return nil
}

func createFileCore(
	logLevel zapcore.Level,
	encoderConfig zapcore.EncoderConfig,
	logFilePath string,
) (*zapcore.Core, error) {
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	err := os.MkdirAll(filepath.Dir(logFilePath), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return nil, err
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o775)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		zap.NewAtomicLevelAt(logLevel),
	)
	return &core, nil
}

func createConsoleCore(
	logLevel zapcore.Level,
	encoderConfig zapcore.EncoderConfig,
) *zapcore.Core {
	encoderConfig.EncodeLevel = levelEncoder
	encoderConfig.EncodeTime = timeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeCaller = callerEncoder
	encoderConfig.ConsoleSeparator = " "

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),

		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(logLevel),
	)
	return &core
}
