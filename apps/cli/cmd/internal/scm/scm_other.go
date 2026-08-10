//go:build !windows

package scm

import (
	"context"
	"errors"
)

// errUnsupported is what every method returns off Windows.
var errUnsupported = errors.New("the Service Control Manager is only available on Windows")

// Manager is the off-Windows stand-in for the real manager.
//
// The Windows implementation cannot compile anywhere else —
// golang.org/x/sys/windows/svc does not build for Linux — but a package whose
// every file is excluded by a build constraint is a package the toolchain
// refuses to load at all, which would break `go build ./...` and `go vet
// ./...` on the machines this project is developed on. This keeps the package
// loadable and satisfies the same interface, refusing rather than pretending.
//
// Nothing reaches it in practice: the installer selects its controller in
// controller_windows.go and controller_other.go, and only the first imports
// this package.
type Manager struct {
	// LogPath mirrors the field the Windows manager reads its logs from.
	LogPath string
}

// NewManager returns a manager that refuses every operation.
func NewManager() *Manager { return &Manager{} }

func (m *Manager) UnitPath(unit string) string { return ServiceKeyPrefix + unit }

func (m *Manager) WriteUnit(string, string) (bool, error) { return false, errUnsupported }

func (m *Manager) DaemonReload(context.Context) error { return errUnsupported }

func (m *Manager) Enable(context.Context, string) error { return errUnsupported }

func (m *Manager) Disable(context.Context, string) error { return errUnsupported }

func (m *Manager) Start(context.Context, string) error { return errUnsupported }

func (m *Manager) Stop(context.Context, string) error { return errUnsupported }

func (m *Manager) Restart(context.Context, string) error { return errUnsupported }

func (m *Manager) IsActive(context.Context, string) (bool, error) { return false, errUnsupported }

func (m *Manager) Restarts(context.Context, string) (int, error) { return 0, errUnsupported }

func (m *Manager) Logs(context.Context, string, int) (string, error) { return "", errUnsupported }

func (m *Manager) RemoveUnit(string) error { return errUnsupported }
