package logger

import (
	"log/slog"
	"os"
)

func InitLogger(fmtInput string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	var logger *slog.Logger
	switch fmtInput {
	case "text":
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	case "color":
		logger = slog.New(newColorHandler(opts))
	default:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	slog.SetDefault(logger)
	return logger
}
