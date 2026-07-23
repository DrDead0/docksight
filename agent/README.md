# DockSight Agent

Lightweight Go agent that connects Docker hosts to the DockSight backend.

## Status

Foundation only. Docker Engine API integration and WebSocket communication
are intentionally not implemented yet.

## Structure

```
agent/
├── cmd/agent/           # Entrypoint
└── internal/
    ├── docker/          # Docker Engine API client (future)
    ├── communication/   # Backend WebSocket gateway (future)
    ├── collector/       # Metrics and status collection (future)
    ├── executor/        # Command execution (future)
    ├── config/          # Configuration loading
    └── security/        # Auth and message validation (future)
```

## Local development

Requires Go 1.22+.

```bash
cd agent
go run ./cmd/agent
```
