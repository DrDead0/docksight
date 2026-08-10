//go:build windows

package scm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// Recovery tuning, chosen to match the systemd unit: Restart=always with
// RestartSec=5s. Windows expresses the same intent as failure actions, and
// repeats the last action indefinitely once the list runs out, so three
// entries give "always restart, five seconds apart".
const (
	restartDelay = 5 * time.Second

	// resetPeriod is how long the service must stay up before Windows forgets
	// its failure count. A day is long enough that a slow crash loop still
	// escalates through the action list rather than resetting between crashes.
	resetPeriod = uint32(24 * 60 * 60)
)

// stateTimeout bounds how long a stop or start may take to be observed. The
// SCM accepts the request and returns immediately; the transition is watched
// afterwards.
const (
	stateTimeout = 30 * time.Second
	statePoll    = 250 * time.Millisecond
)

// Manager drives the Service Control Manager.
//
// It is created per command and holds no handles: every method connects,
// acts, and disconnects. The only state it keeps is the process ID last seen
// for a service, which is what makes Restarts answerable — see there.
type Manager struct {
	// LogPath is the agent's log file, read by Logs. Empty means the default.
	LogPath string

	mu sync.Mutex

	// observed is the process ID last seen for each service, and how many
	// times it has changed.
	observed map[string]uint32
	restarts map[string]int
}

// NewManager returns a manager for the local Service Control Manager.
func NewManager() *Manager {

	return &Manager{
		observed: map[string]uint32{},
		restarts: map[string]int{},
	}
}

// UnitPath is where the service is defined.
//
// Windows has no unit file. Service configuration lives in the SCM database,
// which is backed by this registry key, so that is what an operator can
// actually go and look at.
func (m *Manager) UnitPath(unit string) string {
	return ServiceKeyPrefix + unit
}

// WriteUnit creates or updates the service.
//
// There is no file to write. The installer's "unit content" is, on this
// platform, the command line the SCM stores and executes — that is the whole
// of what a service definition is here, and rendering it is the analogue of
// rendering a unit file. Everything else the systemd unit expresses is set
// alongside it: automatic start, and failure actions standing in for
// Restart=always.
//
// It reports whether anything changed, so a repeat install can say so rather
// than claiming to have rewritten a service it left alone. Reinstalling over
// an existing service updates it in place; the service is never deleted and
// recreated, which would drop its recovery configuration and its Event Log
// registration on every upgrade.
func (m *Manager) WriteUnit(unit string, content string) (bool, error) {

	manager, err := connect()

	if err != nil {
		return false, err
	}

	defer manager.Disconnect()

	service, err := manager.OpenService(unit)

	if err != nil {

		if !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return false, fmt.Errorf("failed to open the %s service: %w", unit, elevationAware(err))
		}

		return true, create(manager, unit, content)
	}

	defer service.Close()

	config, err := service.Config()

	if err != nil {
		return false, fmt.Errorf("failed to read the %s service configuration: %w", unit, err)
	}

	unchanged := config.BinaryPathName == content &&
		config.StartType == mgr.StartAutomatic &&
		config.DisplayName == DisplayName

	if unchanged {
		return false, nil
	}

	config.BinaryPathName = content
	config.StartType = mgr.StartAutomatic
	config.DisplayName = DisplayName
	config.Description = Description

	if err := service.UpdateConfig(config); err != nil {
		return false, fmt.Errorf("failed to update the %s service: %w", unit, elevationAware(err))
	}

	// Recovery actions are reapplied rather than compared: they are not part
	// of the config struct, and an operator who cleared them by hand in the
	// Services console should get them back on the next install.
	if err := applyRecovery(service); err != nil {
		return true, err
	}

	registerEventSource(unit)

	return true, nil
}

// create registers a new service.
func create(manager *mgr.Mgr, unit string, content string) error {

	// CreateService builds BinaryPathName itself, escaping the executable and
	// each argument separately, which is not necessarily the string that was
	// rendered. The configuration is written again immediately so the stored
	// command line is exactly the one WriteUnit was handed — otherwise the
	// next install would compare the two, find a difference that is not one,
	// and rewrite a service that had not changed.
	service, err := manager.CreateService(unit, content, mgr.Config{
		DisplayName:  DisplayName,
		Description:  Description,
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
	})

	if err != nil {
		return fmt.Errorf("failed to create the %s service: %w", unit, elevationAware(err))
	}

	defer service.Close()

	config, err := service.Config()

	if err != nil {
		return fmt.Errorf("failed to read the %s service configuration: %w", unit, err)
	}

	config.BinaryPathName = content

	if err := service.UpdateConfig(config); err != nil {
		return fmt.Errorf("failed to set the %s command line: %w", unit, err)
	}

	if err := applyRecovery(service); err != nil {
		return err
	}

	registerEventSource(unit)

	return nil
}

