package cmd

import (
	"fmt"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/installer"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/system"

	"github.com/spf13/cobra"
)

var installCMD = &cobra.Command{
	Use:   "install",
	Short: "Install DockSight",
	Long: "Install the DockSight CLI onto PATH, install the platform bundle " +
		"and start the stack.",

	RunE: func(cmd *cobra.Command, args []string) error {

		ui.Banner()

		if err := system.ValidateSystem(); err != nil {
			return err
		}

		cfg := config.Default()

		showConfig(cfg)

		if err := installer.New(cfg, consoleReporter{}).Install(cmd.Context()); err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("DockSight is running on http://localhost:%d", cfg.Port),
		)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCMD)
}
