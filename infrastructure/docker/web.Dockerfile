FROM node:20-alpine AS base
WORKDIR /app

COPY package.json package-lock.json* ./
COPY apps/web/package.json ./apps/web/
COPY packages/types/package.json ./packages/types/
COPY packages/config/package.json ./packages/config/
COPY packages/utils/package.json ./packages/utils/

RUN npm install --workspace=@docksight/web --include-workspace-root

COPY apps/web ./apps/web
COPY packages ./packages

WORKDIR /app/apps/web

EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
