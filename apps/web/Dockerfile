# syntax=docker/dockerfile:1
#
# DockSight web console (static build served by nginx).
# Build from the REPOSITORY ROOT, not from this directory:
#   docker build -f apps/web/Dockerfile -t docksight/web .
#
# Same monorepo rule as the server image: every workspace manifest has to be
# present before `npm ci`, otherwise npm installs nothing for the workspaces
# and `tsc` / `vite` are missing at build time.

FROM node:22-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json ./
COPY apps/server/package.json apps/server/
COPY apps/web/package.json apps/web/
COPY packages/config/package.json packages/config/
COPY packages/protocol/package.json packages/protocol/
COPY packages/types/package.json packages/types/
COPY packages/utils/package.json packages/utils/

# packages/protocol runs tsc in a `prepare` script during install, so its
# sources must exist even though the web app does not import it.
COPY packages/protocol/tsconfig.json packages/protocol/
COPY packages/protocol/src packages/protocol/src

# The web app has no native dependencies, so install scripts can be skipped.
RUN npm ci --ignore-scripts

COPY . .

# Vite inlines env vars at build time, so the API URL is baked into the bundle
# and cannot be changed by the container's environment later.
#
# The default is relative on purpose: in the compose deployment the browser
# talks to the front nginx, which proxies /api to the server. An absolute URL
# would hard-code a hostname that is only correct on one machine.
ARG VITE_API_URL=/api
ENV VITE_API_URL=${VITE_API_URL}

RUN npm run build --workspace=@docksight/web

FROM nginx:alpine AS runner

# SPA fallback + asset caching. Without this, reloading /hosts or /login
# returns 404 because no such file exists on disk.
COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf

# In a workspace build the output lands in apps/web/dist, not /app/dist.
COPY --from=builder /app/apps/web/dist /usr/share/nginx/html

EXPOSE 80

# nginx:alpine already sets this CMD; kept explicit for readability.
CMD ["nginx", "-g", "daemon off;"]
