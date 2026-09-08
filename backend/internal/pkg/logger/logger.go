// Package logger configures the process-wide structured logger.
package logger

import (
	"log/slog"
	"os"
)

func NewLogger(isProduction bool) *slog.Logger {
	level := slog.LevelDebug
	if isProduction {
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if isProduction {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
