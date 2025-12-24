# Contributing

## Adding a New Skill

1. Create a directory: `plugins/skillbox/skills/<skill-name>/`
2. Add `SKILL.md` with frontmatter:

```yaml
---
name: skill-name
description: What this skill does + when Claude should use it.
# allowed-tools: Read, Grep, Glob  # optional
---
```

3. Test locally:

```bash
claude --plugin-dir ./plugins/skillbox
```

4. Submit a PR

## Skill Guidelines

- Keep skills focused (one capability per skill)
- Write clear descriptions that help Claude know when to use it
- Include concrete examples
- Use `allowed-tools` to restrict access when appropriate

## Naming

- Use kebab-case for skill names
- Name should reflect what the skill does
