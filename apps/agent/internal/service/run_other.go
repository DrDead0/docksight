//go:build !windows

package service

import "context"

// IsService is always false off Windows. systemd runs the agent as an ordinary
// foreground process and stops it with SIGTERM, which lifecycle already
// handles — there is no separate supervised mode to detect.
func IsService() bool { return false }

// Run executes the agent directly. Signal handling belongs to lifecycle, so
// there is nothing to wrap here.
func Run(run func(context.Context) error) error {
	return run(context.Background())
}
