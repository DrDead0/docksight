package cmd

import (
	"github.com/spf13/cobra"
)

// agentCMD groups the commands that manage a DockSight Agent on this host.
var agentCMD = &cobra.Command{
	Use:   "agent",
	Short: "Manage the DockSight Agent on this host",
	Long: "The DockSight Agent runs beside a Docker Engine on a remote host " +
		"and connects outbound to a DockSight platform. It exposes no inbound ports.",
}

func init() {
	rootCmd.AddCommand(agentCMD)
}
