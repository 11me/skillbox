# claude-skillbox

[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-blueviolet?style=flat-square&logo=anthropic)](https://claude.ai)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)
[![CI](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml/badge.svg)](https://github.com/11me/claude-skillbox/actions/workflows/ci.yaml)

Personal skills marketplace for Claude Code.

## Installation

```bash
# Add marketplace
/plugin marketplace add 11me/claude-skillbox

# Install plugin
/plugin install skillbox@11me-skillbox
```

## Skills

### Core Skills

| Skill | Description |
|-------|-------------|
| [conventional-commit](plugins/skillbox/skills/core/conventional-commit/) | Generate git commits following Conventional Commits spec |
| [skill-creator](plugins/skillbox/skills/core/skill-creator/) | Create new Claude Code skills with proper structure |
| [beads-workflow](plugins/skillbox/skills/core/beads-workflow/) | Task tracking with beads (bd CLI) |
| [serena-navigation](plugins/skillbox/skills/core/serena-navigation/) | Semantic code navigation with Serena MCP |

### Kubernetes Skills

| Skill | Description |
|-------|-------------|
| [helm-chart-developer](plugins/skillbox/skills/k8s/helm-chart-developer/) | Build production Helm charts with GitOps (Flux + Kustomize + ESO) |

## Commands

| Command | Description |
|---------|-------------|
| `/commit` | Create git commit with Conventional Commits message |
| `/skill-scaffold` | Scaffold a new skill directory with SKILL.md template |
| `/helm-scaffold` | Scaffold complete GitOps structure for a new app |
| `/helm-validate` | Validate Helm chart (lint, template, dry-run) |
| `/helm-checkpoint` | Create checkpoint summary of current Helm work |

## Hooks

| Hook | Event | Description |
|------|-------|-------------|
| session-context | SessionStart | Inject date, project context, beads status |
| flow-check | SessionStart | Check workflow compliance (CLAUDE.md, pre-commit) |
| skill-suggester | SessionStart | Auto-detect project type and suggest skills |
| git-push-guard | PreToolUse:Bash | Require confirmation before git push |
| pretool-secret-guard | PreToolUse:Write\|Edit | Block secrets in values.yaml |
| prompt-guard | UserPromptSubmit | Block scaffold without required params |
| stop-done-criteria | Stop | Require validation before session end |

## Structure

```
plugins/skillbox/
├── .claude-plugin/
│   └── plugin.json
├── skills/
│   ├── core/
│   │   ├── conventional-commit/
│   │   │   ├── SKILL.md
│   │   │   └── REFERENCE.md
│   │   ├── skill-creator/
│   │   │   ├── SKILL.md
│   │   │   ├── FRONTMATTER-REFERENCE.md
│   │   │   ├── BEST-PRACTICES.md
│   │   │   └── templates/
│   │   ├── beads-workflow/
│   │   │   └── SKILL.md
│   │   └── serena-navigation/
│   │       └── SKILL.md
│   └── k8s/
│       ├── .claude-plugin/
│       │   └── plugin.json
│       └── helm-chart-developer/
│           ├── SKILL.md
│           ├── reference-gitops-eso.md
│           ├── VERSIONS.md
│           └── snippets/
├── commands/
│   ├── commit.md
│   ├── skill-scaffold.md
│   ├── helm-scaffold.md
│   ├── helm-validate.md
│   └── helm-checkpoint.md
├── hooks/
│   └── hooks.json
└── scripts/
    ├── validate-helm.sh
    └── hooks/
        ├── session-context.sh
        ├── flow-check.sh
        ├── skill-suggester.sh
        ├── git-push-guard.py
        ├── pretool-secret-guard.py
        ├── prompt-guard.py
        └── stop-done-criteria.py
```

## Development

### Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/)

### Setup

```bash
# Install pre-commit via uv
uv tool install pre-commit

# Setup git hooks
pre-commit install

# Run all checks (same as CI)
pre-commit run --all-files
```

### Testing Locally

```bash
# Test plugin locally
claude --plugin-dir ./plugins/skillbox
```

### Adding a New Skill

1. Create directory: `plugins/skillbox/skills/<domain>/<skill-name>/`
2. Add `SKILL.md` with required frontmatter (name, description)
3. Add supporting files (snippets, references)
4. Update this README
5. Test locally

## Versioning

This project follows [Semantic Versioning](https://semver.org/).

See [CLAUDE.md](CLAUDE.md) for versioning guidelines.

## License

MIT
