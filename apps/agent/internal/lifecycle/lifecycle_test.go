package lifecycle

import (
	"context"
	"testing"
	"time"
)

// Cancelling the parent is the only way a Windows service can be stopped: the
// SCM delivers a control code, never a signal. If this stops working, the SCM
// would ask the agent to stop, get no response, and kill it — every stop
// looking like a crash.
func TestParentCancellationTriggersShutdown(t *testing.T) {
	t.Parallel()

	parent, cancel := context.WithCancel(context.Background())

	manager := New(parent)

	stopped := make(chan struct{})
	ranHook := false

	go func() {
		manager.Wait(func() { ranHook = true })
		close(stopped)
	}()

	cancel()

	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("cancelling the parent did not shut the lifecycle down")
	}

	if !ranHook {
		t.Error("shutdown hook did not run; the SCM stop path must reuse it")
	}
}

// A nil parent must not panic: it is the natural thing for a caller with no
// supervisor to pass.
func TestNilParentDefaultsToBackground(t *testing.T) {
	t.Parallel()

	manager := New(nil)

	select {
	case <-manager.Context().Done():
		t.Fatal("context is already cancelled")
	default:
	}

	manager.Shutdown()

	select {
	case <-manager.Context().Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown did not cancel the context")
	}
}

// Shutdown runs hooks exactly once however many times it is called. The
// service handler cancels the parent and Wait may also fire, so a double
// shutdown is a real path, not a hypothetical one.
func TestShutdownHooksRunOnce(t *testing.T) {
	t.Parallel()

	manager := New(context.Background())

	calls := 0

	manager.Shutdown(func() { calls++ })
	manager.Shutdown(func() { calls++ })

	if calls != 1 {
		t.Fatalf("hook ran %d times, want 1", calls)
	}
}
