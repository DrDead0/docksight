# ADR 002: Database Selection

- **Status:** Accepted
- **Date:** 2026-07-23
- **Deciders:** DockSight maintainers

## Context

DockSight needs a durable data store for control-plane concerns such as identities, host inventory, environment grouping, agent registration, permissions, and audit history. The database must integrate cleanly with the NestJS modular monolith and Prisma ORM, while remaining simple to operate for open-source and self-hosted deployments.

## Decision

DockSight uses **PostgreSQL** as the primary relational database, accessed through **Prisma**.

## Why PostgreSQL

### Relational data

Docker management state is inherently relational: users relate to hosts, hosts relate to containers and agents, environments group resources, and notifications reference entities. PostgreSQL’s relational model fits these structures without forcing document-shaped workarounds.

### Permissions

Role-based and resource-scoped access control maps naturally to relational tables and constraints. PostgreSQL supports the integrity guarantees needed for authorization data as DockSight grows beyond a single admin account.

### Audit history

Operational platforms need append-friendly, queryable audit trails (who changed what, when, on which host). PostgreSQL handles chronological records, indexes, and retention-friendly schemas reliably.

### Reliability

PostgreSQL is battle-tested for durability, transactions, and concurrent access. DockSight’s control plane should not lose configuration or audit events during normal process restarts.

### Open-source ecosystem

PostgreSQL is free, widely hosted, and easy to run locally with Docker Compose. That matches DockSight’s open-source posture and keeps contributor onboarding lightweight.

## Consequences

### Positive

- Clear fit for permissions, inventory, and audit use cases
- Strong NestJS + Prisma TypeScript ergonomics
- Portable local and production deployments via Compose or managed Postgres

### Trade-offs

- Schema design and migrations require discipline as domains grow
- Not intended as a high-volume metrics time-series store (those can use specialized paths later)

## Alternatives considered

- **SQLite** — simpler local setup, weaker multi-user concurrency and hosting model
- **MongoDB** — flexible documents, weaker fit for permissions and relational inventory
- **MySQL / MariaDB** — viable RDBMS, but PostgreSQL has stronger default fit for this stack and contributor familiarity in similar OSS tools
