# claude-skillbox

> Specialized workflow plugins for Claude Code — modular, focused, production-ready.

[![CI](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml/badge.svg)](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml)
[![Version](https://img.shields.io/badge/version-1.0.0-blue?style=flat-square)](https://github.com/11me/claude-skillbox/releases)
[![Python](https://img.shields.io/badge/python-3.12+-blue?style=flat-square&logo=python&logoColor=white)](https://python.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-blueviolet?style=flat-square&logo=anthropic)](https://docs.anthropic.com/en/docs/claude-code)

## Plugins

Install only what you need:

| Plugin | Description |
|--------|-------------|
| **[core](plugins/core/)** | Beads task tracking, Serena navigation, commits, discovery |
| **[go-dev](plugins/go-dev/)** | Go development: services, repositories, OpenAPI |
| **[k8s](plugins/k8s/)** | Kubernetes: Helm charts, Flux GitOps |
| **[ts-dev](plugins/ts-dev/)** | TypeScript: Vitest, Drizzle, conventions |
| **[tdd](plugins/tdd/)** | Test-Driven Development workflow |
| **[infra](plugins/infra/)** | Ansible automation, Ubuntu hardening |
| **[harness](plugins/harness/)** | Long-running agent harness |
| **[python-dev](plugins/python-dev/)** | Python test generation (pytest) |
| **[rust-dev](plugins/rust-dev/)** | Rust test generation |

## Quick Start

```bash
# Add the marketplace
/plugin marketplace add 11me/claude-skillbox

# Install specific plugins
/plugin install core@11me-skillbox
/plugin install go-dev@11me-skillbox
/plugin install k8s@11me-skillbox
```

Or test locally:

```bash
claude --plugin-dir ./plugins/core
claude --plugin-dir ./plugins/go-dev
```

## Plugin Details

### core

Core workflow tools for cross-session development.

**Skills:**
- `beads-workflow` — Task tracking with beads CLI
- `serena-navigation` — Semantic code navigation
- `conventional-commit` — Structured commit messages
- `unified-workflow` — Complete task-to-commit workflow
- `context-engineering` — Long-session context management
- `skill-patterns` — Quality patterns (Do/Verify/Repair)
- `secrets-guardian` — Secrets protection (gitleaks)
- `discovery` — Self-questioning + Ralph pattern

**Commands:** `/commit`, `/checkpoint`, `/discover`, `/loop`, `/init`, `/notify`, `/secrets`, `/scaffold`

**Agents:** task-tracker, session-checkpoint, code-navigator, feature-supervisor, verification-worker, discovery-explorer

---

### go-dev

Go development toolkit with production patterns.

**Skills:**
- `go-development` — Services, repositories, handlers, testing
- `openapi-development` — Spec-first API with oapi-codegen

**Commands:** `/add-service`, `/add-repository`, `/add-model`, `/review`, `/openapi-init`, `/openapi-add-path`, `/openapi-generate`

**Agents:** project-init, test-generator, code-reviewer

---

### k8s

Kubernetes and GitOps toolkit.

**Skills:**
- `helm-chart-developer` — Production Helm charts
- `flux-gitops-scaffold` — Flux GitOps scaffolding
- `flux-gitops-refactor` — Restructure existing GitOps repos

**Commands:** `/helm-scaffold`, `/helm-validate`, `/helm-checkpoint`, `/flux-init`, `/flux-add-app`, `/flux-add-infra`, `/flux-refactor`

---

### ts-dev

TypeScript development patterns.

**Skills:**
- `ts-conventions` — Code conventions
- `ts-database-patterns` — Drizzle ORM
- `ts-project-setup` — pnpm, Biome, Vite
- `ts-testing-patterns` — Vitest
- `ts-type-patterns` — Generics, utility types

**Agents:** project-init, test-generator

---

### tdd

Test-Driven Development workflow.

**Skills:**
- `tdd-enforcer` — Red-Green-Refactor discipline

**Commands:** `/tdd`

**Agents:** tdd-coach, test-analyzer

---

### infra

Infrastructure automation.

**Skills:**
- `ansible-automation` — Ansible practices, Ubuntu hardening

**Commands:** `/ansible-scaffold`, `/ansible-validate`

---

### harness

Long-running agent patterns for multi-session features.

**Skills:**
- `agent-harness` — Feature tracking, verification enforcement

**Commands:** `/harness-init`, `/harness-supervisor`, `/harness-status`, `/harness-verify`, `/harness-update`, `/harness-auto`

---

### python-dev

Python development.

**Agents:** test-writer (pytest patterns)

---

### rust-dev

Rust development.

**Agents:** test-generator

## Architecture

```
plugins/
├── core/                    # Core workflows
│   ├── .claude-plugin/
│   ├── commands/
│   ├── agents/
│   ├── skills/
│   ├── hooks/
│   └── scripts/
├── go-dev/                  # Go development
│   ├── .claude-plugin/
│   ├── commands/
│   ├── agents/
│   ├── skills/
│   └── templates/
├── k8s/                     # Kubernetes/GitOps
│   ├── .claude-plugin/
│   ├── commands/
│   └── skills/
├── ts-dev/                  # TypeScript
│   ├── .claude-plugin/
│   ├── agents/
│   ├── skills/
│   └── templates/
├── tdd/                     # TDD
│   ├── .claude-plugin/
│   ├── commands/
│   ├── agents/
│   └── skills/
├── infra/                   # Infrastructure
│   ├── .claude-plugin/
│   ├── commands/
│   └── skills/
├── harness/                 # Long-running agents
│   ├── .claude-plugin/
│   ├── commands/
│   └── skills/
├── python-dev/              # Python
│   ├── .claude-plugin/
│   └── agents/
└── rust-dev/                # Rust
    ├── .claude-plugin/
    └── agents/
```

## Development

```bash
# Clone
git clone https://github.com/11me/claude-skillbox.git
cd claude-skillbox

# Install pre-commit
uv tool install pre-commit
pre-commit install

# Test a plugin locally
claude --plugin-dir ./plugins/core
```

## License

[MIT](LICENSE)
