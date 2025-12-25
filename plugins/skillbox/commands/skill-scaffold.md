---
description: Scaffold a new Agent Skill directory with SKILL.md template
---

# /skill-scaffold

Create the directory structure and SKILL.md file for a new Agent Skill.

## Usage

```
/skill-scaffold <skill-name> [--type basic|multi|full] [--location skillbox|personal|project]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `skill-name` | Yes | Name in kebab-case (e.g., `api-docs-generator`) |
| `--type` | No | Template type: `basic` (default), `multi`, `full` |
| `--location` | No | Where to create: `skillbox` (default), `personal`, `project` |

## Steps

1. **Validate skill name**:
   - Must be kebab-case: `[a-z0-9-]+`
   - Maximum 64 characters
   - Must be unique in target location

2. **Determine target directory**:
   - `skillbox`: `plugins/skillbox/skills/<skill-name>/`
   - `personal`: `~/.claude/skills/<skill-name>/`
   - `project`: `.claude/skills/<skill-name>/`

3. **Create directory structure** based on type:

   **basic**:
   ```
   skill-name/
   └── SKILL.md
   ```

   **multi**:
   ```
   skill-name/
   ├── SKILL.md
   ├── REFERENCE.md
   └── snippets/
   ```

   **full**:
   ```
   skill-name/
   ├── SKILL.md
   ├── REFERENCE.md
   └── snippets/
   ```
   Plus command file in `commands/<skill-name>-*.md`

4. **Generate SKILL.md** with placeholder content:

```yaml
---
name: <skill-name>
description: <TODO: What this skill does>. Use when <TODO: trigger scenarios>.
---

# <Skill Name>

## Purpose / When to Use

Use this skill when:
- TODO: Scenario 1
- TODO: Scenario 2

## Instructions

TODO: Add step-by-step instructions.

## Examples

Prompts that should activate this skill:

1. "TODO: Example prompt"

## Version History

- 1.0.0 — Initial release
```

5. **Report created files** and next steps.

## Examples

### Basic skill in skillbox

```
/skill-scaffold code-formatter
```

Creates:
```
plugins/skillbox/skills/code-formatter/
└── SKILL.md
```

### Multi-file skill in personal directory

```
/skill-scaffold api-client --type multi --location personal
```

Creates:
```
~/.claude/skills/api-client/
├── SKILL.md
├── REFERENCE.md
└── snippets/
```

### Full skill with command

```
/skill-scaffold k8s-deployer --type full
```

Creates:
```
plugins/skillbox/skills/k8s-deployer/
├── SKILL.md
├── REFERENCE.md
└── snippets/

plugins/skillbox/commands/k8s-deployer-validate.md
```

## After Scaffolding

1. Edit `SKILL.md` to fill in:
   - Description with "what + when" format
   - Purpose / When to Use scenarios
   - Step-by-step instructions
   - Example prompts

2. Add `allowed-tools` if needed:
   ```yaml
   allowed-tools: Read, Grep, Glob
   ```

3. For multi/full types:
   - Add content to `REFERENCE.md`
   - Create snippet files in `snippets/`

4. Test the skill:
   ```bash
   claude --plugin-dir ./plugins/skillbox
   ```
