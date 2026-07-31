package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/compose"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
)

// fakeStack stands in for docker compose.
type fakeStack struct {
	started   int
	restarted int
	services  []compose.Service
	startErr  error
	readyErr  error
}

func (f *fakeStack) Start(ctx context.Context, progress io.Writer) error {
	f.started++
	return f.startErr
}

func (f *fakeStack) Restart(ctx context.Context, progress io.Writer) error {
	f.restarted++
	return f.startErr
}

func (f *fakeStack) WaitReady(ctx context.Context) ([]compose.Service, error) {
	return f.services, f.readyErr
}

// recordingReporter captures the messages the cmd layer would print.
type recordingReporter struct {
	messages []string
}

func (r *recordingReporter) Step(message string)    { r.messages = append(r.messages, "step: "+message) }
func (r *recordingReporter) Success(message string) { r.messages = append(r.messages, "ok: "+message) }
func (r *recordingReporter) Warn(message string)    { r.messages = append(r.messages, "warn: "+message) }
func (r *recordingReporter) Progress() io.Writer    { return io.Discard }

func (r *recordingReporter) contains(fragment string) bool {

	for _, message := range r.messages {
		if strings.Contains(message, fragment) {
			return true
		}
	}

	return false
}

// bundleArchive builds a platform bundle exactly as the release pipeline
// publishes it: a top-level bundle/ directory with the compose file, the env
// template and a VERSION marker.
func bundleArchive(t *testing.T, version string) []byte {
	t.Helper()

	var buffer bytes.Buffer

	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	files := map[string]string{
		"bundle/" + config.ComposeFileName: "services: {}\n",
		"bundle/" + config.EnvExampleFileName: "POSTGRES_DB=docksight\n" +
			"POSTGRES_PASSWORD=docksight-password\n" +
			"POSTGRES_PORT=5432\n" +
			"JWT_SECRET=shipped-secret\n",
		"bundle/VERSION":      version + "\n",
		"bundle/default.conf": "server {}\n",
	}

	for name, content := range files {

		header := &tar.Header{
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}

		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}

	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	return buffer.Bytes()
}

// githubStub serves a release and its assets.
type githubStub struct {
	server   *httptest.Server
	bundle   []byte
	cli      []byte
	version  string
	requests []string
}

func newGithubStub(t *testing.T, version string) *githubStub {
	t.Helper()

	stub := &githubStub{
		bundle:  bundleArchive(t, version),
		cli:     []byte("#!/bin/sh\necho docksight " + version + "\n"),
		version: version,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/assets/platform", func(w http.ResponseWriter, r *http.Request) {
		w.Write(stub.bundle)
	})

	mux.HandleFunc("/assets/cli", func(w http.ResponseWriter, r *http.Request) {
		w.Write(stub.cli)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		stub.requests = append(stub.requests, r.URL.Path)

		fmt.Fprintf(w, `{
			"tag_name": %q,
			"assets": [
				{"name": %q, "browser_download_url": %q},
				{"name": %q, "browser_download_url": %q}
			]
		}`,
			stub.version,
			release.PlatformBundleName(stub.version),
			stub.server.URL+"/assets/platform",
			release.CLIBinaryName(stub.version, release.CurrentTarget()),
			stub.server.URL+"/assets/cli",
		)
	})

	stub.server = httptest.NewServer(mux)

	t.Cleanup(stub.server.Close)

	return stub
}

// harness wires an installer against the stub, a temp filesystem and a fake
// stack — everything except the real network and real Docker.
func harness(t *testing.T, version string) (*Installer, *githubStub, *fakeStack, *recordingReporter) {
	t.Helper()

	root := t.TempDir()

	cfg := config.Config{
		InstallationDir: filepath.Join(root, "opt", "docksight"),
		DataDir:         filepath.Join(root, "var", "docksight"),
		BinaryPath:      filepath.Join(root, "usr", "local", "bin", "docksight"),
		Port:            2002,
	}

	stub := newGithubStub(t, version)

	stack := &fakeStack{
		services: []compose.Service{
			{Service: "postgres", State: "running", Health: "healthy"},
			{Service: "web", State: "running"},
		},
	}

	reporter := &recordingReporter{}

	install := New(cfg, reporter)
	install.Releases.BaseURL = stub.server.URL
	install.Stack = stack

	return install, stub, stack, reporter
}

func TestInstall(t *testing.T) {

	install, _, stack, reporter := harness(t, "v0.0.4")

	if err := install.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	cfg := install.Config

	// The CLI landed on PATH.
	if _, err := os.Stat(cfg.BinaryPath); err != nil {
		t.Fatalf("CLI not installed: %v", err)
	}

	// The platform landed in the install directory, without the bundle/
	// wrapper directory.
	for _, name := range []string{config.ComposeFileName, "VERSION", "default.conf"} {

		if _, err := os.Stat(filepath.Join(cfg.InstallationDir, name)); err != nil {
			t.Errorf("%s not installed: %v", name, err)
		}
	}

	// Runtime files the bundle only ships templates for.
	generated, err := os.ReadFile(cfg.EnvPath())

	if err != nil {
		t.Fatalf(".env not generated: %v", err)
	}

	if strings.Contains(string(generated), "shipped-secret") {
		t.Error("the shipped JWT secret was installed verbatim")
	}

	if !strings.Contains(string(generated), "POSTGRES_PORT=5432") {
		t.Error("non-secret values were not preserved")
	}

	// The data directory exists even though the bundle does not contain it.
	if _, err := os.Stat(cfg.DataDir); err != nil {
		t.Errorf("data directory not created: %v", err)
	}

	if stack.started != 1 {
		t.Errorf("stack started %d times, want 1", stack.started)
	}

	// The version was recorded for a future update.
	recorded, err := install.State()

	if err != nil {
		t.Fatal(err)
	}

	if recorded.PlatformVersion != "v0.0.4" {
		t.Errorf("platform version recorded as %q", recorded.PlatformVersion)
	}

	if recorded.InstalledAt.IsZero() || recorded.UpdatedAt.IsZero() {
		t.Error("install timestamps not recorded")
	}

	if !reporter.contains("Starting DockSight services") {
		t.Error("progress was not reported")
	}
}