// applyRecovery configures what Windows does when the agent exits
// unexpectedly. It is the counterpart of Restart=always in the systemd unit.
func applyRecovery(service *mgr.Service) error {

	actions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: restartDelay},
		{Type: mgr.ServiceRestart, Delay: restartDelay},
		{Type: mgr.ServiceRestart, Delay: restartDelay},
	}

	if err := service.SetRecoveryActions(actions, resetPeriod); err != nil {
		return fmt.Errorf("failed to set recovery actions for %s: %w", service.Name, err)
	}

	// Without this, Windows only counts a crash as a failure. The agent stops
	// with a non-zero exit code when its configuration is unusable, which is
	// exactly the case that must be retried rather than left stopped.
	if err := service.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		return fmt.Errorf("failed to set recovery actions for %s: %w", service.Name, err)
	}

	return nil
}

// registerEventSource lets the agent write lifecycle events under its own
// name in the Event Log.
//
// Registration needs Administrator rights, which is why it happens at install
// time and not in the agent. The error is discarded on purpose, and covers
// two cases that are both fine: the source already exists, which is what a
// reinstall always finds; or it could not be created, which costs an
// administrator the entries in Event Viewer and nothing else. The agent's
// full log still goes to disk, and it starts and runs either way — refusing
// to install over that would be out of all proportion.
func registerEventSource(unit string) {

	_ = eventlog.InstallAsEventCreate(
		unit,
		eventlog.Info|eventlog.Warning|eventlog.Error,
	)
}

// DaemonReload does nothing.
//
// systemd reads unit files from disk and needs telling when they change. The
// SCM has no such cache: WriteUnit changed the service database directly, and
// there is nothing left to reload. Implemented as a no-op rather than removed
// from the interface so the Linux path keeps the call it genuinely needs.
func (m *Manager) DaemonReload(context.Context) error {
	return nil
}

// Enable starts the service at boot.
func (m *Manager) Enable(_ context.Context, unit string) error {

	return m.withService(unit, func(service *mgr.Service) error {

		config, err := service.Config()

		if err != nil {
			return err
		}

		if config.StartType == mgr.StartAutomatic {
			return nil
		}

		config.StartType = mgr.StartAutomatic

		return service.UpdateConfig(config)
	})
}

// Disable stops the service from starting at boot.
func (m *Manager) Disable(_ context.Context, unit string) error {

	return m.withService(unit, func(service *mgr.Service) error {

		config, err := service.Config()

		if err != nil {
			return err
		}

		config.StartType = mgr.StartDisabled

		return service.UpdateConfig(config)
	})
}

// Start starts the service, and is a no-op when it is already running.
func (m *Manager) Start(ctx context.Context, unit string) error {

	return m.withService(unit, func(service *mgr.Service) error {

		if err := service.Start(); err != nil {

			if errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
				return nil
			}

			return fmt.Errorf("failed to start %s: %w", unit, elevationAware(err))
		}

		return await(ctx, service, svc.Running)
	})
}

// Stop stops the service, and is a no-op when it is already stopped.
func (m *Manager) Stop(ctx context.Context, unit string) error {

	return m.withService(unit, func(service *mgr.Service) error {
		return stop(ctx, service)
	})
}

// Restart restarts the service, starting it if it is not running.
//
// The SCM has no restart verb, and the two halves cannot simply be issued
// back to back: StartService fails while the previous process is still
// shutting down. The stop is therefore waited out before the start.
func (m *Manager) Restart(ctx context.Context, unit string) error {

	return m.withService(unit, func(service *mgr.Service) error {

		if err := stop(ctx, service); err != nil {
			return err
		}

		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start %s: %w", unit, elevationAware(err))
		}

		return await(ctx, service, svc.Running)
	})
}

// IsActive reports whether the service is running.
//
// It also records the process ID behind the answer, which is what lets
// Restarts detect a crash loop — see there.
func (m *Manager) IsActive(_ context.Context, unit string) (bool, error) {

	running := false

	err := m.withService(unit, func(service *mgr.Service) error {

		status, err := service.Query()

		if err != nil {
			return err
		}

		m.observe(unit, status.ProcessId)

		running = status.State == svc.Running

		return nil
	})

	if err != nil {

		// A service that does not exist is not running. That is an answer,
		// not a failure, and matches what systemctl is-active reports.
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return false, nil
		}

		return false, err
	}

	return running, nil
}

