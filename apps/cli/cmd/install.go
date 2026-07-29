package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/installer"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

// downloadDir is where the installer bundle is staged before extraction.
const downloadDir = "/tmp"


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

		// 4. Download installer package
		destination := filepath.Join(downloadDir, asset.Name)

		ui.Info("Downloading installer package")

		if err := release.Download(asset.URL, destination); err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Download completed: %s", destination),
		)

    ui.Info(
	"Extracting installer package",
)
extractPath := "/tmp/docksight"


err = release.ExtractTarGz(
	destination,
	extractPath,
)


if err != nil {
	return err
}


ui.Success(
	"Installer extracted",
)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCMD)
}
