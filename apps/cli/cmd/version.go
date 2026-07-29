package cmd

import (

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var versionCMD = &cobra.Command{
	Use: "version",
	Short: "Show DockSight version",
	Run: func(cmd *cobra.Command, args []string){
		ui.Info("Docker CLI version:" + version)
	},
}

func init (){
	rootCmd.AddCommand(versionCMD)
}