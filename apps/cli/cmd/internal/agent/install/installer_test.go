package install

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/progress"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/release"
)

// fakeUnits stands in for systemd.
type fakeUnits struct {
	unitDir  string
	written  map[string]string
	reloaded int
	enabled  []string
	restarts []string

	active      bool
	activeAfter int // number of IsActive calls before reporting active
	calls       int
	restartLoop int
	logs        string

	writeErr   error
	restartErr error
}

func newFakeUnits(t *testing.T) *fakeUnits {
	t.Helper()

	return &fakeUnits{
		unitDir: t.TempDir(),
		written: map[string]string{},
		active:  true,
		logs:    "level=INFO msg=\"websocket connected\" url=wss://platform.example.com/agents\n",
	}
}

func (f *fakeUnits) UnitPath(unit string) string { return filepath.Join(f.unitDir, unit) }

func (f *fakeUnits) WriteUnit(unit string, content string) (bool, error) {

	if f.writeErr != nil {
		return false, f.writeErr
	}

	changed := f.written[unit] != content
	f.written[unit] = content

	return changed, nil
}

func (f *fakeUnits) DaemonReload(context.Context) error { f.reloaded++; return nil }

func (f *fakeUnits) Enable(_ context.Context, unit string) error {
	f.enabled = append(f.enabled, unit)
	return nil
}

func (f *fakeUnits) Restart(_ context.Context, unit string) error {

	if f.restartErr != nil {
		return f.restartErr
	}

	f.restarts = append(f.restarts, unit)

	return nil
}

func (f *fakeUnits) IsActive(context.Context, string) (bool, error) {

	f.calls++

	if f.calls <= f.activeAfter {
		return false, nil
	}

	return f.active, nil
}

func (f *fakeUnits) Restarts(context.Context, string) (int, error) { return f.restartLoop, nil }

func (f *fakeUnits) Logs(context.Context, string, int) (string, error) { return f.logs, nil }

// requestedPaths records the API paths the installer asked for. Tests run
// sequentially and every request completes before Install returns, so a
// package-level recorder is safe here.
var requestedPaths []string

func requestedPath(t *testing.T, path string) bool {
	t.Helper()

	for _, requested := range requestedPaths {

		// Recorded paths carry the repo prefix, e.g.
		// /repos/Open-Source-Kigali/docksight/releases/latest
		if strings.HasSuffix(requested, path) {
			return true
		}
	}

	return false
}

// releaseStub serves a release with an agent binary attached.
func releaseStub(t *testing.T, version string, withAgent bool) string {
	t.Helper()

	requestedPaths = nil

	mux := http.NewServeMux()

	var serverURL string

	mux.HandleFunc("/assets/agent", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("#!/bin/sh\necho docksight-agent " + version + "\n"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		requestedPaths = append(requestedPaths, r.URL.Path)

		assets := fmt.Sprintf(
			`{"name": %q, "browser_download_url": %q}`,
			release.PlatformBundleName(version),
			serverURL+"/assets/platform",
		)

		if withAgent {
			assets += fmt.Sprintf(
				`,{"name": %q, "browser_download_url": %q}`,
				release.AgentBinaryName(version, release.CurrentTarget()),
				serverURL+"/assets/agent",
			)
		}

		fmt.Fprintf(w, `{"tag_name": %q, "assets": [%s]}`, version, assets)
	})

	server := httptest.NewServer(mux)
	serverURL = server.URL

	t.Cleanup(server.Close)

	return server.URL
}

type recordingReporter struct{ messages []string }

func (r *recordingReporter) Step(m string)       { r.messages = append(r.messages, m) }
func (r *recordingReporter) Success(m string)    { r.messages = append(r.messages, m) }
func (r *recordingReporter) Warn(m string)       { r.messages = append(r.messages, "warn: "+m) }
func (r *recordingReporter) Progress() io.Writer { return io.Discard }

func (r *recordingReporter) contains(fragment string) bool {

	for _, message := range r.messages {
		if strings.Contains(message, fragment) {
			return true
		}
	}

	return false
}

