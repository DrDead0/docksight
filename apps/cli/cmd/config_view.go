package cmd

import (
	"fmt"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

// showConfig prints the resolved installation layout. Rendering lives in the
// cmd layer so the config package stays free of UI concerns.
func showConfig(cfg config.Config) {

	ui.Info("DockSight configuration:")
	ui.Info("-----------------------")

	ui.Info("CLI binary:        " + cfg.BinaryPath)
	ui.Info("Install directory: " + cfg.InstallationDir)
	ui.Info("Data directory:    " + cfg.DataDir)
	ui.Info(fmt.Sprintf("Port:              %d", cfg.Port))
}
