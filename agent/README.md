# DockSight Agent

Lightweight Go agent that runs beside a Docker host and registers with the DockSight backend.

## Current status (PI-007)

- Bootstrap, config, identity, logging, lifecycle
- WebSocket registration with DockSight Server
- Heartbeats every 30s + automatic reconnect

Docker Engine API integration is **not** implemented yet.

## Structure

```
agent/
├── cmd/agent/
├── config.yaml
└── internal/
    ├── app/
    ├── communication/   # WebSocket client (register + heartbeat)
    ├── config/
    ├── identity/
    ├── lifecycle/
    ├── logger/
    └── version/
```

## Run

1. Start infra + server:

```bash
docker compose up -d
npm run dev:server
```

2. Start the agent:

```bash
cd agent
go run ./cmd/agent
```

WebSocket URL defaults to `ws://localhost:3000/agents` (`server.url` in `config.yaml`).
