package release

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTarGz unpacks a gzipped tarball into destination.
//
// Entries that would escape destination are rejected: install runs as root,
// so an archive entry named "../../etc/cron.d/x" would otherwise write
// anywhere on the host.
func ExtractTarGz(source string, destination string) error {

	file, err := os.Open(source)

	if err != nil {
		return err
	}

	defer file.Close()

	gzipReader, err := gzip.NewReader(file)

	if err != nil {
		return fmt.Errorf("%s is not a gzip archive: %w", filepath.Base(source), err)
	}

	defer gzipReader.Close()

	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}

	root, err := filepath.Abs(destination)

	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)

	for {

		header, err := tarReader.Next()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		target, err := safeJoin(root, header.Name)

		if err != nil {
			return err
		}

		switch header.Typeflag {

		case tar.TypeDir:

			if err := os.MkdirAll(target, os.FileMode(header.Mode)|0o700); err != nil {
				return err
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}

			if err := writeFile(target, tarReader, os.FileMode(header.Mode)); err != nil {
				return err
			}

		default:
			// Symlinks, devices and hard links have no place in a bundle and
			// are the usual vehicle for escaping the destination directory.
			continue
		}
	}
}

// safeJoin resolves name inside root, rejecting anything that climbs out.
func safeJoin(root string, name string) (string, error) {

	if isAbsoluteEntry(name) {
		return "", fmt.Errorf("archive entry escapes the destination: %s", name)
	}

	cleaned := filepath.Clean(filepath.FromSlash(name))

	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry escapes the destination: %s", name)
	}

	target := filepath.Join(root, cleaned)

	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry escapes the destination: %s", name)
	}

	return target, nil
}

// isAbsoluteEntry reports whether an archive entry names an absolute path.
// Tar entries always use forward slashes, so this cannot rely on
// filepath.IsAbs: on Windows "/etc/passwd" is not an absolute path, and the
// entry would be silently accepted on one platform and rejected on another.
func isAbsoluteEntry(name string) bool {

	if name == "" {
		return false
	}

	if name[0] == '/' || name[0] == '\\' {
		return true
	}

	// "C:\..." — only meaningful on Windows, where VolumeName parses it.
	return filepath.VolumeName(name) != ""
}

func writeFile(target string, source io.Reader, mode os.FileMode) error {

	if mode == 0 {
		mode = 0o644
	}

	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())

	if err != nil {
		return err
	}

	if _, err := io.Copy(file, source); err != nil {
		file.Close()
		return err
	}

	return file.Close()
}
