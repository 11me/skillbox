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
| *coming soon* | |

## Local Development

```bash
# Test plugin locally
claude --plugin-dir ./plugins/skillbox

# Validate
/plugin validate .
```

## Adding a New Skill

1. Create a directory in `plugins/skillbox/skills/<skill-name>/`
2. Add `SKILL.md` (see `templates/SKILL.template.md`)
3. Test locally with `claude --plugin-dir ./plugins/skillbox`

## License

MIT
