package cmd

import (
	"fmt"
	"os"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/agent/install"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/system"

	"github.com/spf13/cobra"
)

var agentLogsLines int

var agentLogsCMD = &cobra.Command{
	Use:   "logs",
	Short: "Show the agent service logs",
	Long: "Print recent agent service output. On Linux this reads the journal; " +
		"on Windows it reads the rotating log under ProgramData. Output is passed " +
		"through unmodified. Use the platform follow command from `agent status` " +
		"for streaming (`--follow` is out of scope for this command).",

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

		output, err := manager.Logs(cmd.Context(), unit, agentLogsLines)
		if err != nil {
			if isPermissionError(err) {
				return fmt.Errorf(
					"could not read logs for %s: %w\n  %s",
					unit,
					err,
					system.ElevationHint(),
				)
			}
			return fmt.Errorf("could not read logs for %s: %w", unit, err)
		}

		// Pass through as-is; do not reformat or colourise.
		fmt.Fprint(cmd.OutOrStdout(), output)
		if len(output) > 0 && output[len(output)-1] != '\n' {
			fmt.Fprintln(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	agentLogsCMD.Flags().IntVarP(
		&agentLogsLines,
		"lines",
		"n",
		50,
		"number of log lines to show",
	)
	agentCMD.AddCommand(agentLogsCMD)
}
