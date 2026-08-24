package utils

import (
	"log/slog"
	"os"
)

func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(),
	})
	slog.SetDefault(slog.New(handler))
}

func getLogLevel() slog.Level {
	tempLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "ERROR":
		return slog.LevelError
	case "WARNING":
		return slog.LevelWarn
	case "INFO":
		return slog.LevelInfo
	case "DEBUG":
		return slog.LevelDebug
	default:
		tempLogger.Warn("LOG_LEVEL env var is either not set or invalid. Using default value.",
			slog.String("provided_value", logLevel),
			slog.String("fallback_value", "INFO"),
		)
		return slog.LevelInfo
	}
}
