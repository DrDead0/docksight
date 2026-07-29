package config

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

func Show(cfg Config){
	ui.Info("DockSight configuration:")
	ui.Info("-----------------------")

	ui.Info("Install directory:"+ cfg.InstallationDir)
	ui.Info("Data directory:"+ cfg. 	DataDir)
	ui.Info("Port: " + string(rune(cfg.Port)))
	ui.Info("Version:"+ cfg.Version)
}