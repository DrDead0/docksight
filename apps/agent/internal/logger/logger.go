package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	defaultLogger *slog.Logger = slog.Default()

	// output is where Printf writes. Tracked separately from the slog handler
	// because Printf predates Setup taking a writer and used fmt.Printf, which
	// goes to stdout unconditionally — under a Windows service there is no
	// console, so the entire startup summary was written to a handle nobody
	// could read.
	output io.Writer = os.Stdout

	// outputMu guards output. Setup runs during startup while nothing else is
	// logging, but Printf is reachable from any goroutine afterwards.
	outputMu sync.RWMutex
)

// Setup configures the package-level structured logger.
func Setup(level string, out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	outputMu.Lock()
	output = out
	outputMu.Unlock()

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

// Printf writes a plain line to the configured output (used for the human
// startup summary).
func Printf(format string, args ...any) {
	outputMu.RLock()
	out := output
	outputMu.RUnlock()

	fmt.Fprintf(out, format, args...)
}
