//go:build !windows

package logger

// ServiceLogPath is empty off Windows.
//
// systemd captures the agent's stdout into the journal, which already gives
// operators rotation, retention and journalctl. Writing a second copy to a
// file would duplicate all of it.
func ServiceLogPath() string { return "" }
