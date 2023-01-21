package worker

import (
	"golang.org/x/exp/slog"

	"fmt"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (logger *Logger) Print(args ...any) {
	slog.Log(slog.LevelInfo, "", fmt.Sprintf("%v", args))
}

func (logger *Logger) Debug(args ...any) {
	slog.Log(slog.LevelDebug, "", fmt.Sprintf("%v", args))
}

func (logger *Logger) Info(args ...any) {
	slog.Log(slog.LevelInfo, "", fmt.Sprintf("%v", args))
}

func (logger *Logger) Warn(args ...any) {
	slog.Log(slog.LevelWarn, "", fmt.Sprintf("%v", args))
}

func (logger *Logger) Error(args ...any) {
	slog.Log(slog.LevelError, "", fmt.Sprintf("%v", args))
}

func (logger *Logger) Fatal(args ...any) {
	slog.Log(slog.LevelError, "", fmt.Sprintf("%v", args))
}
