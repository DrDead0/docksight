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

Reserved later: `logs`, `metrics`, `event`, `error`.
