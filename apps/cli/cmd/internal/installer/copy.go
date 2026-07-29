package installer

import (
	"io"
	"os"
	"path/filepath"
)

// CopyDir recursively copies the contents of source into destination,
// preserving file modes. Both paths are supplied by the caller so this stays
// independent of any particular configuration.
func CopyDir(source string, destination string) error {

	info, err := os.Stat(source)

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return copyFile(source, destination, info.Mode())
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)

		if err != nil {
			return err
		}

		target := filepath.Join(destination, relative)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		if !info.Mode().IsRegular() {
			// Symlinks and devices are not part of an installation bundle.
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src string, dst string, perm os.FileMode) error {

	in, err := os.Open(src)

	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)

	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Chmod(dst, perm)
}
