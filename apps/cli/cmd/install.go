package cmd

import (
	"fmt"

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
			return err
		}

		cfg := config.Default()

		showConfig(cfg)

		if err := installer.New(cfg, consoleReporter{}).Install(cmd.Context()); err != nil {
			return err
		}

		host := system.PrimaryIPv4()
		ui.Success(
			fmt.Sprintf("DockSight is running on http://%s:%d", host, cfg.Port),
		)
		if host != "localhost" {
			ui.Info(
				fmt.Sprintf("Local-only access: http://localhost:%d", cfg.Port),
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCMD)
}
