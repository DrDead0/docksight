package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/installer"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/system"

	"github.com/spf13/cobra"
)

// bundleRootDir is the top-level directory inside the installer archive.
const bundleRootDir = "bundle"

var installCMD = &cobra.Command{
	Use:   "install",
	Short: "Install DockSight",

	RunE: func(cmd *cobra.Command, args []string) error {

		ui.Banner()

		// 1. Validate system
		if err := system.ValidateSystem(); err != nil {
			return err
		}

		// 2. Resolve configuration
		cfg := config.Default()

		showConfig(cfg)

		// 3. Prepare installation directories
		ui.Info("Preparing installation directories")

		if err := filesystem.CreateDirectories(
			cfg.InstallationDir,
			cfg.DataDir,
		); err != nil {
			return err
		}

		ui.Success("Directories created")

		// 4. Get latest release
		ui.Info("Checking latest DockSight release")

		githubRelease, err := release.LatestGithubRelease()

		if err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Latest version %s found", githubRelease.TagName),
		)

		// 5. Find installer asset
		asset, err := release.FindInstallerAsset(githubRelease)

		if err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("Installer package %s found", asset.Name),
		)

		// 6. Download installer package into a private staging directory
		workspace, err := filesystem.TempWorkspace()

		if err != nil {
			return err
		}

		// The workspace only stages the bundle on its way to the install
		// directory, so it is removed once this command returns.
		defer os.RemoveAll(workspace)

		archive := filepath.Join(workspace, asset.Name)

		ui.Info("Downloading installer package")

		if err := release.Download(asset.URL, archive); err != nil {
			return err
		}

		ui.Success("Download completed")

		// 7. Extract installer package
		ui.Info("Extracting installer package")

		extractPath := filepath.Join(workspace, "extracted")

		if err := release.ExtractTarGz(archive, extractPath); err != nil {
			return err
		}

		ui.Success("Installer extracted")

		// 8. Copy the bundle into the installation directory
		ui.Info("Installing bundle into " + cfg.InstallationDir)

		if err := installer.CopyDir(
			filepath.Join(extractPath, bundleRootDir),
			cfg.InstallationDir,
		); err != nil {
			return err
		}

		ui.Success(
			fmt.Sprintf("DockSight %s installed in %s", githubRelease.TagName, cfg.InstallationDir),
		)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCMD)
}
