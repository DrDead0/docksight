package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Default returns the standard installation layout for this host.
func Default() Config {

	if runtime.GOOS == "windows" {
		return WindowsDefault()
	}

	return LinuxDefault()
}

// LinuxDefault is the standard installation on a Linux host.
func LinuxDefault() Config {

	return Config{
		InstallationDir: "/opt/docksight",
		DataDir:         "/var/lib/docksight",
		BinaryPath:      "/usr/local/bin/docksight",
		Port:            2002,
	}
}

// WindowsDefault is the standard installation on a Windows host.
//
// The split follows the platform's own convention. Program Files holds the
// executable, which only an installer may write. ProgramData holds the
// installation — the compose file, the generated .env and the state record —
// because that is machine-wide state that must survive any particular user
// account being deleted.
//
// The stack's own data is not here. Postgres and Redis write to named Docker
// volumes managed by the Engine, so DataDir holds only what the CLI itself
// puts there.
//
// The elements are joined with an explicit separator rather than with
// filepath.Join. filepath.Join would give the same answer on the only host
// this is reached from — unlike the agent's Layout, nothing here is written
// into a file another machine reads, so there is no correctness problem to
// avoid. What it would cost is the ability to assert any of this from a
// Linux test runner, and these paths are exactly the kind of thing that is
// worth pinning in CI.
func WindowsDefault() Config {

	programData := environmentOr("ProgramData", `C:\ProgramData`)

	return Config{
		InstallationDir: windowsJoin(programData, "DockSight", "platform"),
		DataDir:         windowsJoin(programData, "DockSight", "data"),
		BinaryPath: windowsJoin(
			environmentOr("ProgramFiles", `C:\Program Files`),
			"DockSight",
			"docksight.exe",
		),
		Port: 2002,
	}
}

func windowsJoin(elements ...string) string {
	return strings.Join(elements, `\`)
}

// InstallDirectory is the directory the CLI binary is installed into. It is
// what has to be on PATH for `docksight` to resolve in a new shell.
func (c Config) InstallDirectory() string {
	return filepath.Dir(c.BinaryPath)
}

func environmentOr(name string, fallback string) string {

	if value := strings.TrimRight(os.Getenv(name), `\`); value != "" {
		return value
	}

	return fallback
}
