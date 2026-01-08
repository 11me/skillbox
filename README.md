# claude-skillbox

> Specialized workflow layer for Claude Code — cross-session task tracking, semantic code memory, platform engineering patterns.

[![CI](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml/badge.svg)](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml)
[![Version](https://img.shields.io/badge/version-0.67.0-blue?style=flat-square)](https://github.com/11me/claude-skillbox/releases)
[![Python](https://img.shields.io/badge/python-3.12+-blue?style=flat-square&logo=python&logoColor=white)](https://python.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-blueviolet?style=flat-square&logo=anthropic)](https://docs.anthropic.com/en/docs/claude-code)

Skillbox extends Claude Code with specialized workflows that complement Anthropic's official plugins:

- **Workflow Orchestration** — Beads task tracking + Serena code memory + conventional commits
- **Platform Engineering** — Production Kubernetes (Helm, Flux, GitOps)
- **Testing Excellence** — TDD discipline and quality patterns
- **Desktop Notifications** — notify-send alerts when Claude needs attention

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
| [production-flow](plugins/skillbox/skills/core/production-flow/) | Unified development workflow (INIT→PLAN→DEVELOP→VERIFY→REVIEW→SHIP) |
| [discovery](plugins/skillbox/skills/core/discovery/) | Self-questioning + Ralph pattern loop for deep research |
| [secrets-guardian](plugins/skillbox/skills/core/secrets-guardian/) | Multi-layered secrets protection (gitleaks, detect-secrets) |
| [reliable-execution](plugins/skillbox/skills/core/reliable-execution/) | 4-layer persistence for context recovery |
| [agent-harness](plugins/skillbox/skills/core/agent-harness/) | Long-running agent patterns: feature tracking, verification enforcement |

### Go Development

| Skill | Description |
|-------|-------------|
| [go-development](plugins/skillbox/skills/go/go-development/) | Production Go patterns: Database, Cache, Advisory Locks, Services, Repositories |
| [openapi-development](plugins/skillbox/skills/go/openapi-development/) | Spec-first API development with OpenAPI 3.x and oapi-codegen |

### Platform Engineering

| Skill | Description |
|-------|-------------|
| [helm-chart-developer](plugins/skillbox/skills/k8s/helm-chart-developer/) | Production Helm charts with GitOps (Flux + Kustomize + ESO) |
| [flux-gitops-scaffold](plugins/skillbox/skills/k8s/flux-gitops-scaffold/) | Scaffold Flux GitOps projects with image automation |

### Infrastructure Automation

| Skill | Description |
|-------|-------------|
| [ansible-automation](plugins/skillbox/skills/infra/ansible-automation/) | Ansible best practices: project structure, Ubuntu hardening, CI/CD |

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
| [feature-supervisor](plugins/skillbox/agents/core/feature-supervisor.md) | haiku | Orchestrate feature development with feature-dev |
| [verification-worker](plugins/skillbox/agents/core/verification-worker.md) | sonnet | Verify features with RCA and retry |
| [discovery-explorer](plugins/skillbox/agents/core/discovery-explorer.md) | sonnet | Deep 4-phase problem exploration |
| [go-project-init](plugins/skillbox/agents/go/project-init.md) | sonnet | Scaffold Go projects with production patterns |
| [go-test-generator](plugins/skillbox/agents/go/test-generator.md) | sonnet | Generate idiomatic Go tests |
| [go-code-reviewer](plugins/skillbox/agents/go/code-reviewer.md) | sonnet | Review Go code against standards |
| [test-analyzer](plugins/skillbox/agents/tdd/test-analyzer.md) | sonnet | Analyze test coverage |
| [tdd-coach](plugins/skillbox/agents/tdd/tdd-coach.md) | sonnet | Guide TDD workflow |

## Commands

Commands are organized by namespace. Use `/namespace:command` format (e.g., `/go:add-service`).

### Core (`/core:*`)

| Command | Description |
|---------|-------------|
| `/core:commit` | Create conventional commit message |
| `/core:checkpoint` | Save session progress to serena memory |
| `/core:discover` | Self-questioning discovery for problem-solving |
| `/core:loop` | Iterative discovery with Ralph pattern |
| `/core:cancel-loop` | Cancel active discovery loop |
| `/core:init` | Initialize workflow tools (beads + serena) |
| `/core:notify` | Toggle desktop notifications |
| `/core:secrets` | Scan project for secrets |
| `/core:scaffold` | Scaffold a new skill |

### Harness (`/harness:*`)

| Command | Description |
|---------|-------------|
| `/harness:init` | Initialize agent harness with feature tracking |
| `/harness:supervisor` | Automated feature development |
| `/harness:status` | Show feature progress |
| `/harness:verify` | Run verification for features |
| `/harness:update` | Manually update feature status |
| `/harness:auto` | Toggle automatic supervisor mode |

### Go (`/go:*`)

| Command | Description |
|---------|-------------|
| `/go:add-service` | Generate Go service with factory method |
| `/go:add-repository` | Generate Go repository with interface |
| `/go:add-model` | Generate Go model with mapper |
| `/go:review` | Review Go project against production standards |

### OpenAPI (`/openapi:*`)

| Command | Description |
|---------|-------------|
| `/openapi:init` | Initialize modular OpenAPI spec structure |
| `/openapi:add-path` | Add resource path with CRUD operations |
| `/openapi:generate` | Generate Go code from OpenAPI spec |

### Helm (`/helm:*`)

| Command | Description |
|---------|-------------|
| `/helm:scaffold` | Scaffold GitOps structure for app |
| `/helm:validate` | Validate Helm chart |
| `/helm:checkpoint` | Save current Helm work state |

### Flux (`/flux:*`)

| Command | Description |
|---------|-------------|
| `/flux:init` | Initialize Flux GitOps project |
| `/flux:add-app` | Add application with image automation |
| `/flux:add-infra` | Add infrastructure component |
| `/flux:refactor` | Refactor existing GitOps repository |

### Ansible (`/ansible:*`)

| Command | Description |
|---------|-------------|
| `/ansible:scaffold` | Create Ansible project with proper structure |
| `/ansible:validate` | Run lint and security checks |

### TDD (`/tdd:*`)

| Command | Description |
|---------|-------------|
| `/tdd:tdd` | Toggle TDD mode or start TDD workflow |

## Hooks

| Hook | Event | Description |
|------|-------|-------------|
| session_context | SessionStart | Inject project context (Go linter rules, GitOps layout) |
| skill_suggester | SessionStart | Auto-suggest relevant skills |
| git-push-guard | PreToolUse | Confirm before git push |
| golangci-guard | PreToolUse | Protect `.golangci.yml` from modification |
| pretool-secret-guard | PreToolUse | Block secrets in files |
| validate-flux-manifest | PreToolUse | Validate Flux manifests (API versions, required fields) |
| helmrelease-version-check | PostToolUse | Suggest checking HelmRelease chart versions |
| stop-done-criteria | Stop | Quality gate: lint must run if Go files modified |
| stop-tdd-check | Stop | TDD enforcement check on session completion |
| notification | Notification | Desktop notifications via notify-send |
| session_bootstrap | SessionStart | Detect first session, suggest harness initialization |
| pretool-features-guard | PreToolUse | Protect features.json from direct modification |
| stop-verification | Stop | Block session end if unverified features exist |

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
| Ansible project automation | `ansible-automation` (skillbox) |
| Long-running multi-session features | `agent-harness` (skillbox) |

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
│   │   ├── skill-patterns/
│   │   ├── production-flow/
│   │   ├── discovery/           # Self-questioning + Ralph pattern
│   │   ├── secrets-guardian/
│   │   ├── reliable-execution/
│   │   └── agent-harness/       # Long-running agent patterns
│   ├── go/                      # Go development
│   │   ├── go-development/
│   │   └── openapi-development/
│   ├── ts/                      # TypeScript (educational)
│   ├── k8s/                     # Platform engineering
│   │   ├── helm-chart-developer/
│   │   └── flux-gitops-scaffold/
│   └── infra/                   # Infrastructure automation
│       └── ansible-automation/
├── agents/                      # Autonomous agents
│   ├── core/                    # Core agents
│   ├── go/                      # Go agents
│   └── tdd/                     # TDD agents
├── commands/                    # Slash commands (namespaced)
│   ├── core/                    # /core:* commands
│   ├── harness/                 # /harness:* commands
│   ├── go/                      # /go:* commands
│   ├── openapi/                 # /openapi:* commands
│   ├── helm/                    # /helm:* commands
│   ├── flux/                    # /flux:* commands
│   ├── ansible/                 # /ansible:* commands
│   └── tdd/                     # /tdd:* commands
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
