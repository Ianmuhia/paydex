package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

func GetLogger() *slog.Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout))
	slog.SetDefault(l)
	return l
}
