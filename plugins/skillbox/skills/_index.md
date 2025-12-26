# Skillbox Skills Registry

Specialized workflow layer extending Claude Code with domain expertise.

> **For project scaffolding:** Use Anthropic's `plugin-dev` plugin.
> Skillbox focuses on workflow orchestration, platform engineering, and testing excellence.

## Core Workflows

Task tracking, code memory, and commit traceability.

| Skill | Description | Triggers |
|-------|-------------|----------|
| [workflow-orchestration](core/workflow-orchestration/) | Unified beads + serena + commit workflow | "start feature", "track task" |
| [beads-workflow](core/beads-workflow/) | Cross-session task tracking | "task", "bd", "issue" |
| [serena-navigation](core/serena-navigation/) | Semantic code memory | "find symbol", "serena" |
| [conventional-commit](core/conventional-commit/) | Structured commit messages | "commit", "git message" |
| [context-engineering](core/context-engineering/) | Long-session context management | "context overflow" |

## Platform Engineering

Production Kubernetes and GitOps patterns.

| Skill | Description | Triggers |
|-------|-------------|----------|
| [helm-chart-developer](k8s/helm-chart-developer/) | Helm charts with Flux + ESO | "helm", "chart", "gitops" |

## Testing Excellence

TDD discipline and quality patterns.

| Skill | Description | Triggers |
|-------|-------------|----------|
| [tdd-enforcer](core/tdd-enforcer/) | Red-Green-Refactor workflow | "tdd", "test first" |
| [skill-patterns](core/skill-patterns/) | Do/Verify/Repair, Guardrails | "improve skill" |

## Educational Reference

TypeScript patterns and best practices (non-core, reference material).

| Skill | Description | Triggers |
|-------|-------------|----------|
| [ts-conventions](ts/conventions/) | Code conventions | `*.ts`, `*.tsx` |
| [ts-project-structure](ts/project-structure/) | Monorepo patterns | "tsconfig", "monorepo" |
| [ts-modern-tooling](ts/modern-tooling/) | pnpm, Biome, Vite | "biome", "vite" |
| [ts-type-patterns](ts/type-patterns/) | Generics, utility types | "generics" |
| [ts-testing-patterns](ts/testing-patterns/) | Vitest patterns | "vitest" |
| [ts-api-patterns](ts/api-patterns/) | Hono, tRPC, Zod | "hono", "trpc" |
| [ts-database-patterns](ts/database-patterns/) | Drizzle ORM | "drizzle" |

## Quick Lookup

### By Task

| Task | Skill |
|------|-------|
| Start new feature | workflow-orchestration |
| Track work across sessions | beads-workflow |
| Navigate code semantically | serena-navigation |
| Write commit message | conventional-commit |
| Create Helm chart | helm-chart-developer |
| TDD workflow | tdd-enforcer |

### By Command

| Command | Skill/Action |
|---------|--------------|
| `/init-workflow` | Initialize beads + serena + CLAUDE.md |
| `/commit` | conventional-commit |
| `/helm-scaffold` | helm-chart-developer |
| `/helm-validate` | helm-chart-developer |

## Workflow Integration

```
                    ┌─────────────────────────────────┐
                    │   workflow-orchestration        │
                    │   (unified pattern)             │
                    └─────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│ beads-workflow│◄──►│serena-navigation│◄──►│conventional-  │
│ (tasks)       │    │ (code memory) │    │ commit        │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │
        └──────────┬──────────┘
                   ▼
           context-engineering
           (long sessions)
```

## Adding New Skills

Use Anthropic's `plugin-dev` for skill creation mechanics.
For quality patterns (Do/Verify/Repair, Guardrails), see [skill-patterns](core/skill-patterns/).
