package filesystem

import (
	"io"
	"os"
	"path/filepath"
)

// CopyDir recursively copies the contents of source into destination,
// preserving file modes.
func CopyDir(source string, destination string) error {

	info, err := os.Stat(source)

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return CopyFile(source, destination, info.Mode())
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

		return CopyFile(path, target, info.Mode())
	})
}

// CopyFile copies a single file, replacing destination if it exists.
func CopyFile(source string, destination string, perm os.FileMode) error {

	in, err := os.Open(source)

	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(destination)

	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(destination)
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Chmod(destination, perm)
}
