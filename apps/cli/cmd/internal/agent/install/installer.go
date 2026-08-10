// Package install installs the DockSight Agent on a remote Docker host.
//
// The agent is a separate deliverable from the platform: it runs beside a
// Docker Engine, connects outbound over WebSocket, and exposes no inbound
// ports. Installation is idempotent — running it again upgrades the binary
// and rewrites the configuration while preserving the identity the platform
// issued to this host.
package install

import (
	"context"
	"errors"
	"time"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/progress"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/release"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/system"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/systemd"
)

// UnitController is the service-manager surface the installer needs.
// Depending on an interface rather than a concrete manager is what keeps the
// phases testable on a machine with neither systemd nor an SCM, and it is
// what let Windows support be a second implementation rather than a rewrite.
//
// The vocabulary is systemd's because that is where it came from. Three
// methods have no literal Windows counterpart and are documented at
// scm.Manager: WriteUnit writes no file, DaemonReload does nothing, and
// Restarts counts something the SCM does not record. The interface was left
// as it is rather than generalised — a name that fits both platforms exactly
// would still describe two different mechanisms, and the honest explanation
// belongs at the implementation, not in a renamed method.
type UnitController interface {
	UnitPath(unit string) string
	WriteUnit(unit string, content string) (bool, error)
	DaemonReload(ctx context.Context) error
	Enable(ctx context.Context, unit string) error
	Start(ctx context.Context, unit string) error
	Stop(ctx context.Context, unit string) error
	Restart(ctx context.Context, unit string) error
	IsActive(ctx context.Context, unit string) (bool, error)
	Restarts(ctx context.Context, unit string) (int, error)
	Logs(ctx context.Context, unit string, lines int) (string, error)
}

// compile-time proof that the systemd manager satisfies the interface. The
// SCM manager is proved the same way in controller_windows.go, where it is
// the only file that can import it.
var _ UnitController = (*systemd.Manager)(nil)

// Installer performs the agent installation.
type Installer struct {
	// Layout is where the agent is installed.
	Layout Layout

	// ServerURL is the normalized WebSocket endpoint of the platform.
	ServerURL string

	Releases *release.Client
	Units    UnitController
	Report   progress.Reporter

	// Target selects the agent build. Defaults to this host.
	Target release.Target

	// Version pins the release to install. Empty means the latest published
	// release, which is what a first install wants; pinning matters when an
	// agent must match a platform that has not been updated yet, or when a
	// bad build has to be rolled back on one host without touching the rest.
	Version string

	// Settle is how long to wait for the service to prove it stays up.
	Settle time.Duration

	// SkipValidation bypasses phase 1. Intended for tests only.
	SkipValidation bool
}

// New returns an installer wired with the real GitHub client and the service
// manager this host runs.
func New(layout Layout, serverURL string, reporter progress.Reporter) *Installer {

	if reporter == nil {
		reporter = progress.Discard{}
	}

	return &Installer{
		Layout:    layout,
		ServerURL: serverURL,
		Releases:  release.NewClient(),
		Units:     NewController(layout),
		Report:    reporter,
		Target:    release.CurrentTarget(),
		Settle:    defaultSettle,
	}
}

// Install runs the six installation phases in order.
func (i *Installer) Install(ctx context.Context) (*Health, error) {

	if i.ServerURL == "" {
		return nil, fail(PhaseConfigure, errors.New("no platform URL configured"), "")
	}

	// Phase 1 — system validation
	if err := i.validate(ctx); err != nil {
		return nil, err
	}

	// Phase 2 — locate and download the release
	target, err := i.resolveRelease(ctx)

	if err != nil {
		return nil, err
	}

	// Phase 3 — install the binary
	if err := i.downloadBinary(ctx, target); err != nil {
		return nil, fail(PhaseDownload, err, "")
	}

	// Phase 4 — configuration
	if err := i.configure(); err != nil {
		return nil, err
	}

	// Phase 5 — register the service with the host's service manager
	if err := i.installService(ctx); err != nil {
		return nil, fail(PhaseService, err, system.ElevationHint())
	}

	// Phase 6 — verification
	health, err := i.verify(ctx)

	if err != nil {
		return health, err
	}

	i.reportHealth(health)

	return health, nil
}

func (i *Installer) validate(ctx context.Context) error {

	if i.SkipValidation {
		return nil
	}

	i.Report.Step("Validating the host")

	if err := system.Validate(ctx, i.Report, system.AgentRequirements()); err != nil {
		return fail(PhaseValidate, err, elevationHint(err))
	}

	return nil
}

// elevationHint adds the fix to a validation failure that is about
// privileges, and nothing to any other. Every other check fails for a reason
// that running as root or as Administrator does not address.
func elevationHint(err error) string {

	var notElevated *system.NotElevatedError

	if errors.As(err, &notElevated) {
		return system.ElevationHint()
	}

	return ""
}

// resolveRelease returns the release to install: the pinned one when Version
// is set, the latest published otherwise.
func (i *Installer) resolveRelease(ctx context.Context) (*release.Release, error) {

	if i.Version != "" {

		i.Report.Step("Fetching release " + i.Version)

		pinned, err := i.Releases.ByTag(ctx, i.Version)

		if err != nil {
			return nil, fail(
				PhaseDownload,
				err,
				"Check the tag exists: gh release list",
			)
		}

		i.Report.Success("Using release " + pinned.Version())

		return pinned, nil
	}

	i.Report.Step("Checking the latest DockSight release")

	latest, err := i.Releases.Latest(ctx)

	if err != nil {
		return nil, fail(PhaseDownload, err, "")
	}

	i.Report.Success("Latest release " + latest.Version())

	return latest, nil
}

// configure writes config.yaml and prepares the identity location.
func (i *Installer) configure() error {

	i.Report.Step("Writing " + i.Layout.ConfigPath())

	// Report before writing: after WriteConfig the previous state is gone.
	registered := IdentityExists(i.Layout)

	if err := WriteConfig(i.Layout, i.ServerURL); err != nil {
		return fail(PhaseConfigure, err, system.ElevationHint())
	}

	i.Report.Success("Configured platform " + i.ServerURL)

	if registered {
		i.Report.Success("Existing identity preserved at " + i.Layout.IdentityPath())
	} else {
		i.Report.Step("Identity will be created at " + i.Layout.IdentityPath() + " on first connection")
	}

	return nil
}

func (i *Installer) reportHealth(health *Health) {

	i.Report.Success(i.Layout.ServiceName + " is running")

	if health.Connected {
		i.Report.Success("Agent connected to " + i.ServerURL)
		return
	}

	// Running but unconfirmed is a real state, not a failure: the log may
	// simply not have the connection line yet on a slow link.
	i.Report.Warn(
		"The service is running but no connection to " + i.ServerURL +
			" was confirmed yet. Follow it with: " + i.Layout.LogsCommand(),
	)
}

func (i *Installer) target() release.Target {

	if i.Target.OS == "" || i.Target.Arch == "" {
		return release.CurrentTarget()
	}

	return i.Target
}

// asPhaseError is errors.As, kept here so errors.go does not need the import.
func asPhaseError(err error, target **PhaseError) bool {
	return errors.As(err, target)
}
