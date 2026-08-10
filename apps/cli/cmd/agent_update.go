package cmd

import (
	"fmt"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/agent/install"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var (
	agentUpdateVersion string
	agentUpdateURL     string
)

var agentUpdateCMD = &cobra.Command{
	Use:   "update",
	Short: "Update the DockSight Agent on this host",
	Long: "Install a newer agent build over the current one. The platform URL " +
		"and the identity issued to this host are preserved.",

	RunE: func(cmd *cobra.Command, args []string) error {

		layout := install.DefaultLayout()

		// An update targets an existing installation: the platform URL is
		// already on disk, so requiring it again invites a typo that would
		// silently repoint the agent.
		serverURL := agentUpdateURL

		if serverURL == "" {

			existing, err := install.ReadServerURL(layout)

			if err != nil {
				return fmt.Errorf(
					"no agent installation found at %s: run 'docksight agent install --url ...' first (%w)",
					layout.ConfigPath(),
					err,
				)
			}

			serverURL = existing

		} else {

			normalized, err := install.NormalizeServerURL(serverURL)

			if err != nil {
				return fmt.Errorf("invalid platform URL: %w", err)
			}

			serverURL = normalized
		}

		ui.Info("Platform: " + serverURL)

		installer := install.New(layout, serverURL, consoleReporter{})
		installer.Version = agentUpdateVersion

		health, err := installer.Install(cmd.Context())

		if err != nil {
			return err
		}

		if health.Connected {
			ui.Success("DockSight Agent updated and reconnected")
		} else {
			ui.Success("DockSight Agent updated")
		}

		return nil
	},
}

func init() {

	agentUpdateCMD.Flags().StringVar(
		&agentUpdateVersion,
		"version",
		"",
		"update to a specific release tag instead of the latest, e.g. v0.0.1",
	)

	agentUpdateCMD.Flags().StringVar(
		&agentUpdateURL,
		"url",
		"",
		"repoint the agent at a different platform (defaults to the configured one)",
	)

	agentCMD.AddCommand(agentUpdateCMD)
}
