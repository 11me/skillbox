---
name: ts-project-structure
description: TypeScript project organization with monorepo patterns. Use when setting up TypeScript projects, configuring tsconfig, or organizing monorepos with Turborepo and pnpm.
globs: ["**/tsconfig*.json", "**/package.json", "**/turbo.json", "**/pnpm-workspace.yaml"]
allowed-tools: Read, Grep, Glob
---

# TypeScript Project Structure

Modern TypeScript project organization with monorepo patterns (2025).

## Single Package Structure

```
mypackage/
├── src/
│   ├── index.ts              # Main entry point (barrel export)
│   ├── types.ts              # Shared type definitions
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
│           └── ...
├── tests/
│   ├── setup.ts
│   └── features/
│       └── users.test.ts
├── dist/                     # Built output (gitignored)
├── package.json
├── tsconfig.json
├── biome.json
└── vitest.config.ts
```

## Monorepo Structure (Turborepo + pnpm)

```
monorepo/
├── apps/
│   ├── web/                  # Next.js/SvelteKit app
│   │   ├── src/
│   │   ├── package.json
│   │   └── tsconfig.json
│   └── api/                  # Hono/tRPC backend
│       ├── src/
│       ├── package.json
│       └── tsconfig.json
├── packages/
│   ├── shared/               # Shared types & utilities
│   │   ├── src/
│   │   ├── package.json
│   │   └── tsconfig.json
│   ├── ui/                   # Shared UI components
│   │   └── ...
│   └── config/               # Shared configs
│       ├── tsconfig/
│       │   ├── base.json
│       │   ├── node.json
│       │   └── react.json
│       └── biome/
│           └── biome.json
├── turbo.json
├── pnpm-workspace.yaml
├── package.json
└── biome.json
```

## tsconfig.json Patterns

### Base Configuration

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

### React/Browser Configuration

```json
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "noEmit": true
  }
}
```

## Barrel Exports

```typescript
// src/features/users/index.ts
export { UserService } from './service';
export { UserRepository } from './repository';
export type { User, CreateUserInput } from './types';
```

### Anti-pattern: Circular Dependencies

```typescript
// BAD: Creates circular import
import { PostService } from '../posts'; // posts imports users!

// GOOD: Extract shared types
import type { User } from '@/shared/types';
```

## Path Aliases

### tsconfig.json

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

### Usage

```typescript
// Instead of relative paths
import { UserService } from '../../../features/users';

// Use aliases
import { UserService } from '@/features/users';
```

## Monorepo Setup

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
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "test": {
      "dependsOn": ["build"]
    },
    "lint": {},
    "typecheck": {
      "dependsOn": ["^build"]
    }
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

## Best Practices

### DO:
- Use `src/` for source, `dist/` for output
- Group by feature, not by type
- Use barrel exports (`index.ts`) for clean imports
- Keep shared types in `src/types.ts` or `src/shared/`
- Use path aliases for deep imports
- Use `workspace:*` for monorepo dependencies

### DON'T:
- Mix source and output in same directory
- Create circular dependencies between features
- Export everything from root (breaks tree-shaking)
- Use relative paths like `../../../`
- Commit `node_modules/` or `dist/`

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
