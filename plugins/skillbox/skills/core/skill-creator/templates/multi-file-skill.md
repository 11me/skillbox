# Multi-File Skill Template

Use this template for complex skills that need reference documentation, code snippets, or supporting files.

## When to Use

- Skill covers a complex domain
- Detailed reference documentation needed
- Reusable code snippets or templates
- Progressive disclosure benefits the user

## Template Structure

```
{{skill-name}}/
├── SKILL.md              # Main instructions (always read)
├── REFERENCE.md          # Detailed documentation (read when needed)
├── VERSIONS.md           # API/version compatibility (optional)
└── snippets/             # Ready-to-use templates
    ├── example1.yaml
    └── example2.yaml
```

## SKILL.md Template

```yaml
---
name: {{skill-name}}
description: {{What this skill does}}. Use when {{trigger scenarios}}.
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# {{Skill Name}}

## Purpose / When to Use

Use this skill when:
- {{Scenario 1}}
- {{Scenario 2}}
- {{Scenario 3}}

## Quick Start

{{Minimal steps to get started}}

For detailed reference, see [REFERENCE.md](REFERENCE.md).

## Core Workflow

### 1. {{Phase 1}}

{{Brief instructions}}

### 2. {{Phase 2}}

{{Brief instructions}}

### 3. {{Phase 3}}

{{Brief instructions}}

## Common Patterns

### Pattern 1: {{Name}}

```{{language}}
{{code snippet}}
```

See [snippets/](snippets/) for more templates.

## Definition of Done

Before completing work:

1. [ ] {{Validation step 1}}
2. [ ] {{Validation step 2}}
3. [ ] {{Validation step 3}}

## Related Files

- [REFERENCE.md](REFERENCE.md) — Detailed API reference
- [VERSIONS.md](VERSIONS.md) — Version compatibility matrix
- [snippets/](snippets/) — Ready-to-use templates

## Examples

Prompts that should activate this skill:

1. "{{Example prompt 1}}"
2. "{{Example prompt 2}}"
3. "{{Example prompt 3}}"

## Version History

- 1.0.0 — Initial release
```

## REFERENCE.md Template

```markdown
# {{Skill Name}} Reference

Detailed documentation for {{skill-name}}.

## Table of Contents

1. [Concepts](#concepts)
2. [API Reference](#api-reference)
3. [Configuration](#configuration)
4. [Troubleshooting](#troubleshooting)

## Concepts

### {{Concept 1}}

{{Detailed explanation}}

### {{Concept 2}}

{{Detailed explanation}}

## API Reference

### {{API/Command 1}}

**Usage:**
```bash
{{command}}
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| {{param}} | {{type}} | {{description}} |

**Example:**
```bash
{{example}}
```

## Configuration

### {{Config Option 1}}

```yaml
{{config example}}
```

## Troubleshooting

### {{Problem 1}}

**Symptom:** {{what happens}}

**Cause:** {{why it happens}}

**Solution:** {{how to fix}}
```

## Example: API Documentation Skill

```
api-docs-generator/
├── SKILL.md
├── REFERENCE.md
└── snippets/
    ├── openapi-template.yaml
    └── endpoint-template.md
```

### SKILL.md

```yaml
---
name: api-docs-generator
description: Generate OpenAPI documentation from code. Use when creating API docs, documenting endpoints, or generating OpenAPI specs from existing code.
allowed-tools: Read, Grep, Glob, Write, Edit
---

# API Documentation Generator

## Purpose / When to Use

Use this skill when:
- Creating OpenAPI/Swagger documentation
- Documenting REST API endpoints
- Generating API specs from existing code

## Quick Start

1. Analyze existing API code
2. Generate OpenAPI spec using [snippets/openapi-template.yaml](snippets/openapi-template.yaml)
3. Document each endpoint

For detailed OpenAPI reference, see [REFERENCE.md](REFERENCE.md).

...
```
