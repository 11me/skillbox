# Skillbox Skills Registry

Quick lookup for available skills, triggers, and outputs.

## Categories

### Core (Workflow & Git)

| Skill | Description | Triggers | Outputs |
|-------|-------------|----------|---------|
| [conventional-commit](core/conventional-commit/) | Git commits with Conventional Commits spec | "commit", "git message", "коммит" | Commit message |
| [beads-workflow](core/beads-workflow/) | Task tracking with beads CLI | "task", "todo", "bd", "issue" | Task state |
| [skill-creator](core/skill-creator/) | Create new Claude Code skills | "create skill", "scaffold skill" | SKILL.md |
| [serena-navigation](core/serena-navigation/) | Semantic code navigation | "serena", "find symbol", "symbol search" | Code insights, memories |
| [context-engineering](core/context-engineering/) | AI context window management | "context overflow", "token limits" | Optimized context |
| [tdd-enforcer](core/tdd-enforcer/) | TDD workflow patterns | "tdd", "red-green-refactor", "test first" | Test patterns |

### TypeScript

| Skill | Description | Triggers | Outputs |
|-------|-------------|----------|---------|
| [ts-conventions](ts/conventions/) | TypeScript code conventions | auto-activate on `*.ts`, `*.tsx` | Code review context |
| [ts-project-structure](ts/project-structure/) | Monorepo patterns | "tsconfig", "monorepo", "turborepo" | Project structure |
| [ts-modern-tooling](ts/modern-tooling/) | pnpm, Biome, Vite, tsup | "biome", "vite", "pnpm" | Build config |
| [ts-type-patterns](ts/type-patterns/) | Generics, utility types | "generics", "utility types" | Type patterns |
| [ts-testing-patterns](ts/testing-patterns/) | Vitest testing | "vitest", "mocking" | Test setup |
| [ts-api-patterns](ts/api-patterns/) | Hono, tRPC, Zod | "hono", "trpc", "zod" | API code |
| [ts-database-patterns](ts/database-patterns/) | Drizzle ORM | "drizzle", "orm", "schema" | Database code |

### Kubernetes (GitOps)

| Skill | Description | Triggers | Outputs |
|-------|-------------|----------|---------|
| [helm-chart-developer](k8s/helm-chart-developer/) | Helm charts with Flux + ESO | "helm", "chart", "gitops", "helmrelease" | Validated charts |

## Quick Lookup

### By Task

| Task | Skill |
|------|-------|
| Write commit message | conventional-commit |
| Track tasks/issues | beads-workflow |
| Navigate codebase semantically | serena-navigation |
| Create Helm chart | helm-chart-developer |
| Create new skill | skill-creator |
| Manage AI context | context-engineering |
| TDD workflow | tdd-enforcer |
| TypeScript project setup | ts-modern-tooling, ts-project-structure |
| TypeScript API | ts-api-patterns |
| Database with Drizzle | ts-database-patterns |

### By File Pattern

| Pattern | Skill |
|---------|-------|
| `Chart.yaml`, `values.yaml` | helm-chart-developer |
| `.beads/` | beads-workflow |
| `.serena/` | serena-navigation |
| `SKILL.md` | skill-creator |
| `*.ts`, `*.tsx` | ts-conventions, ts-type-patterns |
| `tsconfig.json` | ts-project-structure |
| `biome.json`, `vite.config.*` | ts-modern-tooling |
| `*.test.ts`, `*.spec.ts` | ts-testing-patterns, tdd-enforcer |
| `drizzle.config.*` | ts-database-patterns |

### By Command

| Command | Skill/Action |
|---------|--------------|
| `/commit` | conventional-commit |
| `/skill-scaffold` | skill-creator |
| `/helm-scaffold` | helm-chart-developer |
| `/helm-validate` | helm-chart-developer |

## Skill Interactions

```
beads-workflow ←→ conventional-commit
     │              (task refs in commits)
     ↓
serena-navigation
     │ (memories persist discoveries)
     ↓
helm-chart-developer
     (validation before complete)

tdd-enforcer ←→ ts-testing-patterns
     │           (Vitest patterns)
     ↓
context-engineering
     (optimize long sessions)
```

## Adding New Skills

See [skill-creator](core/skill-creator/) for templates and best practices.
