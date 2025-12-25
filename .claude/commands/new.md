---
description: Create a new skill in skillbox using skill-creator
---

# /new

Create a new Agent Skill in the skillbox plugin.

## Usage

```
/new <skill-name>
```

## Instructions

When this command is invoked:

1. **Validate the skill name:**
   - Must be kebab-case: `[a-z0-9-]+`
   - Maximum 64 characters
   - Check it doesn't already exist in `plugins/skillbox/skills/`

2. **Read the skill-creator resources:**
   - Read `plugins/skillbox/skills/skill-creator/templates/basic-skill.md` for the template structure
   - Read `plugins/skillbox/skills/skill-creator/BEST-PRACTICES.md` for writing good descriptions
   - Read `plugins/skillbox/skills/skill-creator/FRONTMATTER-REFERENCE.md` for YAML rules

3. **Ask the user:**
   - What does this skill do? (for description)
   - When should Claude use it? (trigger scenarios)
   - Should it have tool restrictions? (allowed-tools)

4. **Create the skill structure:**
   ```
   plugins/skillbox/skills/<skill-name>/
   └── SKILL.md
   ```

5. **Generate SKILL.md** following the template with:
   - Proper YAML frontmatter (name, description)
   - Purpose / When to Use section
   - Instructions section
   - Examples section (trigger prompts)
   - Version History

6. **Report success** and suggest next steps:
   - Test the skill locally
   - Add reference files if needed
   - Consider adding a command for the skill

## Example

```
/new api-docs-generator
```

Creates `plugins/skillbox/skills/api-docs-generator/SKILL.md` after gathering info from user.
