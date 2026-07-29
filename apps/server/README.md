# DockSight Server

NestJS backend for DockSight — open-source Docker management and observability.

## Stack

- NestJS
- TypeScript (strict)
- Prisma ORM
- PostgreSQL
- Redis (client placeholder)
- WebSocket support (infrastructure only)
- Swagger

## Install dependencies

From the monorepo root:

```bash
npm install
```

Or from this package:

```bash
cd apps/server
npm install
```

## Environment

```bash
cp .env.example .env
```

Fill in:

```env
DATABASE_URL=postgresql://docksight:docksight@localhost:5432/docksight?schema=public
REDIS_HOST=localhost
REDIS_PORT=6379
PORT=3000
```

## Database (PostgreSQL + Prisma)

```bash
cp .env.example .env
# set DATABASE_URL (or DATABASE_HOST/PORT/USER/PASSWORD/NAME)
npx prisma generate
```

Connection URL for Prisma CLI lives in `prisma.config.ts` (`DATABASE_URL`).
NestJS loads the same values through `ConfigModule` + `databaseConfig`.

```bash
npm run prisma:generate
npm run prisma:migrate
npm run prisma:studio
```

## Run development server

```bash
npm run start:dev
```

- API base: `http://localhost:3000/api`
- Swagger: `http://localhost:3000/api/docs`

## Project structure

```
src/
├── auth/            # (empty — future)
├── users/           # (empty — future)
├── hosts/           # (empty — future)
├── agents/          # (empty — future)
├── containers/      # (empty — future)
├── logs/            # (empty — future)
├── metrics/         # (empty — future)
├── environments/    # (empty — future)
├── notifications/   # (empty — future)
├── audit/           # (empty — future)
└── common/          # Prisma, Redis, WebSocket foundation
```
