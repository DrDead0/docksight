package installer

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/config"
)

func CopyDir() error {

	cfg := config.Default()

	config.Show(cfg)

	return filepath.Walk(cfg.InstallationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(cfg.BundleDir, path)
		if err != nil {
			return err
		}

		target := filepath.Join(cfg.InstallationDir, relative)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return os.Chmod(dst, perm)
}