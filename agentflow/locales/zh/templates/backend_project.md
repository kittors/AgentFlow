# Backend Project KB Template

## INDEX.md (Backend)

```markdown
# {project_name}

## Overview
{description}

## Tech Stack
- Language: {Python|Node.js|Go|Rust|Java}
- Framework: {FastAPI|Express|Gin|Actix|Spring Boot}
- Database: {PostgreSQL|MySQL|MongoDB|Redis|SQLite}
- ORM: {SQLAlchemy|Prisma|GORM|Diesel|Hibernate}
- Auth: {JWT|OAuth2|Session|API Key}

## Architecture
- Pattern: {MVC|Clean Architecture|Hexagonal|Layered}
- API Style: {REST|GraphQL|gRPC}
- Deployment: {Docker|Kubernetes|Serverless|VM}

## Entry Points
- `{src/main.py}` — Server startup
- `{src/routes/}` — Route handlers
- `{src/models/}` — Data models

## Key Directories
- `src/routes/` — API routes / controllers
- `src/models/` — Data models / schemas
- `src/services/` — Business logic
- `src/middleware/` — Request middleware
- `src/utils/` — Utility functions
- `migrations/` — Database migrations
```

## context.md (Backend)

```markdown
# Project Context

## Dependencies
{dependency_list}

## Database
- Type: {db_type}
- Connection: `{DATABASE_URL}`
- Migrations: `{migration_command}`

## API Endpoints
| Method | Path | Description |
|--------|------|-------------|
| {GET} | {/api/v1/resource} | {description} |

## Build & Run
- Dev: `{python -m uvicorn main:app --reload}`
- Test: `{pytest}`
- Migrate: `{alembic upgrade head}`
- Lint: `{ruff check .}`

## Environment Variables
- `DATABASE_URL` — Database connection string
- `SECRET_KEY` — Application secret key
- `PORT` — Server port (default: {8000})

## Monitoring
- Health check: `{/health}`
- Metrics: `{/metrics}`
```
