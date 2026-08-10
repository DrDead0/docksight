//go:build windows

package logger

import (
	"os"
	"path/filepath"
)

// ServiceLogPath is where the agent writes its log when the Service Control
// Manager runs it.
//
// A Windows service has no console, so stdout goes nowhere and the Linux
// arrangement — write to stdout, let the journal capture it — has no
// equivalent. ProgramData is the conventional location for machine-wide
// service state, and this is the file "docksight agent install" reads back to
// verify the agent connected.
func ServiceLogPath() string {

	base := os.Getenv("ProgramData")

	if base == "" {
		base = `C:\ProgramData`
	}

	return filepath.Join(base, "DockSight", "logs", "agent.log")
}
