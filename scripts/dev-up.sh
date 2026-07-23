#!/usr/bin/env sh
set -e

echo "Starting DockSight infrastructure (PostgreSQL + Redis)..."
docker compose up -d
docker compose ps
