package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// The remaining agent commands are placeholders. They are declared now so the
// command tree is complete and `docksight agent --help` documents the
// intended surface, but each fails loudly rather than pretending to work.
var agentPlaceholders = []struct {
	use   string
	short string
}{
	{"uninstall", "Remove the DockSight Agent from this host"},
	{"start", "Start the agent service"},
	{"stop", "Stop the agent service"},
	{"restart", "Restart the agent service"},
	{"status", "Show the agent service status"},
	{"logs", "Show the agent service logs"},
	{"config", "Show or edit the agent configuration"},
}

func init() {

	for _, placeholder := range agentPlaceholders {

		command := &cobra.Command{
			Use:   placeholder.use,
			Short: placeholder.short,

			RunE: func(cmd *cobra.Command, args []string) error {

				return fmt.Errorf(
					"docksight agent %s is not implemented yet",
					cmd.Name(),
				)
			},
		}

		agentCMD.AddCommand(command)
	}
}
