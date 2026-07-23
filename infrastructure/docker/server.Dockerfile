FROM node:20-alpine AS base
WORKDIR /app

COPY package.json package-lock.json* ./
COPY apps/server/package.json ./apps/server/
COPY packages/types/package.json ./packages/types/
COPY packages/config/package.json ./packages/config/
COPY packages/utils/package.json ./packages/utils/

RUN npm install --workspace=@docksight/server --include-workspace-root

COPY apps/server ./apps/server
COPY packages ./packages

WORKDIR /app/apps/server
RUN npx prisma generate

EXPOSE 3000
CMD ["npm", "run", "start:dev"]
