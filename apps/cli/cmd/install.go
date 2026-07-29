package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/installer"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

var installCMD = &cobra.Command{
	Use:   "install",
	Short: "Install DockSight",

	RunE: func(cmd *cobra.Command, args []string) error {

		ui.Banner()

		// 1. Validate system
		if err := installer.Install(); err != nil {
			return err
		}

		ui.Info("Checking latest DockSight release")

		// 2. Get latest release
		githubRelease, err := release.LatestGithubRelease()

		if err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Latest version %s found", githubRelease.TagName),
		)

		// 3. Find installer asset
		asset, err := release.FindInstallerAsset(githubRelease)

		if err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Installer package %s found", asset.Name),
		)

		// 4. Download installer package into a private staging directory
		workspace, err := filesystem.TempWorkspace()

		if err != nil {
			return err
		}

		destination := filepath.Join(workspace, asset.Name)

		ui.Info("Downloading installer package")

		if err := release.Download(asset.URL, destination); err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Download completed: %s", destination),
		)

		// 5. Extract installer package
		ui.Info("Extracting installer package")

		extractPath := filepath.Join(workspace, "extracted")

		if err := release.ExtractTarGz(destination, extractPath); err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Installer extracted to %s", extractPath),
		)
    ui.Info("Copy exrtacted ")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCMD)
}
