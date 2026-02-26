# Frontend Project KB Template

## INDEX.md (Frontend)

```markdown
# {project_name}

## Overview
{description}

## Tech Stack
- Framework: {React|Vue|Angular|Svelte|Next.js}
- Language: {TypeScript|JavaScript}
- Styling: {Tailwind|CSS Modules|Styled Components|Sass}
- State Management: {Redux|Zustand|Pinia|Vuex|Jotai}
- Package Manager: {npm|pnpm|yarn|bun}
- Build Tool: {Vite|Webpack|Turbopack|esbuild}

## Architecture
- Pattern: {SPA|SSR|SSG|ISR}
- Routing: {file-based|config-based}
- API Layer: {REST|GraphQL|tRPC}

## Entry Points
- `{src/main.tsx}` — App bootstrap
- `{src/App.tsx}` — Root component

## Key Directories
- `src/components/` — Reusable UI components
- `src/pages/` — Route pages
- `src/hooks/` — Custom hooks
- `src/utils/` — Utility functions
- `src/styles/` — Global styles
- `src/api/` — API client layer
- `src/store/` — State management
```

## context.md (Frontend)

```markdown
# Project Context

## Dependencies
- Runtime: {dependency_list}
- Dev: {dev_dependency_list}

## Build & Run
- Dev: `{npm run dev}`
- Build: `{npm run build}`
- Test: `{npm run test}`
- Lint: `{npm run lint}`

## Environment Variables
- `{VITE_API_URL}` — API endpoint
- `{VITE_APP_TITLE}` — App title

## Browser Support
- {browser_targets}

## Design System
- Components: {component_library}
- Icons: {icon_set}
- Fonts: {font_stack}
```
