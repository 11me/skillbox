# claude-skillbox

> Specialized workflow layer for Claude Code — cross-session task tracking, semantic code memory, platform engineering patterns.

[![CI](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml/badge.svg)](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml)
[![Version](https://img.shields.io/badge/version-0.18.0-blue?style=flat-square)](https://github.com/11me/claude-skillbox/releases)
[![Python](https://img.shields.io/badge/python-3.12+-blue?style=flat-square&logo=python&logoColor=white)](https://python.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-blueviolet?style=flat-square&logo=anthropic)](https://docs.anthropic.com/en/docs/claude-code)

Skillbox extends Claude Code with specialized workflows that complement Anthropic's official plugins:

- **Workflow Orchestration** — Beads task tracking + Serena code memory + conventional commits
- **Platform Engineering** — Production Kubernetes (Helm, Flux, GitOps)
- **Testing Excellence** — TDD discipline and quality patterns

> **For project scaffolding:** Use Anthropic's `plugin-dev` plugin.

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

### Core Workflows

| Skill | Description |
|-------|-------------|
| [workflow-orchestration](plugins/skillbox/skills/core/workflow-orchestration/) | Unified beads + serena + commit workflow |
| [beads-workflow](plugins/skillbox/skills/core/beads-workflow/) | Cross-session task tracking with beads CLI |
| [serena-navigation](plugins/skillbox/skills/core/serena-navigation/) | Semantic code navigation with Serena MCP |
| [conventional-commit](plugins/skillbox/skills/core/conventional-commit/) | Structured commit messages |
| [context-engineering](plugins/skillbox/skills/core/context-engineering/) | Long-session context management |

### Go Development

| Skill | Description |
|-------|-------------|
| [go-development](plugins/skillbox/skills/go/go-development/) | Production Go patterns: pgx, squirrel, Service Registry, typed errors |

### Platform Engineering

| Skill | Description |
|-------|-------------|
| [helm-chart-developer](plugins/skillbox/skills/k8s/helm-chart-developer/) | Production Helm charts with GitOps (Flux + Kustomize + ESO) |
| [flux-gitops-scaffold](plugins/skillbox/skills/k8s/flux-gitops-scaffold/) | Scaffold Flux GitOps projects with image automation |

### Testing Excellence

| Skill | Description |
|-------|-------------|
| [tdd-enforcer](plugins/skillbox/skills/core/tdd-enforcer/) | TDD Red-Green-Refactor workflow |
| [skill-patterns](plugins/skillbox/skills/core/skill-patterns/) | Do/Verify/Repair, Guardrails patterns |

### Educational Reference

TypeScript patterns (non-core, reference material):

| Skill | Description |
|-------|-------------|
| [ts-conventions](plugins/skillbox/skills/ts/conventions/) | TypeScript code conventions |
| [ts-project-structure](plugins/skillbox/skills/ts/project-structure/) | Monorepo patterns |
| [ts-modern-tooling](plugins/skillbox/skills/ts/modern-tooling/) | pnpm, Biome, Vite |
| [ts-type-patterns](plugins/skillbox/skills/ts/type-patterns/) | Generics, utility types |
| [ts-testing-patterns](plugins/skillbox/skills/ts/testing-patterns/) | Vitest patterns |
| [ts-api-patterns](plugins/skillbox/skills/ts/api-patterns/) | Hono, tRPC, Zod |
| [ts-database-patterns](plugins/skillbox/skills/ts/database-patterns/) | Drizzle ORM |

## Agents

Autonomous agents for specialized tasks. See [agents/_index.md](plugins/skillbox/agents/_index.md).

| Agent | Model | Description |
|-------|-------|-------------|
| [task-tracker](plugins/skillbox/agents/core/task-tracker.md) | haiku | Manage beads task lifecycle |
| [session-checkpoint](plugins/skillbox/agents/core/session-checkpoint.md) | haiku | Save progress to serena memory |
| [code-navigator](plugins/skillbox/agents/core/code-navigator.md) | sonnet | Semantic code exploration |
| [go-project-init](plugins/skillbox/agents/go/project-init.md) | sonnet | Scaffold Go projects with production patterns |
| [test-analyzer](plugins/skillbox/agents/tdd/test-analyzer.md) | sonnet | Analyze test coverage |
| [tdd-coach](plugins/skillbox/agents/tdd/tdd-coach.md) | sonnet | Guide TDD workflow |

## Commands

| Command | Description |
|---------|-------------|
| `/init-workflow` | Initialize workflow tools (beads + serena + CLAUDE.md) |
| `/checkpoint` | Save session progress to serena memory |
| `/commit` | Create conventional commit message |
| `/helm-scaffold` | Scaffold GitOps structure for app |
| `/helm-validate` | Validate Helm chart |
| `/helm-checkpoint` | Save current Helm work state |
| `/flux-init` | Initialize Flux GitOps project |
| `/flux-add-infra` | Add infrastructure component to GitOps |
| `/flux-add-app` | Add application with image automation |
| `/go-add-service` | Generate Go service with factory method |
| `/go-add-repository` | Generate Go repository with interface |
| `/go-add-model` | Generate Go model with mapper |

## Hooks

| Hook | Event | Description |
|------|-------|-------------|
| session_context | SessionStart | Inject project context |
| skill_suggester | SessionStart | Auto-suggest relevant skills |
| git-push-guard | PreToolUse | Confirm before git push |
| pretool-secret-guard | PreToolUse | Block secrets in files |
| validate-flux-manifest | PreToolUse | Validate Flux manifests (API versions, required fields) |
| helmrelease-version-check | PostToolUse | Suggest checking HelmRelease chart versions |

## When to Use Skillbox vs Official Plugins

| Task | Use |
|------|-----|
| Create new plugin | `plugin-dev` (official) |
| Scaffold TypeScript project | `plugin-dev` (official) |
| Scaffold Go project with production patterns | `go-development` (skillbox) |
| Track tasks across sessions | `beads-workflow` (skillbox) |
| Navigate code semantically | `serena-navigation` (skillbox) |
| Create Helm charts | `helm-chart-developer` (skillbox) |
| Scaffold Flux GitOps projects | `flux-gitops-scaffold` (skillbox) |
| TDD workflow enforcement | `tdd-enforcer` (skillbox) |

## Architecture

```
plugins/skillbox/
├── skills/
│   ├── core/                    # Core workflows
│   │   ├── workflow-orchestration/
│   │   ├── beads-workflow/
│   │   ├── serena-navigation/
│   │   ├── conventional-commit/
│   │   ├── context-engineering/
│   │   ├── tdd-enforcer/
│   │   └── skill-patterns/
│   ├── go/                      # Go development
│   │   └── go-development/      # Production patterns
│   ├── ts/                      # TypeScript (educational)
│   └── k8s/                     # Platform engineering
│       ├── helm-chart-developer/
│       └── flux-gitops-scaffold/
├── agents/                      # Autonomous agents
├── commands/                    # Slash commands
├── hooks/                       # Event hooks
└── scripts/                     # Python implementations
```

## Development

```bash
# Clone
git clone https://github.com/11me/claude-skillbox.git
cd claude-skillbox

# Install pre-commit
uv tool install pre-commit
pre-commit install

# Test locally
claude --plugin-dir ./plugins/skillbox
```

## License

[MIT](LICENSE)
