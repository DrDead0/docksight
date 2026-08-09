package cmd

import (
	"context"
	"fmt"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/agent/install"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/systemd"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var agentConfigSetURL string

type agentServiceRestarter interface {
	Restart(ctx context.Context, unit string) error
}

type agentConfigDetails struct {
	platformURL  string
	binaryPath   string
	configPath   string
	unitName     string
	dockerSocket string
	registered   bool
}

var agentConfigCMD = &cobra.Command{
	Use:   "config",
	Short: "Show or update the agent configuration",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {

		layout := install.DefaultLayout()

		if cmd.Flags().Changed("set-url") {

			serverURL, err := setAgentServerURL(
				cmd.Context(),
				layout,
				systemd.NewManager(),
				agentConfigSetURL,
			)

			if err != nil {
				return err
			}

			ui.Success("Platform URL updated to " + serverURL)
			ui.Success(layout.ServiceName + " restarted")

			return nil
		}

		details, err := readAgentConfig(layout)

		if err != nil {
			return err
		}

		showAgentConfig(details)

		return nil
	},
}

func readAgentConfig(layout install.Layout) (agentConfigDetails, error) {

	serverURL, err := install.ReadServerURL(layout)

	if err != nil {
		return agentConfigDetails{}, fmt.Errorf(
			"read agent configuration at %s: %w",
			layout.ConfigPath(),
			err,
		)
	}

	return agentConfigDetails{
		platformURL:  serverURL,
		binaryPath:   layout.BinaryPath,
		configPath:   layout.ConfigPath(),
		unitName:     layout.ServiceName,
		dockerSocket: layout.DockerSocket,
		registered:   install.IdentityExists(layout),
	}, nil
}

func showAgentConfig(details agentConfigDetails) {

	registered := "no"

	if details.registered {
		registered = "yes"
	}

	ui.Info("Agent configuration:")
	ui.Info("Platform URL:  " + details.platformURL)
	ui.Info("Binary path:   " + details.binaryPath)
	ui.Info("Config path:   " + details.configPath)
	ui.Info("Unit name:     " + details.unitName)
	ui.Info("Docker socket: " + details.dockerSocket)
	ui.Info("Registered:    " + registered)
}

func setAgentServerURL(
	ctx context.Context,
	layout install.Layout,
	restarter agentServiceRestarter,
	rawURL string,
) (string, error) {

	serverURL, err := install.NormalizeServerURL(rawURL)

	if err != nil {
		return "", fmt.Errorf("invalid platform URL: %w", err)
	}

	if err := install.UpdateServerURL(layout, serverURL); err != nil {
		return "", fmt.Errorf("update %s: %w", layout.ConfigPath(), err)
	}

	if err := restarter.Restart(ctx, layout.ServiceName); err != nil {
		return "", fmt.Errorf(
			"configuration updated, but failed to restart %s: %w",
			layout.ServiceName,
			err,
		)
	}

	return serverURL, nil
}

func init() {

	agentConfigCMD.Flags().StringVar(
		&agentConfigSetURL,
		"set-url",
		"",
		"change the DockSight platform URL and restart the agent",
	)

	agentCMD.AddCommand(agentConfigCMD)
}
