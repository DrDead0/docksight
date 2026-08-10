package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLinuxDefault(t *testing.T) {

	cfg := LinuxDefault()

	expected := map[string]string{
		"install":  "/opt/docksight",
		"data":     "/var/lib/docksight",
		"binary":   "/usr/local/bin/docksight",
		"compose":  "/opt/docksight/" + ComposeFileName,
		"env":      "/opt/docksight/" + EnvFileName,
		"state":    "/opt/docksight/" + StateFileName,
		"bindir":   "/usr/local/bin",
		"envexamp": "/opt/docksight/" + EnvExampleFileName,
	}

	got := map[string]string{
		"install":  cfg.InstallationDir,
		"data":     cfg.DataDir,
		"binary":   cfg.BinaryPath,
		"compose":  filepath.ToSlash(cfg.ComposePath()),
		"env":      filepath.ToSlash(cfg.EnvPath()),
		"state":    filepath.ToSlash(cfg.StatePath()),
		"bindir":   filepath.ToSlash(cfg.InstallDirectory()),
		"envexamp": filepath.ToSlash(cfg.EnvExamplePath()),
	}

	for name, want := range expected {

		if got[name] != want {
			t.Errorf("%s is %q, want %q", name, got[name], want)
		}
	}

	if cfg.Port != 2002 {
		t.Errorf("port is %d", cfg.Port)
	}
}

func TestWindowsDefault(t *testing.T) {

	t.Setenv("ProgramData", `C:\ProgramData`)
	t.Setenv("ProgramFiles", `C:\Program Files`)

	cfg := WindowsDefault()

	// The binary and the installation are deliberately not in the same
	// place: one is a program, the other is machine-wide state.
	if cfg.InstallationDir != `C:\ProgramData\DockSight\platform` {
		t.Errorf("installation dir is %q", cfg.InstallationDir)
	}

	if cfg.DataDir != `C:\ProgramData\DockSight\data` {
		t.Errorf("data dir is %q", cfg.DataDir)
	}

	if cfg.BinaryPath != `C:\Program Files\DockSight\docksight.exe` {
		t.Errorf("binary path is %q", cfg.BinaryPath)
	}

	if cfg.Port != 2002 {
		t.Errorf("port is %d", cfg.Port)
	}

	// The directory that has to reach the machine PATH. Derived with
	// filepath, so it is only meaningful on the host the layout describes —
	// the literal fields above are what this test pins everywhere.
	if runtime.GOOS == "windows" && cfg.InstallDirectory() != `C:\Program Files\DockSight` {
		t.Errorf("install directory is %q", cfg.InstallDirectory())
	}
}

// A machine that keeps ProgramData somewhere else must be honoured, and a
// trailing separator in the variable must not double up.
func TestWindowsDefaultHonoursEnvironment(t *testing.T) {

	t.Setenv("ProgramData", `D:\State\`)
	t.Setenv("ProgramFiles", `D:\Apps`)

	cfg := WindowsDefault()

	if !strings.HasPrefix(cfg.InstallationDir, `D:\State\DockSight`) {
		t.Errorf("installation dir is %q", cfg.InstallationDir)
	}

	if strings.Contains(cfg.InstallationDir, `\\`) {
		t.Errorf("doubled separator in %q", cfg.InstallationDir)
	}

	if !strings.HasPrefix(cfg.BinaryPath, `D:\Apps\DockSight`) {
		t.Errorf("binary path is %q", cfg.BinaryPath)
	}
}

func TestDefaultMatchesHost(t *testing.T) {

	windows := strings.Contains(Default().BinaryPath, ".exe")

	if windows != (runtime.GOOS == "windows") {
		t.Fatalf("Default() returned a %s layout on %s", map[bool]string{true: "windows", false: "unix"}[windows], runtime.GOOS)
	}
}
