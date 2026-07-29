# Contributing to DockSight

Thanks for your interest in DockSight.

## Ground rules

- Keep changes focused and well-scoped.
- Prefer the existing modular monolith boundaries in `apps/server`.
- Do not introduce Kubernetes abstractions or microservice splits without an ADR.
- Discuss larger architectural changes in an issue or ADR under `docs/decisions/`.

## Development workflow

1. Fork and clone the repository.
2. Copy `.env.example` files and install dependencies (`npm install`).
3. Start local infra: `npm run docker:infra`.
4. Run `npm run dev:server` and `npm run dev:web`.
5. Create a branch for your change.
6. Open a pull request with a clear summary and test notes.

## Project conventions

- **Frontend:** feature folders under `apps/web/src/features`, shared UI under `components`.
- **Backend:** one NestJS module per domain under `apps/server/src/<domain>`.
- **Agent:** keep Docker and communication logic inside `agent/internal`.
- **Docs:** update architecture docs when boundaries change.

## Commit style

Use concise commit messages that explain why the change exists.

Examples:

- `add NestJS health module foundation`
- `configure Vite path aliases for shadcn`

## Code of conduct

Be respectful and constructive. Harassment or discrimination is not tolerated.
