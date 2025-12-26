---
description: Initialize a new TypeScript project with modern structure
allowed-tools: Bash(pnpm:*), Bash(npm:*), Bash(mkdir:*), Bash(ls:*), Write, Edit, Read
argument-hint: <project-name> [library|api|fullstack|monorepo]
---

# Initialize TypeScript Project

## Context

- Current directory: !`pwd`
- Node version: !`node --version 2>/dev/null || echo "Node not installed"`
- pnpm version: !`pnpm --version 2>/dev/null || echo "pnpm not installed"`

## Task

Create a new TypeScript project:
- **Project name:** $1
- **Project type:** $2 (default: library)

### Project Types
- `library` — Publishable npm package with tsup, ESM + CJS
- `api` — Hono backend with Drizzle ORM, Zod validation
- `fullstack` — Vite frontend + Hono backend in monorepo
- `monorepo` — Turborepo + pnpm workspaces setup

### Structure by Type

**library:**
```
src/index.ts
tests/index.test.ts
```

**api:**
```
src/index.ts
src/routes/health.ts
src/db/schema.ts
tests/routes/health.test.ts
```

**fullstack:**
```
apps/web/src/main.tsx
apps/api/src/index.ts
packages/shared/src/index.ts
```

**monorepo:**
```
apps/
packages/
turbo.json
pnpm-workspace.yaml
```

### Config Files

Use templates from `${CLAUDE_PLUGIN_ROOT}/templates/ts/`:
- `tsconfig.json` — Strict TypeScript
- `biome.json` — Linter/formatter
- `vitest.config.ts` — Tests
- `.gitignore`

### Verify

```bash
pnpm install
pnpm typecheck
pnpm lint
pnpm test:run
```

Use `ts/project-init` agent if complex scaffolding needed.
