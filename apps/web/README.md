# DockSight Web

React dashboard for DockSight — open-source Docker management and observability.

## Stack

- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- React Query
- Zustand
- Lucide React

## Install dependencies

From the monorepo root:

```bash
npm install
```

Or from this package:

```bash
cd apps/web
npm install
```

## Environment

```bash
cp .env.example .env
```

```env
VITE_API_URL=
```

## Run development server

```bash
npm run dev
```

App: `http://localhost:5173`

## Project structure

```
src/
├── app/           # App shell and providers
├── components/    # Reusable UI (shadcn/ui)
├── features/      # Future business features
├── hooks/         # Reusable React hooks
├── services/      # Future API communication
├── stores/        # Future Zustand stores
├── types/         # Shared frontend types
├── utils/         # Helper functions
└── lib/           # Shared utilities (e.g. cn)
```