// The bundle must never carry the CLI: installing the platform may not put a
// binary into the install directory.
func TestInstallKeepsDeliverablesSeparate(t *testing.T) {

	install, _, _, _ := harness(t, "v0.0.4")

	if err := install.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(install.Config.InstallationDir)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range entries {

		if item.Name() == "docksight" {
			t.Fatal("the CLI was installed into the platform directory")
		}
	}
}

// A second install must not rewrite credentials that existing data volumes
// were created with.
func TestInstallTwicePreservesCredentials(t *testing.T) {

	install, _, _, _ := harness(t, "v0.0.4")

	ctx := context.Background()

	if err := install.Install(ctx); err != nil {
		t.Fatal(err)
	}

	first, err := os.ReadFile(install.Config.EnvPath())

	if err != nil {
		t.Fatal(err)
	}

	if err := install.Install(ctx); err != nil {
		t.Fatal(err)
	}

	second, err := os.ReadFile(install.Config.EnvPath())

	if err != nil {
		t.Fatal(err)
	}

	if string(first) != string(second) {
		t.Fatal("reinstalling regenerated the credentials")
	}
}

func TestUpdatePlatformRecreatesStack(t *testing.T) {

	install, stub, stack, _ := harness(t, "v0.0.4")

	ctx := context.Background()

	if err := install.Install(ctx); err != nil {
		t.Fatal(err)
	}

	// A newer release appears.
	stub.version = "v0.0.5"
	stub.bundle = bundleArchive(t, "v0.0.5")

	latest, err := install.LatestRelease(ctx)

	if err != nil {
		t.Fatal(err)
	}

	if err := install.UpdatePlatform(ctx, latest); err != nil {
		t.Fatal(err)
	}

	version, err := os.ReadFile(filepath.Join(install.Config.InstallationDir, "VERSION"))

	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(version)) != "v0.0.5" {
		t.Errorf("bundle not replaced: %q", string(version))
	}

	if stack.restarted != 1 {
		t.Errorf("stack restarted %d times, want 1", stack.restarted)
	}

	recorded, err := install.State()

	if err != nil {
		t.Fatal(err)
	}

	if recorded.PlatformVersion != "v0.0.5" {
		t.Errorf("platform version recorded as %q", recorded.PlatformVersion)
	}
}

// Updating the CLI must not touch the platform or restart the stack.
func TestUpdateCLILeavesPlatformAlone(t *testing.T) {

	install, stub, stack, _ := harness(t, "v0.0.4")

	ctx := context.Background()

	if err := install.Install(ctx); err != nil {
		t.Fatal(err)
	}

	stub.version = "v0.0.5"
	stub.cli = []byte("#!/bin/sh\necho docksight v0.0.5\n")

	latest, err := install.LatestRelease(ctx)

	if err != nil {
		t.Fatal(err)
	}

	if err := install.UpdateCLI(ctx, latest); err != nil {
		t.Fatal(err)
	}

	installed, err := os.ReadFile(install.Config.BinaryPath)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(installed), "v0.0.5") {
		t.Errorf("CLI not replaced: %q", string(installed))
	}

	recorded, err := install.State()

	if err != nil {
		t.Fatal(err)
	}

	if recorded.CLIVersion != "v0.0.5" {
		t.Errorf("CLI version recorded as %q", recorded.CLIVersion)
	}

	if recorded.PlatformVersion != "v0.0.4" {
		t.Errorf("platform version changed to %q", recorded.PlatformVersion)
	}

	if stack.restarted != 0 || stack.started != 1 {
		t.Errorf("the stack was disturbed: started=%d restarted=%d", stack.started, stack.restarted)
	}
}

func TestInstallReportsMissingBundle(t *testing.T) {

	install, _, _, _ := harness(t, "v0.0.4")

	// A release that publishes only the CLI.
	install.Releases.BaseURL = emptyReleaseServer(t)

	err := install.Install(context.Background())

	if err == nil {
		t.Fatal("expected an error when no platform bundle is published")
	}

	if !strings.Contains(err.Error(), "docksight-platform-") {
		t.Fatalf("error does not name the missing asset: %v", err)
	}
}

func TestInstallSurfacesStackFailure(t *testing.T) {

	install, _, stack, _ := harness(t, "v0.0.4")

	stack.services = []compose.Service{
		{Service: "server", State: "exited", ExitCode: 3},
	}
	stack.readyErr = fmt.Errorf("service failed to start: server: exited (code 3)")

	err := install.Install(context.Background())

	if err == nil {
		t.Fatal("expected the stack failure to surface")
	}

	// The platform is still on disk: the failure is in the containers, not
	// the installation, so the user can inspect and retry.
	if _, statErr := os.Stat(install.Config.ComposePath()); statErr != nil {
		t.Errorf("install was rolled back unexpectedly: %v", statErr)
	}
}

func emptyReleaseServer(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"tag_name":"v0.0.4","assets":[{"name":"checksums.txt","browser_download_url":"https://example.com/x"}]}`)
		},
	))

	t.Cleanup(server.Close)

	return server.URL
}
