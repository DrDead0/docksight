# DockSight Architecture

## Project vision

DockSight is an open-source Docker management and observability platform. It helps developers and teams visually manage Docker environments, monitor containers, inspect logs, organize environments, and eventually connect multiple Docker hosts through lightweight agents.

DockSight focuses exclusively on Docker. It does not target Kubernetes or other container orchestration platforms.

## High-level architecture

DockSight is organized as an npm workspace monorepo with three primary runtime parts and shared packages:

```
┌─────────────────┐       HTTP / WS        ┌──────────────────────┐
│  apps/web       │ ◄────────────────────► │  apps/server         │
│  React dashboard│                        │  NestJS modular      │
└─────────────────┘                        │  monolith            │
                                           └──────────┬───────────┘
                                                      │
                         ┌────────────────────────────┼────────────────────────────┐
                         │                            │                            │
                         ▼                            ▼                            ▼
                  ┌─────────────┐              ┌─────────────┐              ┌─────────────┐
                  │ PostgreSQL  │              │    Redis    │              │ Go agents   │
                  │ (source of  │              │ (cache /    │              │ (future host│
                  │  truth)     │              │  pub-sub)   │              │  connectors)│
                  └─────────────┘              └─────────────┘              └─────────────┘
```

## Component responsibilities

### `apps/web`

- Operator-facing React dashboard
- Uses React Query for server state and Zustand for client UI state
- Consumes REST and WebSocket APIs from the backend
- UI primitives via Tailwind CSS and shadcn/ui

### `apps/server`

- NestJS modular monolith (single deployable backend)
- Domain modules: auth, users, hosts, agents, containers, logs, metrics, environments, notifications, audit
- Shared infrastructure under `common/` (Prisma, Redis, WebSocket, health)
- PostgreSQL via Prisma ORM
- Redis as a connection placeholder for caching and realtime fan-out
- Swagger at `/docs`

### `agent`

- Go process intended to run beside a Docker Engine
- Future responsibilities: Docker Engine API integration, metrics collection, command execution, secure WebSocket communication with the backend
- Structure prepared; Docker communication not implemented yet

### `packages/`

- `types` — shared TypeScript contracts
- `config` — shared defaults and constants
- `utils` — small cross-app helpers

### `infrastructure/`

- Docker images and compose-related assets for local and future deployment workflows

## Technology choices

| Layer | Choice | Role |
| --- | --- | --- |
| Frontend | React + TypeScript + Vite | Fast, typed dashboard development |
| UI | Tailwind CSS + shadcn/ui + Lucide | Consistent, composable interface |
| Client state | React Query + Zustand | Server cache + lightweight UI state |
| Backend | NestJS modular monolith | Clear module boundaries without microservice overhead |
| Data | PostgreSQL + Prisma | Relational source of truth with typed access |
| Cache / realtime support | Redis | Sessions, caching, pub/sub (wired as placeholder) |
| Agent | Go | Efficient local Docker host integration |
| Local infra | Docker Compose | Reproducible PostgreSQL, Redis, app services |

## Architectural principles

1. **Modular monolith first** — grow domain modules inside one NestJS app until extraction is justified.
2. **Docker-only scope** — no Kubernetes abstractions in the core product.
3. **Agents for multi-host** — remote Docker hosts connect later via Go agents, not by embedding host credentials in the dashboard alone.
4. **Foundation before features** — structure, configuration, and contracts land before business logic.
5. **Clean boundaries** — web, server, agent, and shared packages remain independently understandable.
6. **Stable agent protocol** — server↔agent WebSocket messages share one JSON envelope; see [protocol.md](./protocol.md).

## Current status

This repository currently contains the project foundation only:

- Monorepo layout
- Frontend and backend scaffolding
- Agent folder structure
- Local Docker Compose stack
- Architecture and ADR documentation

Authentication, Docker management logic, and multi-host agent protocols are intentionally deferred.
