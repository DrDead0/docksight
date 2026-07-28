package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const status = "all active"

var statusCMD = &cobra.Command{
	Use:  "status",
	Short:"check tha  docker sight web, server, and db status",
    Run: func (cmd *cobra.Command, args []string){
		fmt.Print("Docksight status: ", status)
	},
}

func init (){
	rootCmd.AddCommand(statusCMD)
}