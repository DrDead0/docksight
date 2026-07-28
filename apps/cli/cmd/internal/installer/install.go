package installer

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/system"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

func Install() error {

	ui.Banner()

	ui.Info("Starting DockSight installation")


	// 1. Validate system

	if err := system.ValidateSystem(); err != nil {
		return err
	}


	// 2. Load default configuration

	cfg := config.Default()

	config.Show(cfg)


	// 3. Prepare directories

	ui.Step(6, 6, "Preparing installation directories")

	if err := filesystem.CreateDirectories(
		cfg.InstallationDir,
		cfg.DataDir,
	); err != nil {
		return err
	}

	ui.Success("Directories created")


	ui.Success("DockSight installation preparation complete")

	return nil
}