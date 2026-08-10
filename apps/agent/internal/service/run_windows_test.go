//go:build windows

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sys/windows/svc"
)

// drive runs the handler against channels standing in for the Service Control
// Manager, and collects the statuses it reports.
func drive(
	t *testing.T,
	run func(context.Context) error,
	send func(chan<- svc.ChangeRequest),
) ([]svc.Status, uint32) {

	t.Helper()

	requests := make(chan svc.ChangeRequest)
	statuses := make(chan svc.Status, 16)

	type outcome struct {
		code uint32
	}

	done := make(chan outcome, 1)

	go func() {
		_, code := (&handler{run: run}).Execute(nil, requests, statuses)
		done <- outcome{code: code}
	}()

	send(requests)

	var result outcome

	select {
	case result = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return")
	}

	close(statuses)

	var reported []svc.Status

	for status := range statuses {
		reported = append(reported, status)
	}

	return reported, result.code
}

func states(reported []svc.Status) []svc.State {

	var out []svc.State

	for _, status := range reported {
		out = append(out, status.State)
	}

	return out
}

// SERVICE_CONTROL_STOP must cancel the agent's context and wait for it to
// finish before reporting Stopped. Returning early is what makes a service
// look like it crashed.
func TestStopCancelsAgentAndExitsCleanly(t *testing.T) {

	cancelled := make(chan struct{})

	run := func(ctx context.Context) error {
		<-ctx.Done()
		close(cancelled)
		return nil
	}

	reported, code := drive(t, run, func(requests chan<- svc.ChangeRequest) {
		requests <- svc.ChangeRequest{Cmd: svc.Stop}
	})

	select {
	case <-cancelled:
	default:
		t.Fatal("stop did not cancel the agent context")
	}

	if code != 0 {
		t.Errorf("exit code %d, want 0 for a clean stop", code)
	}

	got := states(reported)

	if len(got) < 3 || got[0] != svc.StartPending || got[1] != svc.Running {
		t.Fatalf("status sequence %v, want StartPending then Running", got)
	}

	if got[len(got)-1] != svc.StopPending {
		t.Errorf("last reported state %v, want StopPending", got[len(got)-1])
	}
}

// A reboot sends SHUTDOWN rather than STOP. Without AcceptShutdown the agent
// would be terminated instead of stopped, and the shutdown hooks never run.
func TestShutdownIsAcceptedLikeStop(t *testing.T) {

	cancelled := make(chan struct{})

	run := func(ctx context.Context) error {
		<-ctx.Done()
		close(cancelled)
		return nil
	}

	reported, code := drive(t, run, func(requests chan<- svc.ChangeRequest) {
		requests <- svc.ChangeRequest{Cmd: svc.Shutdown}
	})

	select {
	case <-cancelled:
	default:
		t.Fatal("shutdown did not cancel the agent context")
	}

	if code != 0 {
		t.Errorf("exit code %d, want 0", code)
	}

	if accepts := reported[1].Accepts; accepts&svc.AcceptShutdown == 0 {
		t.Error("handler does not accept SERVICE_CONTROL_SHUTDOWN")
	}

	if accepts := reported[1].Accepts; accepts&svc.AcceptStop == 0 {
		t.Error("handler does not accept SERVICE_CONTROL_STOP")
	}
}

// Interrogate is the SCM's health poll. An unanswered poll makes the service
// look unresponsive, so it must be echoed back and must not stop the agent.
func TestInterrogateIsAnsweredWithoutStopping(t *testing.T) {

	run := func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}

	current := svc.Status{State: svc.Running, Accepts: svc.AcceptStop}

	reported, code := drive(t, run, func(requests chan<- svc.ChangeRequest) {
		requests <- svc.ChangeRequest{Cmd: svc.Interrogate, CurrentStatus: current}
		requests <- svc.ChangeRequest{Cmd: svc.Stop}
	})

	if code != 0 {
		t.Errorf("exit code %d, want 0", code)
	}

	// StartPending, Running, the echoed Interrogate reply, then StopPending.
	if len(reported) != 4 {
		t.Fatalf("reported %d statuses (%v), want 4", len(reported), states(reported))
	}

	if reported[2].State != svc.Running {
		t.Errorf("interrogate answered with %v, want the current status echoed", reported[2].State)
	}
}

// An agent that dies on its own — bad config, unusable Docker endpoint — must
// report failure so configured recovery actions see it, rather than looking
// like an administrator stopped the service.
func TestAgentFailureReportsNonZeroExit(t *testing.T) {

	run := func(context.Context) error {
		return errors.New("docker endpoint unreachable")
	}

	_, code := drive(t, run, func(chan<- svc.ChangeRequest) {})

	if code == 0 {
		t.Fatal("agent failure reported a clean exit")
	}
}

// The agent returning nil unprompted is a normal exit, not a failure.
func TestAgentCleanExitReportsZero(t *testing.T) {

	run := func(context.Context) error { return nil }

	_, code := drive(t, run, func(chan<- svc.ChangeRequest) {})

	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
}

// A control code the handler never registered for must be ignored rather than
// treated as a stop.
func TestUnexpectedControlCodeIsIgnored(t *testing.T) {

	run := func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}

	reported, code := drive(t, run, func(requests chan<- svc.ChangeRequest) {
		requests <- svc.ChangeRequest{Cmd: svc.Pause}
		requests <- svc.ChangeRequest{Cmd: svc.Stop}
	})

	if code != 0 {
		t.Errorf("exit code %d, want 0", code)
	}

	if len(reported) != 3 {
		t.Fatalf("pause changed the reported statuses: %v", states(reported))
	}
}

// A console run must not attempt SCM dispatch, or a developer's foreground run
// fails with "not started by the service control manager".
func TestIsServiceIsFalseInTests(t *testing.T) {

	if IsService() {
		t.Fatal("the test process is not a Windows service")
	}
}
