package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTempWorkspace(t *testing.T) {

	workspace, err := TempWorkspace()

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(workspace)

	info, err := os.Stat(workspace)

	if err != nil {
		t.Fatal(err)
	}

	if !info.IsDir() {
		t.Fatalf("%s is not a directory", workspace)
	}

	// The staging directory must be writable by this process — the bug this
	// replaced was a fixed /tmp path owned by another user.
	probe := filepath.Join(workspace, "probe")

	if err := os.WriteFile(probe, []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTempWorkspaceIsUniquePerCall(t *testing.T) {

	first, err := TempWorkspace()

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(first)

	second, err := TempWorkspace()

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(second)

	if first == second {
		t.Fatalf("two calls returned the same directory: %s", first)
	}
}
