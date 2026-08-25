package cmd

import (
	"fmt"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/system"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var doctorAgent bool

var doctorCMD = &cobra.Command{
	Use:   "doctor",
	Short: "Check whether this host can run DockSight",
	Long: "Run the same preflight checks as install, without changing anything " +
		"on the host. Exits non-zero if any check fails, so it is safe to use " +
		"in scripts (`docksight doctor && ...`).\n\n" +
		"By default this validates the platform install requirements. Pass " +
		"--agent to validate agent install requirements instead (service manager " +
		"and outbound connectivity in addition to Docker).\n\n" +
		"Unlike install, doctor reports every failing check rather than " +
		"stopping at the first one.",

	RunE: func(cmd *cobra.Command, args []string) error {

		ui.Banner()

		requirements := system.PlatformRequirements()
		label := "platform"

		if doctorAgent {
			requirements = system.AgentRequirements()
			label = "agent"
		}

		ui.Info(fmt.Sprintf("Running %s preflight checks", label))

		// ValidateAll (not Validate): a diagnostic should surface every
		// problem at once. Install still uses Validate so it fails closed
		// before writing anything to the host.
		if err := system.ValidateAll(cmd.Context(), consoleReporter{}, requirements); err != nil {
			return err
		}

		ui.Success("All checks passed")
		return nil
	},
}

func init() {
	doctorCMD.Flags().BoolVar(
		&doctorAgent,
		"agent",
		false,
		"run agent install checks (service manager + connectivity) instead of platform checks",
	)
	rootCmd.AddCommand(doctorCMD)
}
