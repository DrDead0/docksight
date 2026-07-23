# @docksight/protocol

Shared **agent communication protocol** contracts for DockSight.

This package is the TypeScript source of truth for WebSocket message shapes exchanged between:

- `apps/server` (NestJS)
- `agent` (Go — mirrors these contracts; see [docs/protocol.md](../../docs/protocol.md))

It defines types and constants only. It does **not** open sockets, talk to Docker, or touch the database.

## Purpose

Keep server and agent aligned on:

- One message envelope
- Stable `domain.action` type names
- Typed payloads for agent registration and heartbeats

Human-readable protocol rules (lifecycle, reserved future domains, versioning) live in:

- [docs/protocol.md](../../docs/protocol.md)

## Message envelope

Every WebSocket JSON message uses:

```ts
{
  type: string   // domain.action, e.g. "agent.register"
  payload: object // always a JSON object — never null/array/primitive
}
```

Helpers:

- `createEnvelope(type, payload)`
- `isMessageEnvelope(value)`
- `isJsonObject(value)`

## Supported messages (v0.1)

| Constant | Type string | Direction |
| --- | --- | --- |
| `AGENT_REGISTER` | `agent.register` | Agent → Server |
| `AGENT_REGISTERED` | `agent.registered` | Server → Agent |
| `AGENT_HEARTBEAT` | `agent.heartbeat` | Agent → Server |

### `agent.register`

```ts
{
  uuid: string
  hostname: string
  os: string
  architecture: string
  version: string
}
```

### `agent.registered`

```ts
{
  id: string
  uuid: string
  status: string
  message: string
}
```

### `agent.heartbeat`

```ts
{
  uuid: string
}
```

### Status values

- `AGENT_STATUS.ONLINE`
- `AGENT_STATUS.OFFLINE`
- `AGENT_STATUS.UNKNOWN`

Also exported as the `AgentStatus` string-union type.

## How `apps/server` should use it

```ts
import {
  AGENT_REGISTER,
  AGENT_REGISTERED,
  AGENT_HEARTBEAT,
  AGENT_STATUS,
  createEnvelope,
  isMessageEnvelope,
  type AgentRegisterPayload,
  type AgentRegisteredPayload,
} from '@docksight/protocol'

// Parse incoming WS JSON, then narrow on envelope.type
if (envelope.type === AGENT_REGISTER) {
  const payload = envelope.payload as AgentRegisterPayload
  // ...
}

const response = createEnvelope(AGENT_REGISTERED, {
  id: agent.id,
  uuid: agent.uuid,
  status: AGENT_STATUS.ONLINE,
  message: 'Registration successful',
})
```

Add the workspace dependency in `apps/server/package.json`:

```json
"@docksight/protocol": "*"
```

Then replace local duplicates under `src/agents/messages.ts` with imports from this package.

## How the Go agent should use it

The Go agent cannot import this TypeScript package directly. It SHOULD:

1. Treat [docs/protocol.md](../../docs/protocol.md) + this package as the contract
2. Keep Go structs/`const` values in sync (`agent.register`, etc.)
3. Prefer the same field names and JSON tags

## Extensibility

Reserved future domains (not implemented in this package yet):

- `container`
- `logs`
- `metrics`
- `event`
- `error`

Add new files (e.g. `src/container.ts`) and re-export from `src/index.ts` when those features land. Do not break the `{ type, payload }` envelope.

## Install / build

From the monorepo root:

```bash
npm install
npm run build --workspace=@docksight/protocol
```
