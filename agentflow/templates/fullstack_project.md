# Fullstack Project KB Template

## INDEX.md (Fullstack)

```markdown
# {project_name}

## Overview
{description}

## Tech Stack
### Frontend
- Framework: {Next.js|Nuxt|SvelteKit|Remix}
- Language: {TypeScript}
- Styling: {Tailwind|CSS Modules}

### Backend
- Runtime: {Node.js|Python|Go}
- Framework: {Express|FastAPI|Gin}
- Database: {PostgreSQL|MongoDB}
- ORM: {Prisma|SQLAlchemy|Drizzle}

### Infrastructure
- Hosting: {Vercel|AWS|GCP|Railway}
- CI/CD: {GitHub Actions|GitLab CI}
- Containerization: {Docker|Docker Compose}

## Architecture
- Monorepo: {yes|no}
- API Pattern: {REST|tRPC|GraphQL}
- Auth: {NextAuth|Clerk|custom JWT}

## Entry Points
- Frontend: `{apps/web/src/app/page.tsx}`
- Backend: `{apps/api/src/main.ts}`
- Shared: `{packages/shared/}`

## Key Directories
- `apps/web/` — Frontend application
- `apps/api/` — Backend API
- `packages/shared/` — Shared types / utilities
- `packages/ui/` — Shared UI components
- `prisma/` — Database schema & migrations
```

## context.md (Fullstack)

```markdown
# Project Context

## Monorepo Structure
- Tool: {Turborepo|Nx|pnpm workspaces}
- Package manager: {pnpm|npm|yarn}

## Build & Run
- Dev (all): `{pnpm dev}`
- Dev (frontend): `{pnpm --filter web dev}`
- Dev (backend): `{pnpm --filter api dev}`
- Build: `{pnpm build}`
- Test: `{pnpm test}`

## Database
- Dev: `{docker compose up db}`
- Migrate: `{pnpm prisma migrate dev}`
- Seed: `{pnpm prisma db seed}`
- Studio: `{pnpm prisma studio}`

## Environment Variables
### Frontend (.env.local)
- `NEXT_PUBLIC_API_URL` — Backend API URL

### Backend (.env)
- `DATABASE_URL` — Database connection
- `JWT_SECRET` — Auth secret

## Cross-Cutting Concerns
- Shared types in `packages/shared/`
- API client generated from {OpenAPI|tRPC router}
```
