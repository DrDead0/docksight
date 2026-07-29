package system

import (
	"os/exec"
	"os"
	"path/filepath"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

func CheckDockerPermissions() error {

	cmd := exec.Command(
		"docker",
		"info",
	)

	if err := cmd.Run(); err != nil {
		 ui.Error(
			"cannot access Docker daemon: permission denied",
		)
		return  nil
	}

	ui.Info("Docker permission OK")

	return nil
}

func CheckInstallDirectoryPermission() error {

	testFile := filepath.Join(
		InstallationDir,
		".permission-test",
	)


	file, err := os.Create(testFile)

	if err != nil {
		return err
	}


	file.Close()

	os.Remove(testFile)


	ui.Info("✓ Installation directory writable")

	return nil
}

func IsRoot() bool {

	return os.Geteuid() == 0
}

 