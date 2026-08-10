package install

import (
	"strings"
	"testing"
)

func TestRenderUnitForLinux(t *testing.T) {

	layout := LinuxLayout()

	rendered := RenderUnit(layout)

	for _, required := range []string{
		"[Unit]",
		"[Service]",
		"Restart=always",
		"RestartSec=5s",
		"ExecStart=/usr/local/bin/docksight-agent --config /etc/docksight-agent/config.yaml",
		"WantedBy=multi-user.target",
	} {

		if !strings.Contains(rendered, required) {
			t.Errorf("unit is missing %q:\n%s", required, rendered)
		}
	}
}

// On Windows the service definition is the command line the SCM stores, and
// both paths must be quoted: the default install path contains a space, and
// an unquoted argument would make the rendered string depend on where a
// host's ProgramData happens to be.
func TestRenderUnitForWindows(t *testing.T) {

	t.Setenv("ProgramFiles", `C:\Program Files`)
	t.Setenv("ProgramData", `C:\ProgramData`)

	layout := WindowsLayout()

	rendered := RenderUnit(layout)

	want := `"C:\Program Files\DockSight\docksight-agent.exe" ` +
		`--config "C:\ProgramData\DockSight\agent\config.yaml"`

	if rendered != want {
		t.Fatalf("got  %s\nwant %s", rendered, want)
	}

	// A Windows command line does not escape backslashes inside quotes, and
	// doubling them would send the SCM looking for a path that does not exist.
	if strings.Contains(rendered, `\\`) {
		t.Errorf("backslashes were escaped into the command line: %s", rendered)
	}

	if strings.Contains(rendered, "[Service]") {
		t.Errorf("a systemd unit was rendered for a Windows host:\n%s", rendered)
	}
}

// The rendered definition must be stable, or every reinstall would look like
// a change and rewrite a service that had not moved.
func TestRenderUnitIsStable(t *testing.T) {

	for _, layout := range []Layout{LinuxLayout(), WindowsLayout()} {

		if RenderUnit(layout) != RenderUnit(layout) {
			t.Errorf("rendering is not stable for windows=%v", layout.Windows)
		}
	}
}
