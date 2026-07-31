package install

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Verification tuning. The settle window exists because systemctl start
// returns as soon as the process is forked: a service that dies on a bad
// config is still "active" for a moment afterwards.
const (
	defaultSettle   = 5 * time.Second
	defaultPoll     = time.Second
	defaultLogLines = 40
)

// Connection markers, taken from the agent's own log output in
// apps/agent/internal/communication/client.go. Matching real strings keeps
// verification honest: if those messages change, this reports "unconfirmed"
// rather than inventing a success.
var (
	connectedMarkers = []string{
		"websocket connected",
		"registration acknowledged",
		"registration successful",
	}

	failureMarkers = []string{
		"websocket dial",
		"registration failed",
		"invalid server url",
		"agent session ended",
		"connection refused",
		"no such host",
		"i/o timeout",
	}
)

// Health is the outcome of verifying an installation.
type Health struct {
	// Active is whether systemd reports the service as running.
	Active bool

	// Restarts is how many times systemd has had to restart it.
	Restarts int

	// Connected is whether the agent reported reaching the platform.
	Connected bool

	// Detail is the log line that decided Connected, when there is one.
	Detail string

	// Logs is the recent journal output, for diagnostics.
	Logs string
}

// verify waits for the service to settle and reports what it observes.
//
// A service can be "active" and still be broken: systemd restarts it, it
// dies again, and the loop repeats. Verification therefore checks liveness
// twice with a wait in between, and reads the journal to find out whether the
// agent actually reached the platform.
func (i *Installer) verify(ctx context.Context) (*Health, error) {

	unit := i.Layout.ServiceName

	active, err := i.waitActive(ctx)

	if err != nil {
		return nil, err
	}

	health := &Health{Active: active}

	health.Logs, _ = i.Units.Logs(ctx, unit, defaultLogLines)

	if !active {
		return health, &PhaseError{
			Phase: PhaseVerify,
			Err:   fmt.Errorf("%s is not running", unit),
			Hint:  "Inspect the service with: journalctl -u " + unit + " -n 50",
			Logs:  health.Logs,
		}
	}

	// Give the agent a moment to dial the platform, then look again.
	if err := sleep(ctx, i.settle()); err != nil {
		return health, err
	}

	health.Active, _ = i.Units.IsActive(ctx, unit)
	health.Restarts, _ = i.Units.Restarts(ctx, unit)
	health.Logs, _ = i.Units.Logs(ctx, unit, defaultLogLines)

	connected, detail := inspectLogs(health.Logs)

	health.Connected = connected
	health.Detail = detail

	if !health.Active || health.Restarts > 0 {
		return health, &PhaseError{
			Phase: PhaseVerify,
			Err: fmt.Errorf(
				"%s is restarting (%d restarts), the agent is not staying up",
				unit,
				health.Restarts,
			),
			Hint: "The usual cause is an unreachable platform URL in " + i.Layout.ConfigPath(),
			Logs: health.Logs,
		}
	}

	return health, nil
}

// waitActive polls until the unit reports active or the settle window passes.
func (i *Installer) waitActive(ctx context.Context) (bool, error) {

	deadline := time.Now().Add(i.settle())

	for {

		active, err := i.Units.IsActive(ctx, i.Layout.ServiceName)

		if err != nil {
			return false, err
		}

		if active {
			return true, nil
		}

		if time.Now().After(deadline) {
			return false, nil
		}

		if err := sleep(ctx, defaultPoll); err != nil {
			return false, err
		}
	}
}

// inspectLogs decides whether the agent reached the platform. A failure
// marker seen after a success marker wins: the agent connected and then lost
// the connection, which is not a healthy install.
func inspectLogs(logs string) (bool, string) {

	connected := false
	detail := ""

	for _, line := range strings.Split(logs, "\n") {

		lower := strings.ToLower(line)

		if matchesAny(lower, failureMarkers) {
			return false, strings.TrimSpace(line)
		}

		if !connected && matchesAny(lower, connectedMarkers) {
			connected = true
			detail = strings.TrimSpace(line)
		}
	}

	return connected, detail
}

func matchesAny(line string, markers []string) bool {

	for _, marker := range markers {

		if strings.Contains(line, marker) {
			return true
		}
	}

	return false
}

func (i *Installer) settle() time.Duration {

	if i.Settle <= 0 {
		return defaultSettle
	}

	return i.Settle
}

// sleep waits, but gives up if the command is cancelled.
func sleep(ctx context.Context, duration time.Duration) error {

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {

	case <-ctx.Done():
		return ctx.Err()

	case <-timer.C:
		return nil
	}
}
