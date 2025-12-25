# claude-skillbox

[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-blueviolet?style=flat-square&logo=anthropic)](https://claude.ai)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)

Personal skills marketplace for Claude Code.

## Installation

```bash
# Add marketplace
/plugin marketplace add 11me/claude-skillbox

# Install plugin
/plugin install skillbox@11me-skillbox
```

## Skills

| Skill | Description |
|-------|-------------|
| [conventional-commit](plugins/skillbox/skills/conventional-commit/) | Generate beautiful git commits following Conventional Commits spec |
| [helm-chart-developer](plugins/skillbox/skills/helm-chart-developer/) | Build production Helm charts with GitOps (Flux + Kustomize + ESO) |

## Commands

| Command | Description |
|---------|-------------|
| `/commit` | Create git commit with Conventional Commits message |
| `/helm-scaffold` | Scaffold complete GitOps structure for a new app |
| `/helm-validate` | Validate Helm chart (lint, template, dry-run) |
| `/checkpoint` | Create checkpoint summary of current work |

## Hooks

| Hook | Event | Description |
|------|-------|-------------|
| git-push-guard | PreToolUse | Require user confirmation before git push |
| secret-prevent | PreToolUse | Block secrets in values.yaml |
| session-context | SessionStart | Inject GitOps layout and rules |
| prompt-guard | UserPromptSubmit | Block scaffold without required params |
| stop-done-criteria | Stop | Require validation before session end |

## Structure

```
plugins/skillbox/
├── skills/
│   ├── conventional-commit/
│   │   ├── SKILL.md
│   │   └── REFERENCE.md
│   └── helm-chart-developer/
│       ├── SKILL.md
│       ├── reference-gitops-eso.md
│       └── snippets/
├── commands/
│   ├── commit.md
│   ├── helm-scaffold.md
│   ├── helm-validate.md
│   └── checkpoint.md
├── hooks/
│   └── hooks.json
└── scripts/
    ├── validate-helm.sh
    └── hooks/
        └── git-push-guard.py
```

## Local Development

```bash
# Test plugin locally
claude --plugin-dir ./plugins/skillbox

# Validate marketplace
/plugin validate .
```

## Adding a New Skill

1. Create directory: `plugins/skillbox/skills/<skill-name>/`
2. Add `SKILL.md` (see `templates/SKILL.template.md`)
3. Add supporting files (snippets, references)
4. Test locally with `claude --plugin-dir ./plugins/skillbox`

## License

MIT
