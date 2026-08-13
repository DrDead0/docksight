// Package installer orchestrates the two DockSight deliverables: the CLI
// binary on PATH and the platform bundle in the installation directory.
//
// Each deliverable is installed by its own method, so `docksight update` can
// refresh one without touching the other, and neither method knows how
// progress is displayed.
package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/config"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/buildinfo"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/compose"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/env"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/release"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/selfinstall"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/state"
)

// Timeouts. Starting the stack is dominated by image pulls on a first
// install; readiness is bounded by container healthchecks.
const (
	DefaultStartTimeout = 15 * time.Minute
	DefaultReadyTimeout = 5 * time.Minute
)

// Installer installs and updates DockSight. Construct it with New.
type Installer struct {
	Config   config.Config
	Releases *release.Client
	Stack    Stack
	Host     Host
	Report   Reporter

	// Target selects which CLI build to fetch when updating. Defaults to the
	// platform this binary runs on.
	Target release.Target

	StartTimeout time.Duration
	ReadyTimeout time.Duration
}

// New returns an installer wired with the real GitHub client and compose
// runner. Any field may be replaced afterwards, which is how tests inject
// fakes.
func New(cfg config.Config, reporter Reporter) *Installer {

	if reporter == nil {
		reporter = Discard{}
	}

	return &Installer{
		Config:       cfg,
		Releases:     release.NewClient(),
		Stack:        compose.NewRunner(cfg.InstallationDir, config.ComposeFileName),
		Host:         LocalHost{},
		Report:       reporter,
		Target:       release.CurrentTarget(),
		StartTimeout: DefaultStartTimeout,
		ReadyTimeout: DefaultReadyTimeout,
	}
}

// Install performs a complete first installation: it puts the running CLI on
// PATH, installs the platform bundle, starts the stack and records what was
// installed.
func (i *Installer) Install(ctx context.Context) error {

	if err := i.InstallCLI(ctx); err != nil {
		return err
	}

	latest, err := i.LatestRelease(ctx)

	if err != nil {
		return err
	}

	if err := i.InstallPlatform(ctx, latest); err != nil {
		return err
	}

	// Before the stack starts, so a host that refuses the rule fails while
	// nothing is listening rather than after the platform is up and quietly
	// unreachable.
	if err := i.openFirewall(ctx); err != nil {
		return err
	}

	if err := i.StartStack(ctx, false); err != nil {
		return err
	}

	return i.record(func(current *state.State) {
		current.PlatformVersion = latest.Version()
		current.CLIVersion = buildinfo.Version
	})
}

// LatestRelease fetches the newest published release.
func (i *Installer) LatestRelease(ctx context.Context) (*release.Release, error) {

	i.Report.Step("Checking the latest DockSight release")

	latest, err := i.Releases.Latest(ctx)

	if err != nil {
		return nil, err
	}

	i.Report.Success("Latest release " + latest.Version())

	return latest, nil
}

// InstallCLI copies the executable of the running process to the configured
// binary path. The process keeps running afterwards: the binary is replaced
// by renaming a new file over it, so this process retains its own open inode.
func (i *Installer) InstallCLI(ctx context.Context) error {

	i.Report.Step("Installing the CLI into " + i.Config.BinaryPath)

	source, err := selfinstall.Running()

	if err != nil {
		return err
	}

	installed, err := selfinstall.InstallRunning(i.Config.BinaryPath)

	if err != nil {
		return fmt.Errorf("failed to install the CLI: %w", err)
	}

	if !installed {
		i.Report.Success("CLI already installed at " + i.Config.BinaryPath)
	} else {
		i.Report.Success(fmt.Sprintf("CLI installed from %s", source))
	}

	return i.ensureOnPath()
}

// ensureOnPath makes the installed CLI resolvable by name.
//
// This runs even when the binary was already in place: the copy succeeding
// says nothing about whether its directory is searched, and an interrupted
// earlier install could have done one without the other.
func (i *Installer) ensureOnPath() error {

	directory := i.Config.InstallDirectory()

	added, err := i.host().EnsureOnPath(directory)

	if err != nil {
		return err
	}

	if added {
		i.Report.Success("Added " + directory + " to the system PATH")
		i.Report.Warn("Open a new terminal for `docksight` to resolve there")
	}

	return nil
}

