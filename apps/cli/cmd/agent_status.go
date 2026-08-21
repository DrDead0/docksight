package cmd

import (
	"fmt"
	"os"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/agent/install"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/system"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var agentStatusCMD = &cobra.Command{
	Use:   "status",
	Short: "Show the agent service status",
	Long: "Report whether the agent is installed and running, which platform " +
		"URL it targets, and whether an identity has been issued. Exits non-zero " +
		"when the agent is not installed or not running.",

	RunE: func(cmd *cobra.Command, args []string) error {
		layout := agentLayout()

		if _, err := os.Stat(layout.ConfigPath()); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(
					"agent is not installed (no config at %s)",
					layout.ConfigPath(),
				)
			}
			return err
		}

		manager := install.NewController(layout)
		unit := layout.ServiceName

		active, err := manager.IsActive(cmd.Context(), unit)
		if err != nil {
			if isPermissionError(err) {
				return fmt.Errorf(
					"could not query %s status: %w\n  %s",
					unit,
					err,
					system.ElevationHint(),
				)
			}
			return fmt.Errorf("could not query %s status: %w", unit, err)
		}

		serverURL, err := install.ReadServerURL(layout)
		if err != nil {
			serverURL = "(unavailable: " + err.Error() + ")"
		}

		identity := "not issued"
		if install.IdentityExists(layout) {
			identity = "issued"
		}

		if active {
			ui.Success(unit + ": running")
		} else {
			ui.Warning(unit + ": not running")
		}
		ui.Info("Platform URL: " + serverURL)
		ui.Info("Identity: " + identity)
		ui.Info("Follow logs: " + layout.LogsCommand())

		if !active {
			return fmt.Errorf("%s is not running", unit)
		}
		return nil
	},
}

func init() {
	agentCMD.AddCommand(agentStatusCMD)
}
