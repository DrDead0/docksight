# @docksight/protocol

Shared **agent communication protocol** contracts for DockSight.

Used by:

- `apps/server` (NestJS)
- `agent` (Go mirrors these contracts; see [docs/protocol.md](../../docs/protocol.md))

## Envelope

```ts
{ type: string; payload: object }
```

- `type` is lowercase `domain.action`
- `payload` is always a JSON object

## Supported messages

| Constant | Type |
| --- | --- |
| `AGENT_REGISTER` | `agent.register` |
| `AGENT_REGISTERED` | `agent.registered` |
| `AGENT_HEARTBEAT` | `agent.heartbeat` |
| `CONTAINER_LIST` | `container.list` |
| `CONTAINER_LISTED` | `container.listed` |
| `CONTAINER_START` | `container.start` |
| `CONTAINER_STOP` | `container.stop` |
| `CONTAINER_RESTART` | `container.restart` |
| `CONTAINER_RESULT` | `container.result` |
| `LOGS_SUBSCRIBE` | `logs.subscribe` |
| `LOGS_CHUNK` | `logs.chunk` |
| `LOGS_UNSUBSCRIBE` | `logs.unsubscribe` |

Lifecycle and log streams always include `requestId` for correlation.

Reserved later: `metrics`, `event`, `error`, `container.remove`.
