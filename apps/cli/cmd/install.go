package cmd

import (

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/installer"
	"github.com/spf13/cobra"

)

var installCMD = &cobra.Command{
	Use: "install",
	Short: "Install DockSight",
	RunE: func (cmd *cobra.Command, args []string) error  {
      if err :=installer.Install(); err!=nil{
		ui.Error(err.Error())
		return err
	}
	 return nil
	},
}

func init (){
	rootCmd.AddCommand(installCMD)
}


