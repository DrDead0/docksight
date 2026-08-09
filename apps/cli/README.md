# DockSight CLI

The `docksight` binary installs and manages DockSight on a host: the platform
stack (`docksight install` / `update`) and the remote agent
(`docksight agent …`). It is a Cobra CLI written in Go.

User-facing command reference: [docs/cli.md](../../docs/cli.md).

## Build

```bash
cd apps/cli
go build -o docksight .
```

Requires the Go version in [`go.mod`](./go.mod) (currently 1.26.5).

## Test

```bash
cd apps/cli
go test ./...
```

## Package layout

| Path | Responsibility |
|------|----------------|
| `cmd/` | Cobra commands and the `consoleReporter` that talks to `ui` |
| `cmd/internal/agent/` | Agent install/update phases (download, config, systemd, verify) |
| `cmd/internal/buildinfo/` | Version string injected at build time |
| `cmd/internal/compose/` | Docker Compose runner for the platform stack |
| `cmd/internal/env/` | Credential / `.env` generation |
| `cmd/internal/filesystem/` | Directory setup and file copy helpers |
| `cmd/internal/installer/` | Platform install orchestration |
| `cmd/internal/progress/` | `Reporter` interface (no presentation) |
| `cmd/internal/release/` | GitHub Release asset discovery and extraction |
| `cmd/internal/selfinstall/` | Atomic CLI binary self-replacement |
| `cmd/internal/state/` | Install `state.json` |
| `cmd/internal/system/` | Host preflight checks (`PlatformRequirements`, `AgentRequirements`) |
| `cmd/internal/systemd/` | Unit file write + `systemctl` / `journalctl` manager |
| `cmd/internal/ui/` | Banner, success/error/info printing |
| `cmd/config/` | Default platform configuration |

### Rule: `internal/` packages never import `ui`

Business-logic packages under `cmd/internal/` report progress through the
`progress.Reporter` interface only. The `cmd` layer supplies
`consoleReporter`, which is the only bridge into `ui`. New code that prints
from an `internal` package breaks that boundary — put the print in `cmd` or
report through `Reporter`.

## Platform notes

Some commands (systemd unit management, Docker socket access) only work on
Linux with root privileges. A Linux VM or cloud VPS is the expected
development environment for those paths.