// UpdateCLI replaces the installed CLI with the binary published in the given
// release. Unlike InstallCLI it downloads a fresh build rather than copying
// the running process, so it can move the CLI forward or back.
func (i *Installer) UpdateCLI(ctx context.Context, rel *release.Release) error {

	asset, err := rel.CLIBinary(i.target())

	if err != nil {
		return err
	}

	i.Report.Step("Downloading " + asset.Name)

	workspace, err := filesystem.TempWorkspace()

	if err != nil {
		return err
	}

	defer os.RemoveAll(workspace)

	downloaded := filepath.Join(workspace, asset.Name)

	if err := i.Releases.Download(ctx, *asset, downloaded); err != nil {
		return err
	}

	if err := i.verifyDownload(ctx, rel, asset.Name, downloaded); err != nil {
		return err
	}

	binary := downloaded

	if asset.IsArchive() {

		extracted := filepath.Join(workspace, "cli")

		if err := release.ExtractTarGz(downloaded, extracted); err != nil {
			return err
		}

		binary, err = findExecutable(extracted)

		if err != nil {
			return err
		}
	}

	if err := selfinstall.Install(binary, i.Config.BinaryPath); err != nil {
		return fmt.Errorf("failed to install the CLI: %w", err)
	}

	i.Report.Success("CLI updated to " + rel.Version())

	return i.record(func(current *state.State) {
		current.CLIVersion = rel.Version()
	})
}

// InstallPlatform downloads the platform bundle of a release and installs it
// into the installation directory, generating any missing runtime files.
func (i *Installer) InstallPlatform(ctx context.Context, rel *release.Release) error {

	asset, err := rel.PlatformBundle()

	if err != nil {
		return err
	}

	i.Report.Step("Downloading " + asset.Name)

	workspace, err := filesystem.TempWorkspace()

	if err != nil {
		return err
	}

	// The workspace only stages the bundle on its way to the install
	// directory.
	defer os.RemoveAll(workspace)

	archive := filepath.Join(workspace, asset.Name)

	if err := i.Releases.Download(ctx, *asset, archive); err != nil {
		return err
	}

	if err := i.verifyDownload(ctx, rel, asset.Name, archive); err != nil {
		return err
	}

	i.Report.Success("Downloaded " + asset.Name)

	i.Report.Step("Installing the platform into " + i.Config.InstallationDir)

	extracted := filepath.Join(workspace, "extracted")

	if err := release.ExtractTarGz(archive, extracted); err != nil {
		return err
	}

	if err := filesystem.CreateDirectories(
		i.Config.InstallationDir,
		i.Config.DataDir,
	); err != nil {
		return err
	}

	if err := filesystem.CopyDir(bundleRoot(extracted), i.Config.InstallationDir); err != nil {
		return err
	}

	i.Report.Success("Platform installed")

	if err := i.ensureRuntimeFiles(); err != nil {
		return err
	}

	return nil
}

// UpdatePlatform installs a new platform bundle over an existing
// installation and recreates the stack. The generated .env is preserved, so
// credentials keep matching the existing data volumes.
func (i *Installer) UpdatePlatform(ctx context.Context, rel *release.Release) error {

	if err := i.InstallPlatform(ctx, rel); err != nil {
		return err
	}

	if err := i.StartStack(ctx, true); err != nil {
		return err
	}

	return i.record(func(current *state.State) {
		current.PlatformVersion = rel.Version()
	})
}

// StartStack brings the compose stack up and waits until every service is
// ready. Pass recreate to replace running containers with new definitions.
func (i *Installer) StartStack(ctx context.Context, recreate bool) error {

	i.Report.Step("Starting DockSight services")

	startCtx, cancelStart := context.WithTimeout(ctx, i.startTimeout())
	defer cancelStart()

	start := i.Stack.Start

	if recreate {
		start = i.Stack.Restart
	}

	if err := start(startCtx, i.Report.Progress()); err != nil {
		return err
	}

	i.Report.Step("Waiting for services to become ready")

	readyCtx, cancelReady := context.WithTimeout(ctx, i.readyTimeout())
	defer cancelReady()

	services, err := i.Stack.WaitReady(readyCtx)

	for _, service := range services {

		if service.Ready() {
			i.Report.Success(service.Describe())
		} else {
			i.Report.Warn(service.Describe())
		}
	}

	return err
}

