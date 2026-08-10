package install

import (
	"context"
	"fmt"
)

// unitTemplate is the systemd service definition.
//
// Restart=always with a delay keeps the agent connected across platform
// restarts and network drops; without RestartSec a crash-looping agent would
// hammer the platform. After=docker.service ensures the Engine socket exists
// before the agent tries to read it, and network-online.target avoids a first
// connection attempt against an interface that has no address yet.
const unitTemplate = `[Unit]
Description=DockSight Agent
Documentation=https://github.com/Open-Source-Kigali/docksight
After=docker.service network.target network-online.target
Wants=network-online.target
Requires=docker.service

[Service]
Type=simple
ExecStart=%s --config %s
Restart=always
RestartSec=5s

# The agent talks to the Docker socket, which is root-owned.
User=root
Group=root

StandardOutput=journal
StandardError=journal
SyslogIdentifier=docksight-agent

[Install]
WantedBy=multi-user.target
`

// RenderUnit builds the service definition for an agent installation.
//
// "Unit" is systemd's word, kept because the installer is written against
// UnitController and both platforms need the same thing from it: one string
// that fully describes the service, so a repeat install can compare it
// against what is already there and know whether anything changed.
//
// On Linux that string is the unit file. On Windows there is no file at all —
// service configuration lives in the Service Control Manager database — and
// the equivalent is the command line the SCM stores and runs. The rest of
// what the unit expresses has no textual form there and is applied directly
// by scm.Manager: automatic start, and failure actions in place of
// Restart=always.
func RenderUnit(layout Layout) string {

	if layout.Windows {
		return RenderCommand(layout)
	}

	return fmt.Sprintf(unitTemplate, layout.BinaryPath, layout.ConfigPath())
}

// RenderCommand builds the command line the Service Control Manager runs.
//
// Both paths are quoted unconditionally. The default install path contains a
// space ("C:\Program Files\DockSight") and the SCM splits an unquoted command
// line on whitespace, so an unquoted executable would have Windows look for
// "C:\Program.exe". Quoting the argument too keeps the rendered string stable
// no matter where a host's ProgramData lives, which is what lets a reinstall
// compare it against the stored one and correctly report no change.
//
// The quotes are plain: unlike the YAML config, a Windows command line does
// not escape the backslashes inside them.
func RenderCommand(layout Layout) string {
	return quote(layout.BinaryPath) + " --config " + quote(layout.ConfigPath())
}

func quote(value string) string {
	return `"` + value + `"`
}

// installService writes the service definition, reloads the service manager
// and starts the service.
//
// Restart rather than start: on a repeat install the service is already
// running the previous binary, and start would be a silent no-op that leaves
// the old version live.
func (i *Installer) installService(ctx context.Context) error {

	i.Report.Step("Creating the " + i.Layout.ServiceName + " service")

	changed, err := i.Units.WriteUnit(i.Layout.ServiceName, RenderUnit(i.Layout))

	if err != nil {
		return err
	}

	if changed {
		i.Report.Success("Service written to " + i.Units.UnitPath(i.Layout.ServiceName))
	} else {
		i.Report.Success("Service already up to date")
	}

	if err := i.Units.DaemonReload(ctx); err != nil {
		return err
	}

	if err := i.Units.Enable(ctx, i.Layout.ServiceName); err != nil {
		return err
	}

	i.Report.Success("Service enabled at boot")

	i.Report.Step("Starting " + i.Layout.ServiceName)

	if err := i.Units.Restart(ctx, i.Layout.ServiceName); err != nil {
		return err
	}

	return nil
}
