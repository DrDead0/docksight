package systemd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeRunner records systemctl invocations and replays canned output.
type fakeRunner struct {
	calls   [][]string
	outputs map[string]string
	errs    map[string]error
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		outputs: map[string]string{},
		errs:    map[string]error{},
	}
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {

	call := append([]string{name}, args...)

	f.calls = append(f.calls, call)

	key := strings.Join(call, " ")

	return []byte(f.outputs[key]), f.errs[key]
}

func (f *fakeRunner) called(fragment string) bool {

	for _, call := range f.calls {

		if strings.Contains(strings.Join(call, " "), fragment) {
			return true
		}
	}

	return false
}

func testManager(t *testing.T) (*Manager, *fakeRunner) {
	t.Helper()

	runner := newFakeRunner()

	return &Manager{UnitDir: t.TempDir(), Runner: runner}, runner
}

func TestWriteUnit(t *testing.T) {

	manager, _ := testManager(t)

	changed, err := manager.WriteUnit("docksight-agent.service", "[Unit]\n")

	if err != nil {
		t.Fatal(err)
	}

	if !changed {
		t.Error("writing a new unit reported no change")
	}

	content, err := os.ReadFile(filepath.Join(manager.UnitDir, "docksight-agent.service"))

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "[Unit]\n" {
		t.Fatalf("got %q", string(content))
	}
}

// Rewriting identical content must not report a change: the caller uses this
// to decide whether a daemon-reload is needed.
func TestWriteUnitDetectsNoChange(t *testing.T) {

	manager, _ := testManager(t)

	if _, err := manager.WriteUnit("docksight-agent.service", "[Unit]\n"); err != nil {
		t.Fatal(err)
	}

	changed, err := manager.WriteUnit("docksight-agent.service", "[Unit]\n")

	if err != nil {
		t.Fatal(err)
	}

	if changed {
		t.Error("identical content reported as changed")
	}
}

func TestWriteUnitDetectsChange(t *testing.T) {

	manager, _ := testManager(t)

	if _, err := manager.WriteUnit("u.service", "old"); err != nil {
		t.Fatal(err)
	}

	changed, err := manager.WriteUnit("u.service", "new")

	if err != nil {
		t.Fatal(err)
	}

	if !changed {
		t.Error("changed content reported as unchanged")
	}
}

func TestLifecycleCommands(t *testing.T) {

	manager, runner := testManager(t)

	ctx := context.Background()

	if err := manager.DaemonReload(ctx); err != nil {
		t.Fatal(err)
	}

	if err := manager.Enable(ctx, "docksight-agent.service"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Restart(ctx, "docksight-agent.service"); err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"systemctl daemon-reload",
		"systemctl enable docksight-agent.service",
		"systemctl restart docksight-agent.service",
	} {

		if !runner.called(expected) {
			t.Errorf("%q was not run, calls: %v", expected, runner.calls)
		}
	}
}

func TestCommandFailureIncludesOutput(t *testing.T) {

	manager, runner := testManager(t)

	key := "systemctl enable missing.service"

	runner.errs[key] = errors.New("exit status 1")
	runner.outputs[key] = "Failed to enable unit: Unit missing.service does not exist."

	err := manager.Enable(context.Background(), "missing.service")

	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error loses systemctl's message: %v", err)
	}
}

// `systemctl is-active` exits non-zero for a stopped unit. That is an answer,
// not a failure, and must not surface as an error.
func TestIsActive(t *testing.T) {

	manager, runner := testManager(t)

	key := "systemctl is-active docksight-agent.service"

	runner.outputs[key] = "inactive\n"
	runner.errs[key] = errors.New("exit status 3")

	active, err := manager.IsActive(context.Background(), "docksight-agent.service")

	if err != nil {
		t.Fatalf("a stopped unit produced an error: %v", err)
	}

	if active {
		t.Error("an inactive unit reported as active")
	}

	runner.outputs[key] = "active\n"
	runner.errs[key] = nil

	active, err = manager.IsActive(context.Background(), "docksight-agent.service")

	if err != nil {
		t.Fatal(err)
	}

	if !active {
		t.Error("an active unit reported as inactive")
	}
}

func TestRestarts(t *testing.T) {

	manager, runner := testManager(t)

	key := "systemctl show docksight-agent.service --property NRestarts --value"

	runner.outputs[key] = "4\n"

	count, err := manager.Restarts(context.Background(), "docksight-agent.service")

	if err != nil {
		t.Fatal(err)
	}

	if count != 4 {
		t.Fatalf("got %d restarts, want 4", count)
	}
}

// Older systemd omits NRestarts entirely; that must read as zero, not fail.
func TestRestartsWithEmptyProperty(t *testing.T) {

	manager, runner := testManager(t)

	runner.outputs["systemctl show u.service --property NRestarts --value"] = "\n"

	count, err := manager.Restarts(context.Background(), "u.service")

	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("got %d, want 0", count)
	}
}

func TestLogs(t *testing.T) {

	manager, runner := testManager(t)

	key := "journalctl --unit docksight-agent.service --lines 10 --no-pager --output cat"

	runner.outputs[key] = "websocket connected\n"

	logs, err := manager.Logs(context.Background(), "docksight-agent.service", 10)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(logs, "websocket connected") {
		t.Fatalf("got %q", logs)
	}
}

func TestRemoveUnitIsIdempotent(t *testing.T) {

	manager, _ := testManager(t)

	if err := manager.RemoveUnit("absent.service"); err != nil {
		t.Fatalf("removing a missing unit failed: %v", err)
	}
}
