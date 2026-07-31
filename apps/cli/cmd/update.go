package cmd

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/installer"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var (
	updateCLIOnly      bool
	updatePlatformOnly bool
	updateForce        bool
	updateTag          string
)

var updateCMD = &cobra.Command{
	Use:   "update",
	Short: "Update DockSight",
	Long: "Update the CLI binary and the platform bundle. The two are " +
		"versioned independently and can be updated separately.",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg := config.Default()

		install := installer.New(cfg, consoleReporter{})

		ctx := cmd.Context()

		var (
			target *release.Release
			err    error
		)

		if updateTag != "" {
			target, err = install.Releases.ByTag(ctx, updateTag)
		} else {
			target, err = install.LatestRelease(ctx)
		}

		if err != nil {
			return err
		}

		current, err := install.State()

		if err != nil {
			return err
		}

		if !updatePlatformOnly {

			if current.CLIVersion == target.Version() && !updateForce {
				ui.Info("CLI already at " + target.Version())
			} else if err := install.UpdateCLI(ctx, target); err != nil {
				return err
			}
		}

		if !updateCLIOnly {

			if current.PlatformVersion == target.Version() && !updateForce {
				ui.Info("Platform already at " + target.Version())
			} else if err := install.UpdatePlatform(ctx, target); err != nil {
				return err
			}
		}

		ui.Success("DockSight is up to date")

		return nil
	},
}

func init() {

	updateCMD.Flags().BoolVar(&updateCLIOnly, "cli", false, "update only the CLI binary")
	updateCMD.Flags().BoolVar(&updatePlatformOnly, "platform", false, "update only the platform bundle")
	updateCMD.Flags().BoolVar(&updateForce, "force", false, "reinstall even if already at the target version")
	updateCMD.Flags().StringVar(&updateTag, "version", "", "update to a specific release tag instead of the latest")

	updateCMD.MarkFlagsMutuallyExclusive("cli", "platform")

	rootCmd.AddCommand(updateCMD)
}
