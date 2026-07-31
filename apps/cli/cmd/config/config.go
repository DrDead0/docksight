package config

import "path/filepath"

// Filenames inside the platform bundle and the install directory. Declared
// once so no other package spells them out.
const (
	ComposeFileName    = "dockersight-installation.yml"
	EnvFileName        = ".env"
	EnvExampleFileName = ".env.example"
	StateFileName      = "state.json"

	// BundleRootDir is the top-level directory inside the platform archive.
	BundleRootDir = "bundle"
)

// Config describes where the two deliverables live. The CLI and the platform
// are installed independently, so each has its own destination.
type Config struct {
	// InstallationDir holds the platform: compose file, config, env, state.
	InstallationDir string

	// DataDir holds runtime data that must survive a reinstall.
	DataDir string

	// BinaryPath is where the CLI itself is installed.
	BinaryPath string

	// Port is the host port the stack is published on.
	Port int
}

// ComposePath is the stack definition inside the install directory.
func (c Config) ComposePath() string {
	return filepath.Join(c.InstallationDir, ComposeFileName)
}

// EnvPath is the generated environment file.
func (c Config) EnvPath() string {
	return filepath.Join(c.InstallationDir, EnvFileName)
}

// EnvExamplePath is the template shipped in the bundle.
func (c Config) EnvExamplePath() string {
	return filepath.Join(c.InstallationDir, EnvExampleFileName)
}

// StatePath is the installation record used by `docksight update`.
func (c Config) StatePath() string {
	return filepath.Join(c.InstallationDir, StateFileName)
}
