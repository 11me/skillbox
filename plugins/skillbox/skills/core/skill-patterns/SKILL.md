---
name: skill-patterns
description: Quality patterns for robust Claude Code skills. Use when improving existing skills, adding validation workflows, defining safety guardrails, or documenting skill contracts. Provides Do/Verify/Repair, Guardrails, Inputs/Outputs, Prerequisites, and Scope patterns.
---

# Skill Patterns Reference

## Purpose / When to Use

Use this reference when:
- Improving an existing skill with structured workflows
- Adding validation and repair cycles to a skill
- Defining safety guardrails (NEVER/MUST rules)
- Documenting skill inputs and outputs
- Clarifying skill scope and boundaries

**Not for:** Creating new skills from scratch — use Anthropic's `plugin-dev` plugin.

## Available Patterns

| Pattern | Purpose | When to Apply |
|---------|---------|---------------|
| [Do/Verify/Repair](patterns/do-verify-repair.md) | Three-phase workflow | Skills with validation requirements |
| [Guardrails](patterns/guardrails.md) | NEVER/MUST constraints | Safety-critical operations |
| [Inputs/Outputs](patterns/inputs-outputs.md) | Contract tables | Skills with explicit I/O |
| [Prerequisites](patterns/prerequisites.md) | Required environment | Skills needing tools/files |
| [Scope](patterns/scope-boundaries.md) | In/Out boundaries | Preventing skill overreach |

## Quick Reference

### Pattern Selection Guide

```
Does the skill need validation?
├── Yes → Do/Verify/Repair
└── No
    ├── Has safety rules? → Guardrails
    ├── Clear I/O contract? → Inputs/Outputs
    ├── Needs specific tools/files? → Prerequisites
    └── Risk of overreach? → Scope
```

### Combining Patterns

Most production skills combine multiple patterns:

```markdown
## Prerequisites          ← from prerequisites.md
## Inputs / Outputs       ← from inputs-outputs.md
## Workflow               ← from do-verify-repair.md
## Guardrails             ← from guardrails.md
## Scope                  ← from scope-boundaries.md
```

## Real-World Applications

| Skill | Patterns Used |
|-------|---------------|
| helm-chart-developer | All five patterns |
| tdd-enforcer | Do/Verify/Repair (Red-Green-Refactor) |
| conventional-commit | Guardrails (commit message rules) |

## Template

See [templates/enhanced-skill.md](templates/enhanced-skill.md) for a complete skill template with all patterns integrated.

## Related

- **Creating new skills**: Use `plugin-dev` from Anthropic's official plugins
- **Skills registry**: See [../_index.md](../_index.md)

## Version History

- 1.0.0 — Extracted from skill-creator as standalone reference
