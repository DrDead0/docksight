#!/bin/sh
# Applies pending database migrations, then starts the process passed as CMD.
#
# `migrate deploy` is the deployment-safe command: it only replays migrations
# that are already committed to the repository, and it never resets or
# generates anything. Set RUN_MIGRATIONS=false to skip it (for example when a
# separate job owns schema changes).
set -e

if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
  echo "[entrypoint] applying database migrations…"
  cd /app/apps/server
  npx prisma migrate deploy
  cd /app
  echo "[entrypoint] migrations up to date"
else
  echo "[entrypoint] RUN_MIGRATIONS=false — skipping migrations"
fi

exec "$@"
