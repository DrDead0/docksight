// Package state records what is currently installed, so `docksight update`
// can decide what actually needs updating without reinstalling everything.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State is the on-disk record of an installation.
type State struct {
	// CLIVersion is the version of the binary on PATH.
	CLIVersion string `json:"cli_version"`

	// PlatformVersion is the release tag of the bundle in the install
	// directory. The two are versioned independently on purpose: updating the
	// CLI must not force a platform restart, and vice versa.
	PlatformVersion string `json:"platform_version"`

	InstalledAt time.Time `json:"installed_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Installed reports whether a previous install left a usable record.
func (s State) Installed() bool {
	return s.PlatformVersion != "" || s.CLIVersion != ""
}

// Load reads the state file. A missing file is not an error: it simply means
// nothing is installed yet.
func Load(path string) (State, error) {

	content, err := os.ReadFile(path)

	if err != nil {

		if os.IsNotExist(err) {
			return State{}, nil
		}

		return State{}, err
	}

	var loaded State

	if err := json.Unmarshal(content, &loaded); err != nil {
		return State{}, fmt.Errorf("failed to read install state at %s: %w", path, err)
	}

	return loaded, nil
}

// Save writes the state file atomically, so an interrupted write cannot leave
// an unparseable record that blocks every future update.
func Save(path string, current State) error {

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(current, "", "  ")

	if err != nil {
		return err
	}

	staged, err := os.CreateTemp(filepath.Dir(path), ".state-")

	if err != nil {
		return err
	}

	stagedPath := staged.Name()

	defer os.Remove(stagedPath)

	if _, err := staged.Write(append(encoded, '\n')); err != nil {
		staged.Close()
		return err
	}

	if err := staged.Close(); err != nil {
		return err
	}

	if err := os.Chmod(stagedPath, 0o644); err != nil {
		return err
	}

	return os.Rename(stagedPath, path)
}
