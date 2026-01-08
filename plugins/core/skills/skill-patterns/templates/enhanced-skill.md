# Enhanced Skill Template

Complete template integrating all skill patterns.

## Directory Structure

```
skill-name/
├── SKILL.md              # Main skill definition (below)
├── REFERENCE.md          # Detailed documentation (optional)
└── snippets/             # Code templates (optional)
```

## Template

```markdown
---
name: skill-name
description: What it does. Use when trigger scenarios.
allowed-tools: Read, Grep, Glob
globs: ["**/*.ext"]
---

# Skill Name

## Purpose / When to Use

Use this skill when:
- Trigger scenario 1
- Trigger scenario 2
- User mentions: "keyword1", "keyword2"

## Prerequisites

**Required tools:**
- tool1 (version X+)
- tool2

**Required files:**
- File pattern or structure

**Environment:**
- Assumptions about context

## Inputs

| Input | Required | Description |
|-------|----------|-------------|
| context | Yes | What user provides |
| files | No | Optional data |

## Outputs

| Output | Description |
|--------|-------------|
| artifact1 | What skill produces |
| artifact2 | Secondary output |

## Workflow

### Do (Execute)

1. Step 1 — what to do
2. Step 2 — next action
3. Step 3 — complete task

### Verify (Validate)

Run these checks:
- `command1` — validates X
- `command2` — validates Y

Acceptance criteria:
- [ ] Criterion 1 passes
- [ ] Criterion 2 passes

### Repair (If Verify Fails)

1. Read error logs/output
2. Identify root cause
3. Apply minimal fix
4. Re-run Verify
5. Repeat until green

## Guardrails

**NEVER:**
- Never do X (risk explanation)
- Never modify Y without Z

**MUST:**
- Always do Z before completing
- Always check W

## Scope

**In Scope:**
- Task A
- Task B

**Out of Scope:**
- Task C → use `other-skill` instead
- Task D → requires manual intervention

## Examples

Trigger prompts:
- "trigger phrase 1"
- "trigger phrase 2"

## Related Skills

- **skill-name** — when to use instead

## Version History

- 1.0.0 — Initial release
```

## When to Use This Template

Choose **enhanced template** when:
- Skill has multiple phases (do/verify/repair)
- Clear input/output contracts are needed
- Safety guardrails are important
- Skill interacts with other skills

Choose **basic template** (just frontmatter + Purpose + Examples) when:
- Simple, single-purpose skill
- No validation phase needed
- Less than 50 lines of instructions

## Pattern Checklist

Before completing a skill:

- [ ] Prerequisites documented (if any)
- [ ] Inputs/Outputs tables (if applicable)
- [ ] Do/Verify/Repair workflow (if validation needed)
- [ ] Guardrails defined (if safety-critical)
- [ ] Scope boundaries clear (if risk of overreach)
- [ ] Examples provided (always)
- [ ] Version history started (always)
