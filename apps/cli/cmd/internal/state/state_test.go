package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileIsNotAnError(t *testing.T) {

	loaded, err := Load(filepath.Join(t.TempDir(), "state.json"))

	if err != nil {
		t.Fatal(err)
	}

	if loaded.Installed() {
		t.Fatal("a missing state file must read as not installed")
	}
}

func TestSaveAndLoad(t *testing.T) {

	path := filepath.Join(t.TempDir(), "nested", "state.json")

	saved := State{
		CLIVersion:      "v0.0.1",
		PlatformVersion: "v0.0.1",
		InstalledAt:     time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC),
	}

	if err := Save(path, saved); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)

	if err != nil {
		t.Fatal(err)
	}

	if loaded.CLIVersion != "v0.0.1" || loaded.PlatformVersion != "v0.0.1" {
		t.Fatalf("versions not round-tripped: %+v", loaded)
	}

	if !loaded.InstalledAt.Equal(saved.InstalledAt) {
		t.Fatalf("InstalledAt not round-tripped: %v", loaded.InstalledAt)
	}

	if !loaded.Installed() {
		t.Fatal("expected the state to report an installation")
	}
}

// The CLI and the platform move independently.
func TestVersionsAreIndependent(t *testing.T) {

	path := filepath.Join(t.TempDir(), "state.json")

	if err := Save(path, State{CLIVersion: "v0.0.1", PlatformVersion: "v0.0.1"}); err != nil {
		t.Fatal(err)
	}

	current, err := Load(path)

	if err != nil {
		t.Fatal(err)
	}

	// Moving one field to a different version is what makes the other field
	// staying put observable.
	current.CLIVersion = "v0.0.2"

	if err := Save(path, current); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(path)

	if err != nil {
		t.Fatal(err)
	}

	if reloaded.CLIVersion != "v0.0.2" || reloaded.PlatformVersion != "v0.0.1" {
		t.Fatalf("independent update failed: %+v", reloaded)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {

	path := filepath.Join(t.TempDir(), "state.json")

	if err := Save(path, State{PlatformVersion: "v0.0.1"}); err != nil {
		t.Fatal(err)
	}

	// The second value has to differ, or an overwrite that silently did
	// nothing would pass.
	if err := Save(path, State{PlatformVersion: "v0.0.2"}); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)

	if err != nil {
		t.Fatal(err)
	}

	if loaded.PlatformVersion != "v0.0.2" {
		t.Fatalf("got %q", loaded.PlatformVersion)
	}
}

func TestSaveLeavesNoStagingFile(t *testing.T) {

	directory := t.TempDir()

	if err := Save(filepath.Join(directory, "state.json"), State{CLIVersion: "v1"}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(directory)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("unexpected leftovers: %+v", entries)
	}
}

func TestLoadCorruptFile(t *testing.T) {

	path := filepath.Join(t.TempDir(), "state.json")

	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected an error for a corrupt state file")
	}
}
