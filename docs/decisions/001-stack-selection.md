# ADR 001: Stack Selection

- **Status:** Accepted
- **Date:** 2026-07-23
- **Deciders:** DockSight maintainers

## Context

DockSight needs a stack that supports a visual Docker management UI, a durable backend API, realtime updates, and a future multi-host agent model — without premature microservice complexity.

## Decision

### React (frontend)

React provides a mature ecosystem for interactive dashboards, strong TypeScript support, and wide contributor familiarity. Paired with Vite, React Query, Zustand, Tailwind, and shadcn/ui, it enables a fast feedback loop and a maintainable component system.

### NestJS (backend)

NestJS offers structured modules, dependency injection, first-class TypeScript, and built-in support for REST, Swagger, and WebSockets. This fits a modular monolith: domain folders (`hosts`, `containers`, `agents`, etc.) can grow independently while remaining one deployable service.

### PostgreSQL

PostgreSQL is a reliable relational database for users, hosts, environments, audit trails, and future authorization models. Prisma keeps schema evolution and typed data access aligned with the TypeScript backend.

### Redis

Redis covers caching, short-lived session/token data, and pub/sub fan-out for realtime events (logs, metrics, agent heartbeats). The connection is configured early so later features do not require infrastructure redesign.

### Go agent

The agent must sit close to Docker Engine, stream telemetry, and execute management commands with low overhead. Go is a strong fit for long-running system agents, static binaries, and Docker API clients. Keeping the agent separate from the NestJS monolith preserves isolation and multi-host scalability.

## Consequences

### Positive

- Clear separation between dashboard, control plane, and host agents
- Strong typing across web and server
- Room to scale domains inside the monolith before any service split
- Local development can mirror production topology via Docker Compose

### Negative / trade-offs

- Contributors need familiarity with TypeScript and Go
- Redis adds an infrastructure dependency even before heavy caching use
- Modular monolith discipline is required to avoid a tangled codebase

## Alternatives considered

- **Next.js fullstack** — rejected to keep UI and API deployables and concerns separate
- **Express-only backend** — rejected for weaker modular conventions at expected scale
- **Kubernetes-first design** — rejected; DockSight is Docker-focused
- **Microservices from day one** — rejected as premature operational complexity
