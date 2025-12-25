---
name: skill-creator
description: Create and scaffold Agent Skills for Claude Code. Use when creating new skills, designing skill structure, writing SKILL.md files, or integrating skills into skillbox plugins. Handles skill scaffolding, best practices, and validation.
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Skill Creator

## Purpose / When to Use

Use this skill when:
- Creating a new Agent Skill from scratch
- Scaffolding skill directory structure
- Writing or improving SKILL.md files
- Designing skill architecture (single-file vs multi-file)
- Adding skills to a skillbox plugin
- Reviewing existing skills for best practices compliance
- Creating commands or hooks for skills

## Skill Creation Workflow

### 1. Understand the Domain

Before creating a skill:
1. Define what capability the skill provides
2. Identify when Claude should use it (trigger scenarios)
3. Determine if it needs supporting files (references, snippets, templates)
4. Decide on tool restrictions (`allowed-tools`)

### 2. Choose Skill Type

| Type | When to Use | Structure |
|------|-------------|-----------|
| Basic | Simple, focused capability | Just SKILL.md |
| Multi-file | Complex with references | SKILL.md + reference files |
| Full | Plugin integration needed | SKILL.md + commands + hooks |

See templates:
- [templates/basic-skill.md](templates/basic-skill.md)
- [templates/multi-file-skill.md](templates/multi-file-skill.md)
- [templates/full-skill-structure.md](templates/full-skill-structure.md)

### 3. Create Directory Structure

```bash
# For skillbox plugin
mkdir -p plugins/skillbox/skills/<skill-name>/

# For personal skills
mkdir -p ~/.claude/skills/<skill-name>/

# For project skills
mkdir -p .claude/skills/<skill-name>/
```

### 4. Write SKILL.md

Every skill requires a SKILL.md file with:

**YAML Frontmatter (required):**
```yaml
---
name: skill-name
description: What it does + when to use it.
allowed-tools: Read, Grep, Glob  # optional
---
```

**Markdown Content:**
```markdown
# Skill Name

## Purpose / When to Use
Clear scenarios when this skill activates.

## Instructions
Step-by-step guidance for Claude.

## Examples
Concrete usage examples and trigger prompts.

## Version History
- 1.0.0 — Initial release
```

See [FRONTMATTER-REFERENCE.md](FRONTMATTER-REFERENCE.md) for detailed rules.

### 5. Add Supporting Files (if needed)

```
skill-name/
├── SKILL.md              # Required
├── REFERENCE.md          # Detailed documentation
├── VERSIONS.md           # API compatibility matrix
├── snippets/             # Ready-to-use code templates
│   └── example.yaml
└── scripts/              # Utility scripts
    └── helper.py
```

Reference files from SKILL.md using relative links:
```markdown
For advanced usage, see [REFERENCE.md](REFERENCE.md).
```

Claude reads additional files only when needed (progressive disclosure).

## Definition of Done

Before completing skill creation:

1. [ ] **Frontmatter valid**: name (kebab-case), description present
2. [ ] **Description specific**: includes "what" AND "when to use"
3. [ ] **Instructions clear**: step-by-step, actionable
4. [ ] **Examples provided**: trigger prompts that activate the skill
5. [ ] **allowed-tools set** (if restricting access)
6. [ ] **Links work**: all referenced files exist
7. [ ] **Tested locally**: skill activates on relevant prompts

## YAML Frontmatter Rules

| Field | Required | Format | Max Length |
|-------|----------|--------|------------|
| `name` | Yes | kebab-case, lowercase | 64 chars |
| `description` | Yes | "What + When" format | 1024 chars |
| `allowed-tools` | No | Comma-separated list | - |

**Name format**: `[a-z0-9-]+` (lowercase, numbers, hyphens only)

**Description format**: `<What this skill does>. Use when <trigger scenarios>.`

See [FRONTMATTER-REFERENCE.md](FRONTMATTER-REFERENCE.md) for complete reference.

## Best Practices Summary

1. **One skill = one capability** — keep focused
2. **Specific descriptions** — include trigger words users would say
3. **Progressive disclosure** — main info in SKILL.md, details in references
4. **Use allowed-tools** — restrict when read-only or limited scope needed
5. **Version history** — track changes for team coordination

See [BEST-PRACTICES.md](BEST-PRACTICES.md) for detailed guidance.

## Examples

Prompts that should activate this skill:

1. "Create a new skill for API documentation"
2. "Scaffold a skill called code-reviewer"
3. "Help me write SKILL.md for my deployment skill"
4. "What's the correct format for skill frontmatter?"
5. "Add a skill to my skillbox plugin"
6. "Review my skill for best practices"
7. "Set up skill structure with commands and hooks"
8. "Create a read-only skill with allowed-tools"

## Related Files

- [BEST-PRACTICES.md](BEST-PRACTICES.md) — Detailed authoring guidance
- [FRONTMATTER-REFERENCE.md](FRONTMATTER-REFERENCE.md) — YAML frontmatter rules
- [templates/](templates/) — Ready-to-use skill templates

## Version History

- 1.0.0 — Initial release with templates and best practices
