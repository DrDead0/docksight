# DockSight

Open-source Docker management and observability platform.

DockSight helps developers and teams visually manage Docker environments, monitor containers, inspect logs, organize environments, and later connect multiple Docker hosts through lightweight agents.

> Current increment: container start / stop / restart from the dashboard.

## Architecture overview

DockSight is a monorepo with a modular-monolith backend:

| Path | Role |
| --- | --- |
| `apps/web` | React + Vite dashboard |
| `apps/server` | NestJS API (Prisma, Redis, WebSockets, Swagger) |
| `agent` | Go Docker host agent (discovery + lifecycle) |
| `packages/*` | Shared types, protocol, config, and utilities |
| `infrastructure/` | Docker and deployment assets |
| `docs/` | Architecture and ADRs |

```
Web (React) ──HTTP/WS──► Server (NestJS modular monolith)
                              │
                    PostgreSQL + Redis
                              │
                    Go agents ◄──WebSocket──► /agents
                              │
                         Docker Engine
```

See [docs/architecture.md](docs/architecture.md) for details.

## Stack

- **Frontend:** React, TypeScript, Vite, Tailwind CSS, shadcn/ui, React Query, Zustand, Lucide
- **Backend:** NestJS, TypeScript, Prisma, PostgreSQL, Redis, WebSockets, Swagger
- **Agent:** Go (Docker Engine API + WebSocket registration/discovery)
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

### Run the agent (discovery + lifecycle)

```bash
# Terminal A — infrastructure
docker compose up -d

# Terminal B — backend
npm run dev:server

# Terminal C — agent
cd agent
go run ./cmd/agent

# Terminal D — web
npm run dev:web
```

Expected:

1. Agent connects to `ws://localhost:3000/agents`
2. Agent registers (`agent.register` → `agent.registered`)
3. Server sends `container.list`
4. Agent replies with `container.listed`
5. Heartbeats continue every 30s
6. Dashboard Start/Stop/Restart sends REST → WS command → `container.result`

Protocol reference: [docs/protocol.md](docs/protocol.md)

- Web: http://localhost:5173
- API / Swagger: http://localhost:3000/api/docs
- Agent WS: `ws://localhost:3000/agents`

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