// Restarts reports how many times the service has restarted under
// observation.
//
// systemd keeps an NRestarts counter. Windows keeps a failure count, but only
// to decide which recovery action to run next: it is reset by the reset
// period, is not exposed by any documented query, and says nothing about
// restarts the SCM did not itself cause. There is no counter to read.
//
// What is observable is the process ID: when Windows restarts a service the
// new process gets a new one. This counts the changes seen across the calls
// to IsActive that verification already makes — one before the settle window
// and one after — so an agent that dies and is restarted inside that window
// is reported as restarting, which is the case the counter exists to catch.
//
// It cannot see a restart that happens entirely between two observations, and
// it deliberately reports nothing at all for a service nobody has looked at.
// Verification treats that the same way it treats NRestarts: a non-zero count
// fails the install, and zero is only ever corroborating evidence alongside
// the service being active and its log showing a connection.
func (m *Manager) Restarts(_ context.Context, unit string) (int, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.restarts[unit], nil
}

// Logs returns the most recent lines the agent wrote.
//
// A Windows service has no console and there is no journal, so the agent
// writes to a rotating file instead and this reads it back. The Event Log
// holds only lifecycle transitions — started, stopped, exited — and never the
// connection markers verification looks for, so it is not consulted here.
func (m *Manager) Logs(_ context.Context, unit string, lines int) (string, error) {

	if lines <= 0 {
		lines = 50
	}

	path := m.logPath()

	content, err := tailFile(path, lines)

	if err != nil {
		return "", fmt.Errorf("failed to read %s for %s: %w", path, unit, err)
	}

	return content, nil
}

// RemoveUnit deletes the service. It is not part of the installer's
// interface, and exists for the uninstall path.
func (m *Manager) RemoveUnit(unit string) error {

	err := m.withService(unit, func(service *mgr.Service) error {
		return service.Delete()
	})

	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return nil
	}

	return err
}

func (m *Manager) logPath() string {

	if m.LogPath != "" {
		return m.LogPath
	}

	return DefaultLogPath()
}

// observe records a process ID and counts a change as a restart. A zero ID
// means the service is not running and carries no information: a stop is not
// a restart, and the next non-zero ID is compared against the last real one.
func (m *Manager) observe(unit string, pid uint32) {

	if pid == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	previous, seen := m.observed[unit]

	if seen && previous != pid {
		m.restarts[unit]++
	}

	m.observed[unit] = pid
}

// withService opens the service, runs an action against it and closes it.
func (m *Manager) withService(unit string, action func(*mgr.Service) error) error {

	manager, err := connect()

	if err != nil {
		return err
	}

	defer manager.Disconnect()

	service, err := manager.OpenService(unit)

	if err != nil {

		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return err
		}

		return fmt.Errorf("failed to open the %s service: %w", unit, elevationAware(err))
	}

	defer service.Close()

	return action(service)
}

// stop asks the service to stop and waits for it to.
func stop(ctx context.Context, service *mgr.Service) error {

	if _, err := service.Control(svc.Stop); err != nil {

		// Already stopped, or stopped between the query and the request.
		if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
			return nil
		}

		return fmt.Errorf("failed to stop %s: %w", service.Name, elevationAware(err))
	}

	return await(ctx, service, svc.Stopped)
}

// await polls until the service reaches state, or the timeout passes.
func await(ctx context.Context, service *mgr.Service, state svc.State) error {

	deadline := time.Now().Add(stateTimeout)

	for {

		status, err := service.Query()

		if err != nil {
			return err
		}

		if status.State == state {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf(
				"%s did not reach the %s state within %s",
				service.Name,
				describe(state),
				stateTimeout,
			)
		}

		timer := time.NewTimer(statePoll)

		select {

		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()

		case <-timer.C:
		}
	}
}

func describe(state svc.State) string {

	switch state {

	case svc.Running:
		return "running"

	case svc.Stopped:
		return "stopped"

	default:
		return fmt.Sprintf("%d", state)
	}
}

// connect opens the service database for writing.
func connect() (*mgr.Mgr, error) {

	manager, err := mgr.Connect()

	if err != nil {
		return nil, fmt.Errorf("cannot reach the Service Control Manager: %w", elevationAware(err))
	}

	return manager, nil
}

// elevationAware turns the SCM's bare "Access is denied." into something that
// says what to do about it. Validation checks elevation up front, so this
// covers the narrower case of a process that is elevated but still refused —
// a locked-down service ACL, or Group Policy.
func elevationAware(err error) error {

	if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return fmt.Errorf("%w (run this command from an Administrator prompt)", err)
	}

	return err
}
