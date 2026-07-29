package filesystem

import (
	"os"
)

func CreateDirectories(
	installationDir string,
	dataDir string,
) error {

	directories := []string{
		installationDir,
		dataDir,
	}

	for _, dir := range directories {

		err := os.MkdirAll(
			dir,
			0755,
		)

		if err != nil {
			return err
		}
	}

	return nil
}