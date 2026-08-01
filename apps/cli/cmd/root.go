package cmd

import (
	"os"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "docksight",
	Short: "DockSight CLI - Container monitoring platform installer",

	// A failing install is a runtime error, not a usage mistake: printing the
	// flag reference after it buries the actual cause. Errors are reported
	// once, by Execute.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {

	// A mistyped flag is the one case where the usage block does help.
	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		command.Println(command.UsageString())
		return err
	})

	if err := rootCmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}
