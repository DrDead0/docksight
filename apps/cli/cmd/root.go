package cmd


import (
"fmt"
"os"

"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
 Use: "dockersight",
 Short: "DockSight CLI - Container monitoring platform installer",
}

func Execute(){
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}