// verifyDownload checks SHA-256 against checksums.txt before the file is
// installed. Older releases have no checksums file; that is a warning, not
// a hard failure, so v0.1.0 still installs.
func (i *Installer) verifyDownload(ctx context.Context, rel *release.Release, name, path string) error {
	verified, err := i.Releases.VerifyDownloaded(ctx, rel, name, path)
	if err != nil {
		return err
	}
	if !verified {
		i.Report.Warn("release has no checksums.txt — skipped verification for " + name)
		return nil
	}
	i.Report.Success("Verified checksum for " + name)
	return nil
}

// State returns the recorded installation.
func (i *Installer) State() (state.State, error) {
	return state.Load(i.Config.StatePath())
}

// ensureRuntimeFiles creates files the bundle only ships templates for.
func (i *Installer) ensureRuntimeFiles() error {

	created, err := env.Generate(i.Config.EnvExamplePath(), i.Config.EnvPath())

	if err != nil {

		if os.IsNotExist(err) {
			// A bundle without a template has nothing to generate from; that
			// is a packaging choice, not an install failure.
			i.Report.Warn("No " + config.EnvExampleFileName + " in the bundle, skipping environment generation")
			return nil
		}

		return err
	}

	if created {
		i.Report.Success("Generated " + i.Config.EnvPath())
	} else {
		i.Report.Warn("Existing " + config.EnvFileName + " kept, credentials unchanged")
	}

	// Applied whether the file was just written or was already there. The
	// mode env.Generate passes to os.OpenFile means nothing on Windows, and
	// an installation made before this existed has a .env that every local
	// account can read — repairing it is the point of doing this on the
	// existing-file path too.
	if err := i.host().ProtectSecret(i.Config.EnvPath()); err != nil {
		return err
	}

	return nil
}

// openFirewall lets machines other than this one reach the platform.
//
// The agent model is outbound-only: agents dial the platform and the platform
// never dials them. That makes exactly one inbound port the difference
// between a working installation and one where every agent fails to connect
// for a reason that looks nothing like a firewall.
func (i *Installer) openFirewall(ctx context.Context) error {

	opened, err := i.host().AllowPort(ctx, i.Config.Port)

	if err != nil {
		return err
	}

	if opened {
		i.Report.Success(fmt.Sprintf("Opened TCP port %d in Windows Firewall", i.Config.Port))
	}

	return nil
}

// record updates the install state, preserving fields the caller does not set.
func (i *Installer) record(apply func(*state.State)) error {

	current, err := state.Load(i.Config.StatePath())

	if err != nil {
		return err
	}

	apply(&current)

	now := time.Now().UTC()

	if current.InstalledAt.IsZero() {
		current.InstalledAt = now
	}

	current.UpdatedAt = now

	return state.Save(i.Config.StatePath(), current)
}

// host returns the machine integration, defaulting to the real one. An
// Installer built as a literal rather than through New still works.
func (i *Installer) host() Host {

	if i.Host == nil {
		return LocalHost{}
	}

	return i.Host
}

func (i *Installer) target() release.Target {

	if i.Target.OS == "" || i.Target.Arch == "" {
		return release.CurrentTarget()
	}

	return i.Target
}

func (i *Installer) startTimeout() time.Duration {

	if i.StartTimeout <= 0 {
		return DefaultStartTimeout
	}

	return i.StartTimeout
}

func (i *Installer) readyTimeout() time.Duration {

	if i.ReadyTimeout <= 0 {
		return DefaultReadyTimeout
	}

	return i.ReadyTimeout
}

// bundleRoot returns the directory to copy from. Bundles are published with a
// top-level "bundle/" directory, but an archive without one is still usable.
func bundleRoot(extracted string) string {

	nested := filepath.Join(extracted, config.BundleRootDir)

	if info, err := os.Stat(nested); err == nil && info.IsDir() {
		return nested
	}

	return extracted
}

// findExecutable locates the single binary inside an extracted CLI archive.
func findExecutable(root string) (string, error) {

	entries, err := os.ReadDir(root)

	if err != nil {
		return "", err
	}

	for _, entry := range entries {

		if !entry.IsDir() {
			return filepath.Join(root, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no executable found in %s", root)
}
