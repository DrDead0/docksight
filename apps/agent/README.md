# DockSight Agent

Lightweight Go agent that runs beside a Docker host, registers with the DockSight backend, and executes Docker read/lifecycle/log operations.

## Current capabilities

- Bootstrap: config, identity, logging, graceful shutdown
- WebSocket registration + 30s heartbeats + reconnect
- Container discovery (`container.list` / `container.listed`)
- Lifecycle via Docker Go SDK (`start` / `stop` / `restart`)
- Log streaming (`logs.subscribe` / `logs.chunk` / `logs.unsubscribe`)
  - Last N lines, then live follow
  - Multiplexed stdout/stderr decoding
  - Batched chunks (≤50 lines or ≤200ms)
  - Multiple concurrent streams keyed by `requestId`

Not implemented yet: remove container, metrics, auth/TLS.

## Structure

```
agent/
├── cmd/agent/
├── config.yaml
├── data/
└── internal/
    ├── app/
    ├── communication/   # WebSocket client
    ├── config/
    ├── docker/          # Docker Engine SDK wrapper
    ├── identity/
    ├── lifecycle/
    ├── logger/
    ├── logs/            # Log stream service (subscribe/batch/decode)
    └── version/
```

## Run

```bash
docker compose up -d
npm run dev:server
cd agent && go run ./cmd/agent
```

## Tests

```bash
cd agent
go test ./internal/logs/...
```

Protocol reference: [docs/protocol.md](../docs/protocol.md)
