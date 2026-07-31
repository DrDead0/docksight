package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/release"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/selfinstall"
)

// downloadBinary fetches the agent build for target from the given release
// and installs it at the configured path.
//
// It reuses the platform installer's building blocks: release.Client for
// discovery and transfer, and selfinstall for the atomic replace that lets a
// running agent be upgraded in place.
func (i *Installer) downloadBinary(ctx context.Context, rel *release.Release) error {

	asset, err := rel.AgentBinary(i.target())

	if err != nil {
		return err
	}

	i.Report.Step("Downloading " + asset.Name)

	workspace, err := filesystem.TempWorkspace()

	if err != nil {
		return err
	}

	defer os.RemoveAll(workspace)

	downloaded := filepath.Join(workspace, asset.Name)

	if err := i.Releases.Download(ctx, *asset, downloaded); err != nil {
		return err
	}

	binary := downloaded

	// The pipeline may publish the agent raw or archived; both are usable.
	if asset.IsArchive() {

		extracted := filepath.Join(workspace, "agent")

		if err := release.ExtractTarGz(downloaded, extracted); err != nil {
			return err
		}

		binary, err = findExecutable(extracted)

		if err != nil {
			return err
		}
	}

	if err := selfinstall.Install(binary, i.Layout.BinaryPath); err != nil {
		return fmt.Errorf("failed to install the agent binary: %w", err)
	}

	i.Report.Success("Agent binary installed at " + i.Layout.BinaryPath)

	return nil
}

// findExecutable locates the single binary inside an extracted archive.
func findExecutable(root string) (string, error) {

	entries, err := os.ReadDir(root)

	if err != nil {
		return "", err
	}

	for _, entry := range entries {

		if !entry.IsDir() {
			return filepath.Join(root, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no executable found in %s", root)
}
