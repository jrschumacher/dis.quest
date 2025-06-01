package logger

import (
	"log/slog"
	"os"
	"strings"
)

var defaultLogger *slog.Logger

func Init(level string) {
	var slogLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "INFO":
		slogLevel = slog.LevelInfo
	case "WARN":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	defaultLogger = slog.New(h)
}

func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// Logger returns the default logger instance.
func Logger() *slog.Logger {
	return defaultLogger
}

// SetLogger allows replacing the default logger (for tests or customization).
func SetLogger(l *slog.Logger) {
	defaultLogger = l
}
