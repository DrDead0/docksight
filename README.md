# DockSight

Open-source Docker management and observability platform.

DockSight helps developers and teams visually manage Docker environments, monitor containers, inspect logs, organize environments, and later connect multiple Docker hosts through lightweight agents.

> Foundation stage: repository structure and configuration only. Business features are not implemented yet.

## Architecture overview

DockSight is a monorepo with a modular-monolith backend:

| Path | Role |
| --- | --- |
| `apps/web` | React + Vite dashboard |
| `apps/server` | NestJS API (Prisma, Redis, WebSockets, Swagger) |
| `agent` | Go Docker host agent (structure only) |
| `packages/*` | Shared types, config, and utilities |
| `infrastructure/` | Docker and deployment assets |
| `docs/` | Architecture and ADRs |

```
Web (React) ──HTTP/WS──► Server (NestJS modular monolith)
                              │
                    PostgreSQL + Redis
                              │
                    Go agents (future multi-host)
```

See [docs/architecture.md](docs/architecture.md) for details.

## Stack

- **Frontend:** React, TypeScript, Vite, Tailwind CSS, shadcn/ui, React Query, Zustand, Lucide
- **Backend:** NestJS, TypeScript, Prisma, PostgreSQL, Redis, WebSockets, Swagger
- **Agent:** Go (Docker Engine API + WebSocket planned)
- **Infra:** Docker Compose

## Prerequisites

- Node.js 20+
- npm 10+
- Docker and Docker Compose
- Go 1.22+ (only needed for agent work)

## Development Environment

DockSight’s Compose stack runs **infrastructure only** (PostgreSQL + Redis). The frontend, backend, and Go agent stay on your host machine during development.

### 1. Start infrastructure

```bash
cp .env.example .env
docker compose up -d
```

This starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

Both services join the `docksight-network` network and use named volumes `postgres-data` and `redis-data`.

### 2. Verify

```bash
docker ps
```

You should see `docksight-postgres` and `docksight-redis` (healthy).

### 3. Stop

```bash
docker compose down
```

Containers stop; volumes are kept.

### 4. Remove volumes

```bash
docker compose down -v
```

Stops containers and deletes `postgres-data` / `redis-data` (all local DB/cache data).

### Run apps on the host

```bash
npm install
npm run db:generate
npm run dev:server
npm run dev:web
```

- Web: http://localhost:5173
- API / Swagger: http://localhost:3000/api/docs

Point `apps/server/.env` `DATABASE_URL` at:

`postgresql://docksight:change_me@localhost:5432/docksight?schema=public`

(or match the values in the root `.env`).

## Useful commands

| Command | Description |
| --- | --- |
| `npm run dev:web` | Start Vite dashboard |
| `npm run dev:server` | Start NestJS in watch mode |
| `npm run build` | Build workspaces |
| `npm run docker:up` / `docker:infra` | Start Postgres + Redis |
| `npm run docker:down` | Stop Compose stack |
| `npm run db:generate` | Generate Prisma client |
| `npm run db:migrate` | Run Prisma migrations |
| `npm run db:studio` | Open Prisma Studio |

Agent (when Go is installed):

```bash
cd agent
go run ./cmd/agent
```

## Repository structure

```
docksight/
├── apps/
│   ├── web/                 # React dashboard
│   └── server/              # NestJS backend
├── agent/                   # Go agent
├── packages/
│   ├── types/
│   ├── config/
│   └── utils/
├── infrastructure/
│   ├── docker/
│   └── deployment/
├── docs/
│   ├── architecture.md
│   ├── decisions/
│   └── api/
├── docker-compose.yml
├── README.md
├── LICENSE
└── CONTRIBUTING.md
```

## Future roadmap

1. Authentication and user management
2. Local Docker host connection and container inventory
3. Live logs and basic metrics
4. Environment grouping and notifications
5. Multi-host Go agents with secure WebSocket protocol
6. Audit trail and team collaboration features

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
