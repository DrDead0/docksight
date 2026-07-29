package cmd

import (

	"github.com/spf13/cobra"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

const status = "all active"

var statusCMD = &cobra.Command{
	Use:  "status",
	Short:"check tha  docker sight web, server, and db status",
    Run: func (cmd *cobra.Command, args []string){
		ui.Info("Docksight status: "+ status)
	},
}

func init (){
	rootCmd.AddCommand(statusCMD)
}