func harness(t *testing.T, withAgent bool) (*Installer, *fakeUnits, *recordingReporter) {
	t.Helper()

	root := t.TempDir()

	layout := DefaultLayout()
	layout.BinaryPath = filepath.Join(root, "usr", "local", "bin", "docksight-agent")
	layout.ConfigDir = filepath.Join(root, "etc", "docksight-agent")

	units := newFakeUnits(t)
	reporter := &recordingReporter{}

	installer := New(layout, "wss://platform.example.com/agents", reporter)
	installer.Releases.BaseURL = releaseStub(t, "v0.0.9", withAgent)
	installer.Units = units
	installer.SkipValidation = true
	installer.Settle = 10 * time.Millisecond

	return installer, units, reporter
}

func TestInstall(t *testing.T) {

	installer, units, reporter := harness(t, true)

	health, err := installer.Install(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	// Phase 3 — the binary is installed and executable.
	info, err := os.Stat(installer.Layout.BinaryPath)

	if err != nil {
		t.Fatalf("agent binary not installed: %v", err)
	}

	if info.Size() == 0 {
		t.Error("agent binary is empty")
	}

	// Phase 4 — configuration.
	config, err := os.ReadFile(installer.Layout.ConfigPath())

	if err != nil {
		t.Fatalf("config not written: %v", err)
	}

	if !strings.Contains(string(config), "wss://platform.example.com/agents") {
		t.Errorf("config does not carry the platform URL:\n%s", string(config))
	}

	// Phase 5 — the service is created, enabled and started.
	unit := units.written[installer.Layout.ServiceName]

	if unit == "" {
		t.Fatal("no unit file written")
	}

	for _, required := range []string{
		"Restart=always",
		"After=docker.service",
		"network.target",
		installer.Layout.BinaryPath,
		installer.Layout.ConfigPath(),
	} {

		if !strings.Contains(unit, required) {
			t.Errorf("unit is missing %q:\n%s", required, unit)
		}
	}

	if units.reloaded == 0 {
		t.Error("systemd was not reloaded")
	}

	if len(units.enabled) != 1 {
		t.Errorf("service enabled %d times", len(units.enabled))
	}

	if len(units.restarts) != 1 {
		t.Errorf("service started %d times", len(units.restarts))
	}

	// Phase 6 — verification saw the agent connect.
	if !health.Active {
		t.Error("health reports the service as inactive")
	}

	if !health.Connected {
		t.Errorf("health did not detect the connection, logs: %q", health.Logs)
	}

	if !reporter.contains("Agent connected") {
		t.Error("the connection was not reported to the user")
	}
}

// A release without an agent build must fail before touching the host.
func TestInstallWithoutAgentAsset(t *testing.T) {

	installer, units, _ := harness(t, false)

	_, err := installer.Install(context.Background())

	if err == nil {
		t.Fatal("expected an error when no agent binary is published")
	}

	if !strings.Contains(err.Error(), "docksight-agent-") {
		t.Errorf("error does not name the missing asset: %v", err)
	}

	if _, statErr := os.Stat(installer.Layout.ConfigPath()); statErr == nil {
		t.Error("the host was configured despite the download failing")
	}

	if len(units.restarts) != 0 {
		t.Error("the service was started despite the download failing")
	}
}

// Running the installer twice must upgrade in place, not duplicate state.
func TestInstallIsIdempotent(t *testing.T) {

	installer, units, _ := harness(t, true)

	ctx := context.Background()

	if _, err := installer.Install(ctx); err != nil {
		t.Fatal(err)
	}

	// The platform issues an identity on first connection.
	identity := []byte(`{"id":"11111111-2222-3333-4444-555555555555"}`)

	if err := os.WriteFile(installer.Layout.IdentityPath(), identity, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := installer.Install(ctx); err != nil {
		t.Fatal(err)
	}

	// The identity issued by the platform survives: overwriting it would
	// register this host a second time and orphan its history.
	preserved, err := os.ReadFile(installer.Layout.IdentityPath())

	if err != nil {
		t.Fatal(err)
	}

	if string(preserved) != string(identity) {
		t.Fatalf("identity was replaced: %s", string(preserved))
	}

	if len(units.enabled) != 2 || len(units.restarts) != 2 {
		t.Errorf("second install did not re-apply the service: %+v", units)
	}
}

// A changed platform URL must reach the config on reinstall.
func TestInstallUpdatesPlatformURL(t *testing.T) {

	installer, _, _ := harness(t, true)

	ctx := context.Background()

	if _, err := installer.Install(ctx); err != nil {
		t.Fatal(err)
	}

	installer.ServerURL = "wss://new-platform.example.com/agents"

	if _, err := installer.Install(ctx); err != nil {
		t.Fatal(err)
	}

	config, err := os.ReadFile(installer.Layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(config), "new-platform.example.com") {
		t.Fatalf("the new platform URL was not written:\n%s", string(config))
	}
}

// A service that starts and immediately dies must be reported, not celebrated.
func TestInstallDetectsCrashLoop(t *testing.T) {

	installer, units, _ := harness(t, true)

	units.restartLoop = 3
	units.logs = `level=INFO msg="connecting to docksight server"
level=WARN msg="agent session ended; reconnecting" error="websocket dial: connection refused"`

	health, err := installer.Install(context.Background())

	if err == nil {
		t.Fatal("a crash-looping service was reported as a success")
	}

	if health == nil || health.Connected {
		t.Error("a crash loop was reported as connected")
	}

	var phase *PhaseError

	if !asPhaseError(err, &phase) {
		t.Fatalf("error is not a PhaseError: %T", err)
	}

	if phase.Phase != PhaseVerify {
		t.Errorf("failure attributed to %q", phase.Phase)
	}

	// The diagnostic must carry the logs that explain the failure.
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("diagnostic does not include the cause:\n%v", err)
	}
}

// A service that never comes up at all.
func TestInstallDetectsDeadService(t *testing.T) {

	installer, units, _ := harness(t, true)

	units.active = false
	units.logs = "level=ERROR msg=\"invalid server url\"\n"

	_, err := installer.Install(context.Background())

	if err == nil {
		t.Fatal("a dead service was reported as a success")
	}

	if !strings.Contains(err.Error(), "journalctl") {
		t.Errorf("diagnostic gives no next step:\n%v", err)
	}
}

// Running but with no connection line yet is a warning, not a failure.
func TestInstallReportsUnconfirmedConnection(t *testing.T) {

	installer, units, reporter := harness(t, true)

	units.logs = "level=INFO msg=\"connecting to docksight server\"\n"

	health, err := installer.Install(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if health.Connected {
		t.Error("a connection was claimed without evidence")
	}

	if !reporter.contains("warn:") {
		t.Error("the unconfirmed connection was not surfaced")
	}
}

// Pinning must ask GitHub for that tag, not for the latest release.
func TestInstallPinnedVersion(t *testing.T) {

	installer, _, reporter := harness(t, true)

	installer.Version = "v0.0.5"

	if _, err := installer.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !requestedPath(t, "/releases/tags/v0.0.5") {
		t.Fatalf("the pinned tag was not requested, paths: %v", requestedPaths)
	}

	if requestedPath(t, "/releases/latest") {
		t.Error("the latest release was requested despite a pinned version")
	}

	if !reporter.contains("Fetching release v0.0.5") {
		t.Error("the pinned version was not reported")
	}
}

// Without a pin, the latest release is used.
func TestInstallUsesLatestByDefault(t *testing.T) {

	installer, _, _ := harness(t, true)

	if _, err := installer.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !requestedPath(t, "/releases/latest") {
		t.Fatalf("the latest release was not requested, paths: %v", requestedPaths)
	}
}

func TestInstallRequiresServerURL(t *testing.T) {

	installer, _, _ := harness(t, true)

	installer.ServerURL = ""

	if _, err := installer.Install(context.Background()); err == nil {
		t.Fatal("expected an error with no platform URL")
	}
}

func TestNewUsesDiscardReporter(t *testing.T) {

	installer := New(DefaultLayout(), "wss://example.com/agents", nil)

	if _, ok := installer.Report.(progress.Discard); !ok {
		t.Fatalf("reporter is %T, want progress.Discard", installer.Report)
	}
}
