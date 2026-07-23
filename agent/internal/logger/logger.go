package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var defaultLogger *slog.Logger = slog.Default()

// Setup configures the package-level structured logger.
func Setup(level string, out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, attr.Value.Time().UTC().Format(time.RFC3339))
			}
			return attr
		},
	}

	defaultLogger = slog.New(slog.NewTextHandler(out, opts))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug logs a debug-level message.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info-level message.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning-level message.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error-level message.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs an error-level message and exits the process.
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// Printf writes a plain line to stdout (used for the human startup summary).
func Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}
