---
name: init-workflow
description: Initialize workflow tools for existing project (beads + serena + CLAUDE.md)
---

# /init-workflow

Initialize workflow tools for an existing project.

> **Note:** For project scaffolding (TypeScript, Go, etc.), use Anthropic's `plugin-dev` plugin.
> This command sets up workflow tooling on top of any project.

## What This Does

1. **Beads** — Cross-session task tracking with `bd` CLI
2. **Serena** — Code memory and semantic navigation
3. **CLAUDE.md** — Project-specific AI instructions
4. **Pre-commit** — Code quality hooks (optional)

## Usage

Run the initialization script:

```bash
# Full setup (recommended)
python3 ${CLAUDE_PLUGIN_ROOT}/scripts/init-project.py

# With custom project name
python3 ${CLAUDE_PLUGIN_ROOT}/scripts/init-project.py --name "My Project"

# Minimal (beads + CLAUDE.md only)
python3 ${CLAUDE_PLUGIN_ROOT}/scripts/init-project.py --minimal

# Skip pre-commit
python3 ${CLAUDE_PLUGIN_ROOT}/scripts/init-project.py --skip-precommit
```

## After Initialization

1. Review `CLAUDE.md` and customize for your project
2. Update `.serena/memories/overview.md` with architecture details
3. Create first task: `bd create --title "..." -t task`
4. Commit: `git add -A && git commit -m "chore: init workflow"`

## Checklist

```
[ ] .beads/ exists (bd init)
[ ] .serena/project.yml exists
[ ] .serena/memories/overview.md written
[ ] CLAUDE.md with quick start
[ ] .pre-commit-config.yaml configured (optional)
[ ] First commit made
```

## See Also

- **Project scaffolding:** Use `plugin-dev` from Anthropic official plugins
- **Workflow patterns:** See `workflow-orchestration` skill
