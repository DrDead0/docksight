# DockSight Web

React dashboard for DockSight — open-source Docker management and observability.

## Stack

- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- React Query
- React Router
- Lucide React

## Install dependencies

From the monorepo root:

```bash
npm install
```

## Environment

```bash
cp .env.example .env
```

```env
VITE_API_URL=http://localhost:3000/api
```

`VITE_API_URL` is the NestJS API base URL used by the dashboard client.

## Run development server

```bash
npm run dev
```

App: `http://localhost:5173`

## Dashboard feature

Route `/` renders the first read-only dashboard:

- **Hosts** — cards for registered agents (`GET /api/hosts`)
- **Containers** — table for the selected host (`GET /api/hosts/:id/containers`)

No authentication or container actions in this increment.

### Local stack

```bash
# Terminal A — Postgres + Redis
docker compose up -d

# Terminal B — API
npm run dev:server

# Terminal C — Agent (optional, for live host/container data)
cd agent && go run ./cmd/agent

# Terminal D — Web
npm run dev:web
```

## Project structure

```
src/
├── app/                 # App shell, providers, routing
├── components/          # Shared UI (StatusBadge, HostCard, ContainerTable, shadcn)
├── features/dashboard/  # Dashboard page
├── hooks/               # useHosts, useContainers
├── services/            # API client + hosts service
├── types/               # Frontend API types
└── lib/                 # Shared utilities (e.g. cn)
```
