// Package scm manages the agent as a Windows Service Control Manager service.
//
// It is the Windows counterpart of package systemd, and satisfies the same
// interface the agent installer depends on. Three of those methods have no
// literal equivalent on Windows and are documented at their definitions:
// WriteUnit writes no file, DaemonReload does nothing, and Restarts counts
// something the SCM does not itself record.
//
// The real implementation is in scm_windows.go. Everything here compiles
// everywhere, so the parts that are pure string handling stay testable on the
// Linux machines this project is mostly developed on.
package scm

import (
	"os"
	"path/filepath"
	"strings"
)

// ServiceKeyPrefix is where the SCM keeps service configuration. Windows has
// no unit file; this registry key is the closest thing to one, and is what
// UnitPath reports so a written service is still named to the operator.
const ServiceKeyPrefix = `HKLM\SYSTEM\CurrentControlSet\Services\`

// DisplayName and Description are what the Services console shows.
const (
	DisplayName = "DockSight Agent"
	Description = "Reports Docker Engine state to a DockSight platform over an outbound WebSocket connection."
)

// DefaultLogPath is the file the agent writes to when the SCM runs it.
//
// It must match ServiceLogPath in apps/agent/internal/logger/sink_windows.go.
// The two are separate Go modules and cannot share the constant, so the CLI
// repeats it: a Windows service has no console and no journal, and this file
// is the only place installation verification can read the agent's own
// account of whether it reached the platform.
func DefaultLogPath() string {

	base := os.Getenv("ProgramData")

	if base == "" {
		base = `C:\ProgramData`
	}

	return filepath.Join(base, "DockSight", "logs", "agent.log")
}

// tail returns the last n lines of text, dropping a trailing newline so an
// ordinary log file does not yield a blank final line.
func tail(text string, lines int) string {

	if lines <= 0 {
		return text
	}

	trimmed := strings.TrimRight(text, "\r\n")

	if trimmed == "" {
		return ""
	}

	split := strings.Split(trimmed, "\n")

	if len(split) <= lines {
		return trimmed
	}

	return strings.Join(split[len(split)-lines:], "\n")
}

// tailFile returns the last n lines of a file.
//
// The whole file is read rather than seeked backwards through. The agent log
// rotates at 10 MiB, which bounds the cost, and verification reads it exactly
// twice per install.
func tailFile(path string, lines int) (string, error) {

	content, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return tail(string(content), lines), nil
}
