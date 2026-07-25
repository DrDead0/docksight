# DockSight Agent

Lightweight Go agent that runs beside a Docker host, registers with the DockSight backend, and executes read + lifecycle Docker operations.

## Current capabilities

- Bootstrap: config, identity, logging, graceful shutdown
- WebSocket registration with DockSight Server (`/agents`)
- Heartbeats every 30s + automatic reconnect
- Docker Engine discovery (`GetDockerInfo`, `ListContainers`)
- Container lifecycle via Docker Go SDK:
  - `StartContainer` / `StopContainer` / `RestartContainer`
  - Handles `container.start` | `container.stop` | `container.restart`
  - Replies with `container.result` (includes `requestId`)

Not implemented yet: remove container, logs streaming, metrics, auth/TLS.

## Structure

```
agent/
├── cmd/agent/              # Entrypoint
├── config.yaml             # Local runtime config
├── data/                   # Persisted identity (gitignored)
└── internal/
    ├── app/                # Startup orchestration
    ├── communication/      # WebSocket client
    ├── config/             # YAML config loader
    ├── docker/             # Docker Engine SDK client
    ├── identity/           # UUID create/load
    ├── lifecycle/          # SIGINT/SIGTERM handling
    ├── logger/             # Structured logging
    └── version/            # Agent version metadata
```

## Run

```bash
docker compose up -d
npm run dev:server
cd agent && go run ./cmd/agent
```

Protocol reference: [docs/protocol.md](../docs/protocol.md)
