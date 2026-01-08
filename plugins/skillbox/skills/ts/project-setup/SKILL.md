---
name: ts-project-setup
description: This skill should be used when the user asks to "scaffold TypeScript project", "configure pnpm", "set up Biome", "configure Vite", or mentions "tsup", "monorepo", "TypeScript project setup".
allowed-tools: [Read, Grep, Glob, Write, Edit, Bash]
---

# TypeScript Project Setup

Modern TypeScript project structure and tooling (2025).

## Quick Start

```bash
# Create project
mkdir my-project && cd my-project
pnpm init

# Install dependencies
pnpm add -D typescript @types/node tsup vitest @biomejs/biome

# Initialize configs
pnpm tsc --init
pnpm biome init

# Create structure
mkdir src tests
echo 'export const hello = "world";' > src/index.ts
```

## Package Managers

### pnpm (Recommended Default)

```bash
pnpm init
pnpm add zod drizzle-orm
pnpm add -D typescript @types/node vitest
```

**Why pnpm:**
- 70% less disk space (content-addressable storage)
- Strict dependency resolution (no phantom deps)
- Built-in monorepo support
- 2-3x faster than npm

### Bun (For Speed)

```bash
bun init
bun add zod drizzle-orm
bun run src/index.ts  # Run TS directly
```

## Project Structure

### Single Package

```
mypackage/
├── src/
│   ├── index.ts              # Main entry (barrel export)
│   ├── types.ts              # Shared types
│   ├── utils/
│   │   ├── index.ts          # Barrel export
│   │   └── helpers.ts
│   └── features/
│       ├── users/
│       │   ├── index.ts      # Feature barrel
│       │   ├── types.ts
│       │   ├── service.ts
│       │   └── repository.ts
│       └── posts/
├── tests/
│   ├── setup.ts
│   └── features/
├── dist/                     # Built output (gitignored)
├── package.json
├── tsconfig.json
├── biome.json
└── vitest.config.ts
```

### Monorepo (Turborepo + pnpm)

```
monorepo/
├── apps/
│   ├── web/                  # Next.js/SvelteKit app
│   └── api/                  # Hono/tRPC backend
├── packages/
│   ├── shared/               # Shared types & utilities
│   ├── ui/                   # Shared UI components
│   └── config/               # Shared configs
│       ├── tsconfig/
│       │   ├── base.json
│       │   ├── node.json
│       │   └── react.json
│       └── biome/
├── turbo.json
├── pnpm-workspace.yaml
├── package.json
└── biome.json
```

## TypeScript Configuration

### tsconfig.json (Strict Mode)

```json
{
  "$schema": "https://json.schemastore.org/tsconfig",
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022"],
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "resolveJsonModule": true,
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "noPropertyAccessFromIndexSignature": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "dist",
    "esModuleInterop": true,
    "isolatedModules": true,
    "verbatimModuleSyntax": true,
    "skipLibCheck": true,
    "incremental": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

### Path Aliases

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@/features/*": ["src/features/*"]
    }
  }
}
```

## Build Tools

### tsup (Libraries)

```typescript
// tsup.config.ts
import { defineConfig } from 'tsup';

export default defineConfig({
  entry: ['src/index.ts'],
  format: ['esm', 'cjs'],      // Dual ESM/CJS
  dts: true,                    // Generate .d.ts
  splitting: true,
  sourcemap: true,
  clean: true,
  minify: true,
  target: 'es2022',
  outDir: 'dist',
});
```

### Vite (Applications)

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  build: { target: 'es2022', sourcemap: true },
  server: { port: 3000 },
});
```

| Tool | Use Case |
|------|----------|
| **tsup** | Libraries, CLI tools, simple packages |
| **Vite** | Web apps, dev server with HMR |
| **esbuild** | Direct bundling, maximum speed |

## Linting: Biome

### Why Biome (NOT ESLint + Prettier)

- **10-25x faster** than ESLint
- **Single tool** for lint + format
- **Zero config** option
- **No plugin hell**

### biome.json

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.0/schema.json",
  "organizeImports": { "enabled": true },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "correctness": {
        "noUnusedImports": "error",
        "noUnusedVariables": "error"
      },
      "style": {
        "noNonNullAssertion": "warn",
        "useConst": "error"
      },
      "suspicious": { "noExplicitAny": "warn" }
    }
  },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "semicolons": "always",
      "trailingCommas": "all"
    }
  }
}
```

### Usage

```bash
pnpm biome check .          # Check (lint + format)
pnpm biome check --write .  # Fix automatically
pnpm biome ci .             # CI mode (no writes)
```

## Monorepo Configuration

### pnpm-workspace.yaml

```yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

### turbo.json

```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "build": { "dependsOn": ["^build"], "outputs": ["dist/**"] },
    "dev": { "cache": false, "persistent": true },
    "test": { "dependsOn": ["build"] },
    "lint": {},
    "typecheck": { "dependsOn": ["^build"] }
  }
}
```

### Package References

```json
{
  "name": "@monorepo/web",
  "dependencies": {
    "@monorepo/shared": "workspace:*",
    "@monorepo/ui": "workspace:*"
  }
}
```

## package.json Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "test": "vitest",
    "test:coverage": "vitest run --coverage",
    "lint": "biome check .",
    "lint:fix": "biome check --write .",
    "typecheck": "tsc --noEmit",
    "ci": "biome ci . && pnpm typecheck && pnpm test:coverage"
  }
}
```

## Best Practices

### DO:
- Use `src/` for source, `dist/` for output
- Group by feature, not by type
- Use barrel exports (`index.ts`) for clean imports
- Use path aliases for deep imports
- Use `workspace:*` for monorepo dependencies

### DON'T:
- Mix source and output in same directory
- Create circular dependencies between features
- Export everything from root (breaks tree-shaking)
- Use relative paths like `../../../`
- Commit `node_modules/` or `dist/`

## Anti-patterns

| Wrong | Right |
|-------|-------|
| `npm install` | `pnpm add` |
| ESLint + Prettier + 50 plugins | Biome |
| `tsc` for bundling | tsup or Vite |
| `strict: false` | `strict: true` |
| CommonJS (`require`) | ESM (`import`) |

## Related Skills

- **ts-conventions** — Code conventions and API patterns
- **ts-database-patterns** — Drizzle ORM patterns
- **ts-testing-patterns** — Vitest testing

## Version History

- 1.0.0 — Initial release (merged from project-structure + modern-tooling)
