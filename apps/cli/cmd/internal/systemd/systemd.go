// Package systemd manages unit files and the services they define.
package systemd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// DefaultUnitDir is where distribution-independent units belong.
const DefaultUnitDir = "/etc/systemd/system"

// CommandRunner executes a program and returns its combined output. It is an
// interface so callers can drive systemd in tests without root or systemd.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {

	command := exec.CommandContext(ctx, name, args...)

	return command.CombinedOutput()
}

// Manager drives systemctl and journalctl.
type Manager struct {
	UnitDir string
	Runner  CommandRunner
}

// NewManager returns a manager for the system unit directory.
func NewManager() *Manager {

	return &Manager{
		UnitDir: DefaultUnitDir,
		Runner:  execRunner{},
	}
}

// UnitPath is where a unit file is written.
func (m *Manager) UnitPath(unit string) string {
	return filepath.Join(m.unitDir(), unit)
}

// WriteUnit installs a unit file. It reports whether the content changed, so
// callers can skip a daemon-reload on a repeated install.
func (m *Manager) WriteUnit(unit string, content string) (bool, error) {

	path := m.UnitPath(unit)

	if existing, err := os.ReadFile(path); err == nil && string(existing) == content {
		return false, nil
	}

	if err := os.MkdirAll(m.unitDir(), 0o755); err != nil {
		return false, err
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("failed to write %s: %w", path, err)
	}

	return true, nil
}

// RemoveUnit deletes a unit file if it exists.
func (m *Manager) RemoveUnit(unit string) error {

	err := os.Remove(m.UnitPath(unit))

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// DaemonReload makes systemd re-read unit files.
func (m *Manager) DaemonReload(ctx context.Context) error {
	return m.systemctl(ctx, "daemon-reload")
}

// Enable starts the unit on boot.
func (m *Manager) Enable(ctx context.Context, unit string) error {
	return m.systemctl(ctx, "enable", unit)
}

// Disable stops the unit from starting on boot.
func (m *Manager) Disable(ctx context.Context, unit string) error {
	return m.systemctl(ctx, "disable", unit)
}

// Start starts the unit.
func (m *Manager) Start(ctx context.Context, unit string) error {
	return m.systemctl(ctx, "start", unit)
}

// Stop stops the unit.
func (m *Manager) Stop(ctx context.Context, unit string) error {
	return m.systemctl(ctx, "stop", unit)
}

// Restart restarts the unit, starting it if it is not running. Reinstalling
// over a running agent needs this rather than start, which is a no-op when
// the service is already up with the old binary.
func (m *Manager) Restart(ctx context.Context, unit string) error {
	return m.systemctl(ctx, "restart", unit)
}

// IsActive reports whether the unit is currently running.
func (m *Manager) IsActive(ctx context.Context, unit string) (bool, error) {

	// `is-active` exits non-zero for inactive units, which is not an error
	// condition here — the state itself is the answer.
	output, _ := m.Runner.Run(ctx, "systemctl", "is-active", unit)

	return strings.TrimSpace(string(output)) == "active", nil
}

// Property reads a single unit property, e.g. NRestarts or ActiveState.
func (m *Manager) Property(ctx context.Context, unit string, property string) (string, error) {

	output, err := m.Runner.Run(ctx, "systemctl", "show", unit, "--property", property, "--value")

	if err != nil {
		return "", fmt.Errorf("failed to read %s of %s: %s", property, unit, trim(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// Restarts reports how many times systemd has restarted the unit. A service
// that keeps restarting is failing even though it reports active.
func (m *Manager) Restarts(ctx context.Context, unit string) (int, error) {

	value, err := m.Property(ctx, unit, "NRestarts")

	if err != nil {
		return 0, err
	}

	if value == "" {
		return 0, nil
	}

	count, err := strconv.Atoi(value)

	if err != nil {
		return 0, nil
	}

	return count, nil
}

// Logs returns the most recent journal lines for the unit.
func (m *Manager) Logs(ctx context.Context, unit string, lines int) (string, error) {

	if lines <= 0 {
		lines = 50
	}

	output, err := m.Runner.Run(
		ctx,
		"journalctl",
		"--unit", unit,
		"--lines", strconv.Itoa(lines),
		"--no-pager",
		"--output", "cat",
	)

	if err != nil {
		return "", fmt.Errorf("failed to read logs for %s: %s", unit, trim(output))
	}

	return string(output), nil
}

func (m *Manager) systemctl(ctx context.Context, args ...string) error {

	output, err := m.Runner.Run(ctx, "systemctl", args...)

	if err != nil {
		return fmt.Errorf("systemctl %s failed: %s", strings.Join(args, " "), trim(output))
	}

	return nil
}

func (m *Manager) unitDir() string {

	if m.UnitDir == "" {
		return DefaultUnitDir
	}

	return m.UnitDir
}

func trim(output []byte) string {

	message := strings.TrimSpace(string(output))

	if message == "" {
		return "no output"
	}

	return message
}
