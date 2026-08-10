// Package selfinstall places a CLI executable at its final location.
//
// It exists because the CLI is downloaded and executed from an arbitrary path
// ("./docksight", "~/Downloads/docksight-cli-v0.0.1-linux-amd64") and must end
// up at a stable one, without the running process being disturbed.
package selfinstall

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Running returns the path of the executable for this process, with symlinks
// resolved. The filename is never assumed: the user may have renamed the
// download to anything.
func Running() (string, error) {

	executable, err := os.Executable()

	if err != nil {
		return "", fmt.Errorf("failed to locate the running executable: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(executable)

	if err != nil {
		// A missing symlink target is unusual but not fatal; the raw path is
		// still the best answer available.
		return executable, nil
	}

	return resolved, nil
}

// InstallRunning copies the executable of this process to destination.
//
// It reports false when the process is already running from destination —
// copying a file onto itself would truncate it and destroy the installed CLI.
func InstallRunning(destination string) (bool, error) {

	source, err := Running()

	if err != nil {
		return false, err
	}

	same, err := sameFile(source, destination)

	if err != nil {
		return false, err
	}

	if same {
		return false, nil
	}

	return true, Install(source, destination)
}

// Install places the executable at source into destination.
//
// The write is staged in a temporary file next to destination and moved into
// place with a rename. That matters for two reasons: the swap is atomic, so a
// crash mid-copy can never leave a half-written binary on PATH; and on Unix a
// running executable cannot be written to (ETXTBSY) but its directory entry
// can be replaced, which is exactly what lets `docksight install` overwrite
// the very binary it is executing and keep running afterwards. The running
// process holds the old inode open until it exits.
func Install(source string, destination string) error {

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	staged, err := os.CreateTemp(filepath.Dir(destination), ".docksight-")

	if err != nil {
		return fmt.Errorf("failed to stage the binary next to %s: %w", destination, err)
	}

	stagedPath := staged.Name()

	defer os.Remove(stagedPath)

	binary, err := os.Open(source)

	if err != nil {
		staged.Close()
		return err
	}

	defer binary.Close()

	if _, err := io.Copy(staged, binary); err != nil {
		staged.Close()
		return err
	}

	if err := staged.Close(); err != nil {
		return err
	}

	if err := os.Chmod(stagedPath, 0o755); err != nil {
		return err
	}

	return replace(stagedPath, destination)
}

// replace moves staged over destination, working around platforms that refuse
// to rename onto an open file.
func replace(staged string, destination string) error {

	if err := os.Rename(staged, destination); err == nil {
		return nil
	}

	// Windows (and some network filesystems) reject a rename onto a file that
	// is currently open. Move the old binary aside and retry.
	previous := destination + ".old"

	os.Remove(previous)

	if err := os.Rename(destination, previous); err != nil {
		return fmt.Errorf("failed to replace %s: %w", destination, err)
	}

	if err := os.Rename(staged, destination); err != nil {
		// Put the original back rather than leaving nothing on PATH.
		os.Rename(previous, destination)
		return fmt.Errorf("failed to replace %s: %w", destination, err)
	}

	os.Remove(previous)

	return nil
}

func sameFile(source string, destination string) (bool, error) {

	sourceInfo, err := os.Stat(source)

	if err != nil {
		return false, err
	}

	destinationInfo, err := os.Stat(destination)

	if err != nil {

		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return os.SameFile(sourceInfo, destinationInfo), nil
}
