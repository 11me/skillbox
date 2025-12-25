# Contributing to claude-skillbox

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/) — fast Python package manager
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) — for testing skills

### Getting Started

```bash
# Clone the repository
git clone https://github.com/11me/claude-skillbox.git
cd claude-skillbox

# Install pre-commit hooks
uv tool install pre-commit
pre-commit install

# Verify setup
pre-commit run --all-files
```

### Testing Locally

```bash
claude --plugin-dir ./plugins/skillbox
```

## Adding a New Skill

### 1. Create Skill Directory

```bash
# Use the skill-scaffold command in Claude Code
/skill-scaffold

# Or manually create the structure
mkdir -p plugins/skillbox/skills/<domain>/<skill-name>/
```

### 2. Write SKILL.md

Every skill requires a `SKILL.md` file with YAML frontmatter:

```markdown
---
name: my-skill
description: What it does. Use when trigger scenarios.
allowed-tools: Read, Grep, Glob
---

# My Skill

## Purpose / When to Use

Use this skill when:
- Scenario 1
- Scenario 2

## Instructions

Step-by-step guidance for Claude.

## Examples

Trigger prompts:
- "example phrase 1"
- "example phrase 2"

## Version History

- 1.0.0 — Initial release
```

See [skill-creator](plugins/skillbox/skills/core/skill-creator/SKILL.md) for detailed templates and best practices.

### 3. Update Documentation

- Add skill to `README.md` skills table
- Update `plugins/skillbox/skills/_index.md`

### 4. Test Your Skill

```bash
# Validate frontmatter
uv run python scripts/validate-skills.py plugins/skillbox/skills/

# Test with Claude Code
claude --plugin-dir ./plugins/skillbox
```

## Writing Hooks

All hooks must be written in Python. See [HOOKS.md](plugins/skillbox/HOOKS.md) for the complete guide.

### Guidelines

- Use shared utilities from `scripts/hooks/lib/`
- Follow existing patterns for JSON input/output
- Keep hooks focused and deterministic
- No external dependencies (stdlib only)

## Code Style

### Python

We use [ruff](https://github.com/astral-sh/ruff) for linting and formatting:

```bash
uvx ruff format .
uvx ruff check .
```

### Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `build`, `ci`, `chore`

Examples:
- `feat(skills): add python-conventions skill`
- `fix(hooks): handle missing beads directory`
- `docs: update README installation instructions`

## Pull Request Process

1. Fork the repository
2. Create a branch: `git checkout -b feat/my-feature`
3. Make changes and commit
4. Run checks: `pre-commit run --all-files`
5. Push and create a Pull Request

### PR Checklist

- [ ] Pre-commit hooks pass
- [ ] New skills have valid frontmatter
- [ ] README updated if needed
- [ ] Commits follow Conventional Commits

## Versioning

The project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New skills, commands, or features
- **PATCH**: Bug fixes, documentation updates

## Questions?

Open an [issue](https://github.com/11me/claude-skillbox/issues) for questions or suggestions.
