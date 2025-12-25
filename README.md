# claude-skillbox

> Extensible skills marketplace for Claude Code — automate Helm/GitOps workflows, semantic code navigation, conventional commits, and more.

[![CI](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml/badge.svg)](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml)
[![Version](https://img.shields.io/badge/version-0.11.0-blue?style=flat-square)](https://github.com/11me/claude-skillbox/releases)
[![Python](https://img.shields.io/badge/python-3.12+-blue?style=flat-square&logo=python&logoColor=white)](https://python.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-blueviolet?style=flat-square&logo=anthropic)](https://docs.anthropic.com/en/docs/claude-code)

A Claude Code plugin that extends AI-assisted development with specialized skills for Kubernetes/Helm automation, GitOps with Flux CD, semantic code navigation via Serena MCP, and structured commit workflows.

## Features

- **Smart Project Detection** — automatically suggests relevant skills based on your project type (Helm, Go, Python, Node.js, Rust)
- **Helm/GitOps Automation** — production-ready charts with Flux CD, Kustomize overlays, and External Secrets Operator integration
- **Semantic Code Navigation** — Serena MCP integration for intelligent symbol search and code exploration
- **Conventional Commits** — structured commit messages following the Conventional Commits specification
- **Safety Hooks** — prevent accidental secret exposure, require confirmation for git push, enforce validation before completion

## Quick Start

```bash
# Add the marketplace
/plugin marketplace add 11me/claude-skillbox

# Install the plugin
/plugin install skillbox@11me-skillbox
```

Or test locally:

```bash
claude --plugin-dir ./plugins/skillbox
```

## Skills

### Core Skills

| Skill | Description |
|-------|-------------|
| [conventional-commit](plugins/skillbox/skills/core/conventional-commit/) | Generate git commits following Conventional Commits spec |
| [skill-creator](plugins/skillbox/skills/core/skill-creator/) | Create new Claude Code skills with proper structure |
| [beads-workflow](plugins/skillbox/skills/core/beads-workflow/) | Task tracking with beads CLI |
| [serena-navigation](plugins/skillbox/skills/core/serena-navigation/) | Semantic code navigation with Serena MCP |
| [context-engineering](plugins/skillbox/skills/core/context-engineering/) | AI context window management and optimization |
| [tdd-enforcer](plugins/skillbox/skills/core/tdd-enforcer/) | TDD Red-Green-Refactor workflow for Go, TS, Python, Rust |

### TypeScript Skills

| Skill | Description |
|-------|-------------|
| [ts-conventions](plugins/skillbox/skills/ts/conventions/) | TypeScript code conventions and best practices |
| [ts-project-structure](plugins/skillbox/skills/ts/project-structure/) | Monorepo patterns with Turborepo + pnpm |
| [ts-modern-tooling](plugins/skillbox/skills/ts/modern-tooling/) | pnpm, Biome, Vite, tsup (2025 stack) |
| [ts-type-patterns](plugins/skillbox/skills/ts/type-patterns/) | Advanced generics, utility types, type guards |
| [ts-testing-patterns](plugins/skillbox/skills/ts/testing-patterns/) | Vitest testing patterns and mocking |
| [ts-api-patterns](plugins/skillbox/skills/ts/api-patterns/) | Type-safe APIs with Hono, tRPC, Zod |
| [ts-database-patterns](plugins/skillbox/skills/ts/database-patterns/) | Drizzle ORM schemas, queries, migrations |

### Kubernetes Skills

| Skill | Description |
|-------|-------------|
| [helm-chart-developer](plugins/skillbox/skills/k8s/helm-chart-developer/) | Build production Helm charts with GitOps (Flux + Kustomize + ESO) |

## Agents

Autonomous agents for specialized multi-step tasks. See [agents/_index.md](plugins/skillbox/agents/_index.md) for full registry.

### Core Agents

| Agent | Model | Description |
|-------|-------|-------------|
| [task-tracker](plugins/skillbox/agents/core/task-tracker.md) | haiku | Manage beads task lifecycle |
| [session-checkpoint](plugins/skillbox/agents/core/session-checkpoint.md) | haiku | Save progress to serena memory |
| [code-navigator](plugins/skillbox/agents/core/code-navigator.md) | sonnet | Semantic code exploration |

### TDD Agents

| Agent | Model | Description |
|-------|-------|-------------|
| [test-analyzer](plugins/skillbox/agents/tdd/test-analyzer.md) | sonnet | Analyze coverage, find gaps |
| [tdd-coach](plugins/skillbox/agents/tdd/tdd-coach.md) | sonnet | Guide Red-Green-Refactor workflow |

### Language Agents

| Agent | Model | Description |
|-------|-------|-------------|
| [ts/test-generator](plugins/skillbox/agents/ts/test-generator.md) | sonnet | Generate Vitest tests |
| [ts/project-init](plugins/skillbox/agents/ts/project-init.md) | sonnet | Scaffold TypeScript project |
| [go/test-generator](plugins/skillbox/agents/go/test-generator.md) | sonnet | Generate Go table-driven tests |
| [go/project-init](plugins/skillbox/agents/go/project-init.md) | sonnet | Scaffold Go project |
| [python/test-writer](plugins/skillbox/agents/python/test-writer.md) | opus | Generate pytest tests |
| [rust/test-generator](plugins/skillbox/agents/rust/test-generator.md) | sonnet | Generate Rust tests |

## Commands

| Command | Description |
|---------|-------------|
| `/commit` | Create git commit with Conventional Commits message |
| `/skill-scaffold` | Scaffold a new skill directory with SKILL.md template |
| `/helm-scaffold` | Scaffold complete GitOps structure for a new app |
| `/helm-validate` | Validate Helm chart (lint, template, dry-run) |
| `/helm-checkpoint` | Create checkpoint summary of current Helm work |

## Hooks

All hooks are written in Python with shared utilities. See [HOOKS.md](plugins/skillbox/HOOKS.md) for the development guide.

| Hook | Event | Description |
|------|-------|-------------|
| session_context | SessionStart | Inject date, project context, beads status |
| flow_check | SessionStart | Check workflow compliance (CLAUDE.md, pre-commit) |
| skill_suggester | SessionStart | Auto-detect project type and suggest skills |
| git-push-guard | PreToolUse | Require confirmation before git push |
| pretool-secret-guard | PreToolUse | Block secrets in values.yaml |

## Architecture

```
plugins/skillbox/
├── skills/
│   ├── core/                    # Core workflow skills
│   │   ├── conventional-commit/
│   │   ├── skill-creator/
│   │   ├── beads-workflow/
│   │   ├── serena-navigation/
│   │   ├── context-engineering/
│   │   └── tdd-enforcer/
│   ├── ts/                      # TypeScript skills
│   │   ├── conventions/
│   │   ├── project-structure/
│   │   ├── modern-tooling/
│   │   ├── type-patterns/
│   │   ├── testing-patterns/
│   │   ├── api-patterns/
│   │   └── database-patterns/
│   └── k8s/                     # Kubernetes skills
│       └── helm-chart-developer/
├── agents/                      # Autonomous agents
│   ├── core/                    # Workflow agents
│   │   ├── task-tracker.md
│   │   ├── session-checkpoint.md
│   │   └── code-navigator.md
│   ├── tdd/                     # TDD agents
│   │   ├── test-analyzer.md
│   │   └── tdd-coach.md
│   ├── ts/                      # TypeScript agents
│   │   ├── test-generator.md
│   │   └── project-init.md
│   ├── go/                      # Go agents
│   │   ├── test-generator.md
│   │   └── project-init.md
│   ├── python/                  # Python agents
│   │   └── test-writer.md
│   └── rust/                    # Rust agents
│       └── test-generator.md
├── commands/                    # Slash commands (/commit, /helm-*)
├── hooks/                       # Event hooks configuration
│   └── hooks.json
├── scripts/hooks/               # Python hook implementations
│   ├── lib/                     # Shared utilities
│   └── *.py
└── HOOKS.md                     # Hook development guide
```

## Development

### Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/) — fast Python package manager

### Setup

```bash
# Clone the repository
git clone https://github.com/11me/claude-skillbox.git
cd claude-skillbox

# Install pre-commit hooks
uv tool install pre-commit
pre-commit install

# Run all checks
pre-commit run --all-files
```

### Testing Locally

```bash
claude --plugin-dir ./plugins/skillbox
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Adding new skills
- Writing hooks
- Code style (Python: ruff, Commits: Conventional)
- Pull request process

## Versioning

This project follows [Semantic Versioning](https://semver.org/). See [CLAUDE.md](CLAUDE.md) for versioning guidelines.

## License

[MIT](LICENSE)
