package selfinstall

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInstallCopiesBinary(t *testing.T) {

	source := filepath.Join(t.TempDir(), "docksight-cli-v0.0.1-linux-amd64")

	if err := os.WriteFile(source, []byte("#!/bin/sh\necho docksight\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(t.TempDir(), "bin", "docksight")

	if err := Install(source, destination); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(destination)

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "#!/bin/sh\necho docksight\n" {
		t.Fatalf("content mismatch: %q", string(content))
	}
}

// The downloaded file may not be executable; what lands on PATH must be.
func TestInstallSetsExecutableBit(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits are not meaningful on windows")
	}

	source := filepath.Join(t.TempDir(), "downloaded")

	if err := os.WriteFile(source, []byte("binary"), 0o600); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(t.TempDir(), "docksight")

	if err := Install(source, destination); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(destination)

	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("installed binary is not executable: %04o", info.Mode().Perm())
	}
}

func TestInstallReplacesExistingBinary(t *testing.T) {

	destination := filepath.Join(t.TempDir(), "docksight")

	if err := os.WriteFile(destination, []byte("old version"), 0o755); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(t.TempDir(), "new")

	if err := os.WriteFile(source, []byte("new version"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Install(source, destination); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(destination)

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "new version" {
		t.Fatalf("binary was not replaced: %q", string(content))
	}
}

// No temporary staging file may survive a successful install.
func TestInstallLeavesNoStagingFile(t *testing.T) {

	directory := t.TempDir()
	destination := filepath.Join(directory, "docksight")

	source := filepath.Join(t.TempDir(), "new")

	if err := os.WriteFile(source, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Install(source, destination); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(directory)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 || entries[0].Name() != "docksight" {
		t.Fatalf("unexpected leftovers: %+v", entries)
	}
}

func TestInstallMissingSource(t *testing.T) {

	err := Install(
		filepath.Join(t.TempDir(), "absent"),
		filepath.Join(t.TempDir(), "docksight"),
	)

	if err == nil {
		t.Fatal("expected an error for a missing source")
	}
}

// Copying a file onto itself truncates it. Running `sudo docksight install`
// from the installed path must not destroy the CLI.
func TestInstallRunningSkipsSelf(t *testing.T) {

	running, err := Running()

	if err != nil {
		t.Fatal(err)
	}

	before, err := os.Stat(running)

	if err != nil {
		t.Fatal(err)
	}

	installed, err := InstallRunning(running)

	if err != nil {
		t.Fatal(err)
	}

	if installed {
		t.Fatal("the running binary was copied over itself")
	}

	after, err := os.Stat(running)

	if err != nil {
		t.Fatal(err)
	}

	if after.Size() != before.Size() {
		t.Fatalf("the running binary was modified: %d -> %d", before.Size(), after.Size())
	}
}

func TestInstallRunningCopiesToNewPath(t *testing.T) {

	destination := filepath.Join(t.TempDir(), "docksight")

	installed, err := InstallRunning(destination)

	if err != nil {
		t.Fatal(err)
	}

	if !installed {
		t.Fatal("expected the binary to be installed")
	}

	source, err := Running()

	if err != nil {
		t.Fatal(err)
	}

	original, err := os.Stat(source)

	if err != nil {
		t.Fatal(err)
	}

	copied, err := os.Stat(destination)

	if err != nil {
		t.Fatal(err)
	}

	if copied.Size() != original.Size() {
		t.Fatalf("size mismatch: %d != %d", copied.Size(), original.Size())
	}
}

func TestRunningResolvesAPath(t *testing.T) {

	path, err := Running()

	if err != nil {
		t.Fatal(err)
	}

	if !filepath.IsAbs(path) {
		t.Fatalf("expected an absolute path, got %q", path)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
