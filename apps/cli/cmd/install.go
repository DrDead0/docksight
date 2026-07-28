package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

)

var installCMD = &cobra.Command{
	Use: "install",
	Short: "Install DockSight",
	Run: func (cmd *cobra.Command, args []string)  {
		fmt.Println("Installing dockSight")
		fmt.Println("Checking Docker...")
		fmt.Println("Creating /opt/docksight...")
		fmt.Println("Installation preparation complete")
	},
}

func init (){
	rootCmd.AddCommand(installCMD)
}


