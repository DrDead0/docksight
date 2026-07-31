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
Documentation=https://github.com/rodriguecyber/docksight
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

// RenderUnit builds the systemd unit for an agent installation.
func RenderUnit(layout Layout) string {

	return fmt.Sprintf(unitTemplate, layout.BinaryPath, layout.ConfigPath())
}

// installService writes the unit, reloads systemd and starts the service.
//
// Restart rather than start: on a repeat install the service is already
// running the previous binary, and start would be a silent no-op that leaves
// the old version live.
func (i *Installer) installService(ctx context.Context) error {

	i.Report.Step("Creating the systemd service")

	changed, err := i.Units.WriteUnit(i.Layout.ServiceName, RenderUnit(i.Layout))

	if err != nil {
		return err
	}

	if changed {
		i.Report.Success("Unit written to " + i.Units.UnitPath(i.Layout.ServiceName))
	} else {
		i.Report.Success("Unit already up to date")
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
