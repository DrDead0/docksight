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

## Development setup

### 1. Clone and install

```bash
git clone <repository-url> docksight
cd docksight
cp .env.example .env
cp apps/server/.env.example apps/server/.env
cp apps/web/.env.example apps/web/.env
npm install
```

### 2. Start infrastructure

```bash
npm run docker:infra
```

This starts PostgreSQL and Redis.

### 3. Prepare the database client

```bash
npm run db:generate
```

### 4. Run apps locally

In separate terminals:

```bash
npm run dev:server
npm run dev:web
```

- Web: http://localhost:5173
- API: http://localhost:3000/api/health
- Swagger: http://localhost:3000/docs

### Full stack via Compose

```bash
npm run docker:up
```

## Useful commands

| Command | Description |
| --- | --- |
| `npm run dev:web` | Start Vite dashboard |
| `npm run dev:server` | Start NestJS in watch mode |
| `npm run build` | Build workspaces |
| `npm run docker:infra` | Start Postgres + Redis |
| `npm run docker:up` | Start full Compose stack |
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
