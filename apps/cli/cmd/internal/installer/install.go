package installer

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/system"
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

	ui.Step(4, 6, "Downloading DockSight release")

	rel, err := release.Latest()

	if err != nil {
		return err
	}

	ui.Info(
		"Downloading version " + rel.Version,
	)

	ui.Success("DockSight installation preparation complete")

	return nil
}
