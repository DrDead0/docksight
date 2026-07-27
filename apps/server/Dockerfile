# syntax=docker/dockerfile:1
#
# DockSight API server.
# Build from the REPOSITORY ROOT, not from this directory:
#   docker build -f apps/server/Dockerfile -t docksight/server .
#
# This is an npm workspaces monorepo. The dependency tree can only be resolved
# when every workspace manifest is present, so all of them are copied before
# `npm ci`. Copying only the root package.json installs nothing for apps/* and
# packages/*, which is why `nest` and `tsc` were "not found" at build time.

# ---- builder: install the workspace tree, then compile ----------------------
# Install and build share one stage on purpose. npm does not hoist every
# package to the root (e.g. @nestjs/cli lands in apps/server/node_modules), so
# copying only /app/node_modules between stages silently loses binaries. The
# manifests are still copied before the install, so the dependency layer stays
# cached until a package.json actually changes.
FROM node:22-alpine AS builder
WORKDIR /app

# bcrypt is a native addon. It currently ships a linux-x64 musl prebuild, so
# nothing is compiled today — this toolchain is a safety net so a future native
# dependency cannot break the build. It stays in this stage and is never
# shipped in the final image.
RUN apk add --no-cache python3 make g++

COPY package.json package-lock.json ./
COPY apps/server/package.json apps/server/
COPY apps/web/package.json apps/web/
COPY packages/config/package.json packages/config/
COPY packages/protocol/package.json packages/protocol/
COPY packages/types/package.json packages/types/
COPY packages/utils/package.json packages/utils/

# packages/protocol declares a `prepare` script (tsc). npm runs a workspace's
# own prepare during install even under --ignore-scripts, so its sources have
# to be here or the install fails. They are small and change rarely, which
# keeps this layer cacheable.
COPY packages/protocol/tsconfig.json packages/protocol/
COPY packages/protocol/src packages/protocol/src

RUN npm ci

COPY . .

# The server imports @docksight/protocol, whose package main points at dist/,
# so the protocol package must be compiled first.
RUN npm run build --workspace=@docksight/protocol

# Generate the Prisma client inside the image. The host's generated/ directory
# is excluded by .dockerignore on purpose — never ship an artifact built on a
# different machine.
#
# prisma.config.ts resolves DATABASE_URL eagerly and throws when it is unset.
# `generate` never opens a connection, so a placeholder satisfies the config
# loader without baking a real credential into the image.
RUN DATABASE_URL="postgresql://build:build@localhost:5432/build?schema=public" \
    npm run prisma:generate --workspace=@docksight/server

RUN npm run build --workspace=@docksight/server

# Drop devDependencies from node_modules so the runtime copy stays small.
RUN npm prune --omit=dev

# ---- runner: minimal runtime ------------------------------------------------
FROM node:22-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production \
    PORT=8080

# tini reaps zombies and forwards signals, so `docker stop` shuts Nest down
# cleanly instead of waiting out the kill timeout.
RUN apk add --no-cache tini

COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

# node_modules/@docksight/* are symlinks into the workspace directories.
# Copying each manifest keeps every link resolvable; only protocol is actually
# imported by the server, and it is the only one that needs its build output.
COPY --from=builder /app/packages/config/package.json ./packages/config/package.json
COPY --from=builder /app/packages/types/package.json ./packages/types/package.json
COPY --from=builder /app/packages/utils/package.json ./packages/utils/package.json
COPY --from=builder /app/packages/protocol/package.json ./packages/protocol/package.json
COPY --from=builder /app/packages/protocol/dist ./packages/protocol/dist
COPY --from=builder /app/apps/web/package.json ./apps/web/package.json

COPY --from=builder /app/apps/server/package.json ./apps/server/package.json
# npm does not hoist everything: apps/server/node_modules holds real packages.
COPY --from=builder /app/apps/server/node_modules ./apps/server/node_modules
COPY --from=builder /app/apps/server/dist ./apps/server/dist
# Schema, migrations and config are needed by `prisma migrate deploy` on boot.
COPY --from=builder /app/apps/server/prisma ./apps/server/prisma
COPY --from=builder /app/apps/server/prisma.config.ts ./apps/server/prisma.config.ts
COPY --from=builder /app/apps/server/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# No chown: the app only reads from /app, and `chown -R` would follow the
# workspace symlinks. Root-owned, world-readable files are what we want here.
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

USER node

EXPOSE 8080

# nest build keeps the source layout, so main.js lands in dist/src/, not dist/.
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/docker-entrypoint.sh"]
CMD ["node", "apps/server/dist/src/main.js"]
