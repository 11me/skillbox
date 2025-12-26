# Claude Code Instructions for claude-skillbox

## Project Overview

This is a Claude Code plugin providing specialized workflows that complement Anthropic's official plugins.

**Positioning:** Skillbox = Specialized Workflow Layer (not general-purpose)

## When to Use Skillbox vs Official Plugins

| Task | Use |
|------|-----|
| Create new plugin | `plugin-dev` (official) |
| Scaffold TypeScript project | `plugin-dev` (official) |
| Scaffold Go project | `plugin-dev` (official) |
| Track tasks across sessions | `beads-workflow` (skillbox) |
| Navigate code semantically | `serena-navigation` (skillbox) |
| Create Helm charts | `helm-chart-developer` (skillbox) |
| TDD workflow enforcement | `tdd-enforcer` (skillbox) |
| Unified workflow (beads+serena+commit) | `workflow-orchestration` (skillbox) |

**Key principle:** Extend, don't replace. Use official plugins for scaffolding, use Skillbox for specialized workflows.

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes, incompatible API changes
- **MINOR** (0.X.0): New features, backward compatible
- **PATCH** (0.0.X): Bug fixes, documentation updates

**Version file:** `plugins/skillbox/.claude-plugin/plugin.json`

### When to Update Version

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| New skill/command | MINOR | 0.6.0 → 0.7.0 |
| Bug fix | PATCH | 0.6.0 → 0.6.1 |
| Documentation only | PATCH | 0.6.0 → 0.6.1 |
| Breaking change to existing skill | MAJOR | 0.6.0 → 1.0.0 |
| Refactoring (no behavior change) | PATCH | 0.6.0 → 0.6.1 |

**IMPORTANT:** Always update the version when making changes to the plugin.

## README Maintenance

**CRITICAL:** Always update `README.md` when:

1. Adding/removing/renaming skills
2. Adding/removing/renaming commands
3. Adding/removing/renaming hooks
4. Changing directory structure
5. Adding new features or significant changes

The README must accurately reflect:
- All available skills with correct paths
- All available commands
- All hooks with their events
- Current directory structure
- Installation instructions

## Naming Conventions

### Files

- `*.yaml` — Pure YAML files (validated by linters)
- `*.template.yaml` — Go/Helm templates with `{{...}}` syntax (excluded from YAML validation)
- `*.md` — Markdown files
- `SKILL.md` — Skill definition (required frontmatter: name, description)

### Directories

- `plugins/skillbox/skills/core/` — Core skills (git, workflow, etc.)
- `plugins/skillbox/skills/k8s/` — Kubernetes-related skills
- `plugins/skillbox/skills/<domain>/` — Domain-specific skills

### Code Style

- Python: Use `ruff` for linting and formatting
- Shell: Use `shellcheck` for validation
- YAML: Use `yamllint` for validation
- Commits: Follow Conventional Commits spec

## Pre-commit Hooks

Run before committing:

```bash
pre-commit run --all-files
```

## Git Workflow

1. Make changes
2. Run pre-commit hooks
3. Update version if needed
4. Update README if structure changed
5. Commit with conventional commit message
