package cmd

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/config"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/installer"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/system"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var installCMD = &cobra.Command{
	Use:   "install",
	Short: "Install DockSight",
	Long: "Install the DockSight CLI onto PATH, install the platform bundle " +
		"and start the stack.",

	RunE: func(cmd *cobra.Command, args []string) error {

		ui.Banner()

		if err := system.Validate(
			cmd.Context(),
			consoleReporter{},
			system.PlatformRequirements(),
		); err != nil {
			return withElevationHint(err)
		}

		cfg := config.Default()

		showConfig(cfg)

		if err := installer.New(cfg, consoleReporter{}).Install(cmd.Context()); err != nil {
			return err
		}

		reportReachability(cmd.Context(), cfg)

		return nil
	},
}

// withElevationHint adds the fix to a validation failure that is about
// privileges, and leaves every other failure alone. A Docker daemon that is
// down is not started by running as Administrator, and a hint that does not
// apply is worse than none.
//
// The layout matches how the agent installer renders a PhaseError hint, so
// both commands read the same way when they fail for the same reason.
func withElevationHint(err error) error {

	var notElevated *system.NotElevatedError

	if errors.As(err, &notElevated) {
		return fmt.Errorf("%w\n  %s", err, system.ElevationHint())
	}

	return err
}

// reportReachability tells the operator where the platform can be reached
// from, and what would stop it being reachable later.
//
// The distinction matters more than it looks. Agents run on other machines
// and dial in; an operator who only ever checks the dashboard from the host
// itself can have a working localhost URL and a platform no agent can reach,
// and nothing in a successful install would have said so.
func reportReachability(ctx context.Context, cfg config.Config) {

	host := system.PrimaryIPv4()

	ui.Success(fmt.Sprintf("DockSight is running on http://%s:%d", host, cfg.Port))

	if host != "localhost" {
		ui.Info(fmt.Sprintf("Local-only access: http://localhost:%d", cfg.Port))
		ui.Info(fmt.Sprintf("Point agents at:   http://%s:%d", host, cfg.Port))
	} else {
		ui.Warning(
			"No network address was found for this host, so only this machine " +
				"can reach the platform. Agents on other machines will not connect.",
		)
	}

	warnAboutDockerDesktop(ctx)
}

// warnAboutDockerDesktop states the one thing this installation cannot do.
//
// Docker Desktop's engine runs inside a VM started by its desktop
// application, in a user session. A Windows host that reboots to a sign-in
// screen therefore has no engine, and no platform, until somebody signs in —
// the containers are all restart: unless-stopped, but there is nothing
// running to restart them. That is the accepted cost of not managing a WSL
// distribution, and an operator has to hear it at install time rather than
// discover it during the outage it causes.
func warnAboutDockerDesktop(ctx context.Context) {

	if runtime.GOOS != "windows" || !system.DockerDesktop(ctx) {
		return
	}

	ui.Warning(
		"This host runs Docker Desktop, whose engine starts with the Docker " +
			"Desktop application in a user session. After a reboot the platform " +
			"stays down until someone signs in.",
	)

	ui.Info(
		"Enable Docker Desktop > Settings > General > \"Start Docker Desktop when you sign in\", " +
			"and sign in after every reboot. For unattended restarts, run the platform on a Linux host.",
	)
}

func init() {
	rootCmd.AddCommand(installCMD)
}
