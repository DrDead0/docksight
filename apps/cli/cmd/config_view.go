package cmd

import (
	"fmt"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

// showConfig prints the resolved configuration. Rendering lives in the cmd
// layer so the config package stays free of UI concerns.
func showConfig(cfg config.Config) {

	ui.Info("DockSight configuration:")
	ui.Info("-----------------------")

	ui.Info("Install directory: " + cfg.InstallationDir)
	ui.Info("Data directory:    " + cfg.DataDir)
	ui.Info(fmt.Sprintf("Port:              %d", cfg.Port))
	ui.Info("Version:           " + cfg.Version)
}
