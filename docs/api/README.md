# API Documentation

## HTTP / OpenAPI

OpenAPI/Swagger is served by the NestJS backend at:

```
http://localhost:3000/api/docs
```

As domain HTTP endpoints are implemented, their contracts will be documented here and generated from Swagger annotations.

## Agent WebSocket protocol

All DockSight Server ↔ Agent WebSocket messages are defined in:

- [../protocol.md](../protocol.md)

That document is the source of truth for message envelopes, types, and reserved future domains (containers, logs, metrics, events).
