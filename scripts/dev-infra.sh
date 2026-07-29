#!/usr/bin/env sh
set -e

echo "Starting DockSight local infrastructure (PostgreSQL + Redis)..."
docker compose up -d

echo "Waiting for health checks..."
docker compose ps

echo "Done. Next:"
echo "  npm run db:generate"
echo "  npm run dev:server"
echo "  npm run dev:web"
