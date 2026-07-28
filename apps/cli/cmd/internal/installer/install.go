package installer

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/system"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

func Install() error {

	ui.Banner()

	ui.Info("Starting DockSight installation")

	if err := system.ValidateSystem(); err != nil {
		return err
	}

	ui.Success("System validation completed")


	// Future steps will come here:

	// Step 2:
	// config.Load()

	// Step 3:
	// release.Download()

	// Step 4:
	// compose.Generate()

	// Step 5:
	// compose.Up()


	ui.Success("DockSight installation preparation complete")

	return nil
}