#!/usr/bin/env sh
set -e

echo "Starting DockSight full Docker Compose stack..."
docker compose up --build -d
docker compose ps
