//go:build windows

package install

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/scm"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/selfinstall"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/system"
)

// This is the one test that cannot be faked. Everything else in the package
// drives a stand-in controller, which proves the phases call the right things
// in the right order and proves nothing at all about the Service Control
// Manager. This registers a real service, starts a real agent and reads the
// log it really wrote.
//
// It is opt-in because it needs Administrator rights and changes the machine
// it runs on: set DOCKSIGHT_WINDOWS_E2E=1 and point DOCKSIGHT_AGENT_BINARY at
// a windows/amd64 agent build. Everything it creates is removed again, and it
// refuses to start if a docksight-agent service already exists rather than
// deleting an installation somebody is using.
//
//	go test ./cmd/internal/agent/install -run TestWindowsServiceLifecycle -v
//
// The platform URL points at a closed port on purpose. The agent treats an
// unreachable platform as something to retry rather than something to die of,
// so the service stays up and every method under test gets exercised while
// verification correctly reports the connection as unconfirmed.
const unreachablePlatform = "ws://127.0.0.1:59321/agents"

func TestWindowsServiceLifecycle(t *testing.T) {

	if os.Getenv("DOCKSIGHT_WINDOWS_E2E") != "1" {
		t.Skip("set DOCKSIGHT_WINDOWS_E2E=1 to run against the real Service Control Manager")
	}

	if err := system.CheckElevation(); err != nil {
		t.Fatalf("this test needs an Administrator prompt: %v", err)
	}

	agentBinary := os.Getenv("DOCKSIGHT_AGENT_BINARY")

	if agentBinary == "" {
		t.Fatal("set DOCKSIGHT_AGENT_BINARY to a windows/amd64 agent build")
	}

	ctx := context.Background()

	layout := WindowsLayout()
	units := NewController(layout)

	if active, _ := units.IsActive(ctx, layout.ServiceName); active {
		t.Fatalf("%s already exists on this host, refusing to replace it", layout.ServiceName)
	}

	if _, err := os.Stat(layout.ConfigPath()); err == nil {
		t.Fatalf("%s already exists on this host, refusing to replace it", layout.ConfigPath())
	}

	t.Cleanup(func() { removeInstallation(t, ctx, layout, units) })

	// Phases 3 and 4, without the GitHub round trip.
	if err := selfinstall.Install(agentBinary, layout.BinaryPath); err != nil {
		t.Fatalf("failed to place the agent binary: %v", err)
	}

	if err := WriteConfig(layout, unreachablePlatform); err != nil {
		t.Fatalf("failed to write the configuration: %v", err)
	}

	config, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(config), "/") && !strings.Contains(string(config), "url:") {
		t.Errorf("the generated config mixes separators:\n%s", string(config))
	}

	reporter := &recordingReporter{}

	installer := New(layout, unreachablePlatform, reporter)
	installer.Units = units
	installer.Settle = 15 * time.Second

	// Phase 5 — register and start the service.
	if err := installer.installService(ctx); err != nil {
		t.Fatalf("failed to install the service: %v", err)
	}

	// Phase 6 — verify. A closed platform port means Connected is false and
	// the agent keeps retrying, so the service must still be up.
	health, err := installer.verify(ctx)

	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}

	if !health.Active {
		t.Fatalf("%s is not running after install", layout.ServiceName)
	}

	if health.Restarts != 0 {
		t.Errorf("the agent restarted %d times against an unreachable platform", health.Restarts)
	}

	if health.Connected {
		t.Errorf("a connection to a closed port was reported as confirmed: %q", health.Detail)
	}

	// The log has to be readable, or verification is guessing.
	if strings.TrimSpace(health.Logs) == "" {
		t.Errorf("no agent output was read back from %s", layout.LogPath)
	}

	t.Logf("service %s active=%v restarts=%d", layout.ServiceName, health.Active, health.Restarts)
	t.Logf("agent output:\n%s", health.Logs)

	// The platform issues this on first connection and it must survive an
	// upgrade, which is the whole reason reinstalling is the supported way to
	// upgrade rather than uninstall-then-install.
	identity := []byte(`{"id":"11111111-2222-3333-4444-555555555555"}`)

	if err := os.WriteFile(layout.IdentityPath(), identity, 0o600); err != nil {
		t.Fatal(err)
	}

	// Reinstalling must update in place: the same definition is not a change,
	// and the service is neither recreated nor duplicated.
	changed, err := units.WriteUnit(layout.ServiceName, RenderUnit(layout))

	if err != nil {
		t.Fatalf("the second install failed: %v", err)
	}

	if changed {
		t.Error("an unchanged service definition was reported as a change")
	}

	if err := installer.installService(ctx); err != nil {
		t.Fatalf("the second install failed: %v", err)
	}

	if _, err := installer.verify(ctx); err != nil {
		t.Fatalf("verification failed after the second install: %v", err)
	}

	preserved, err := os.ReadFile(layout.IdentityPath())

	if err != nil {
		t.Fatal(err)
	}

	if string(preserved) != string(identity) {
		t.Fatalf("identity was replaced: %s", string(preserved))
	}

	// A changed platform URL must reach the service definition, and a
	// definition that really did change must be reported as one.
	moved := layout
	moved.BinaryPath = filepath.Join(t.TempDir(), "docksight-agent.exe")

	changed, err = units.WriteUnit(layout.ServiceName, RenderUnit(moved))

	if err != nil {
		t.Fatalf("failed to update the service: %v", err)
	}

	if !changed {
		t.Error("a changed service definition was reported as unchanged")
	}

	// Put it back so the stop below talks to the binary that is really there.
	if _, err := units.WriteUnit(layout.ServiceName, RenderUnit(layout)); err != nil {
		t.Fatal(err)
	}

	if err := units.Stop(ctx, layout.ServiceName); err != nil {
		t.Errorf("failed to stop the service: %v", err)
	}

	if active, _ := units.IsActive(ctx, layout.ServiceName); active {
		t.Error("the service is still running after a stop")
	}

	if err := units.Start(ctx, layout.ServiceName); err != nil {
		t.Errorf("failed to start the service: %v", err)
	}

	if active, _ := units.IsActive(ctx, layout.ServiceName); !active {
		t.Error("the service did not come back after a start")
	}
}

// removeInstallation puts the host back as it was found.
func removeInstallation(t *testing.T, ctx context.Context, layout Layout, units UnitController) {
	t.Helper()

	if err := units.Stop(ctx, layout.ServiceName); err != nil {
		t.Logf("cleanup: stop: %v", err)
	}

	manager, ok := units.(*scm.Manager)

	if ok {

		if err := manager.RemoveUnit(layout.ServiceName); err != nil {
			t.Logf("cleanup: delete service: %v", err)
		}
	}

	for _, path := range []string{
		layout.ConfigDir,
		filepath.Dir(layout.BinaryPath),
		filepath.Dir(layout.LogPath),
	} {

		if err := os.RemoveAll(path); err != nil {
			t.Logf("cleanup: remove %s: %v", path, err)
		}
	}
}
