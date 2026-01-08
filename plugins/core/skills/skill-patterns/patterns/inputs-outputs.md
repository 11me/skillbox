# Inputs/Outputs Pattern

Explicit contracts that define what a skill expects and produces.

## Pattern Structure

```markdown
## Inputs

| Input | Required | Description |
|-------|----------|-------------|
| app_name | Yes | Application name |
| namespace | Yes | Kubernetes namespace |
| image | No | Container image (default: app_name) |

## Outputs

| Output | Description |
|--------|-------------|
| Chart.yaml | Helm chart metadata |
| values.yaml | Default configuration |
| templates/ | Kubernetes manifests |
```

## When to Use

- Skills with clear input parameters
- Skills that produce specific artifacts
- Complex skills where inputs aren't obvious
- Skills used by other skills (composability)

## Examples

### API Scaffold

```markdown
## Inputs

| Input | Required | Description |
|-------|----------|-------------|
| name | Yes | API/service name |
| framework | No | hono, fastify, express (default: hono) |
| database | No | postgres, sqlite, none (default: postgres) |
| auth | No | Enable authentication (default: true) |

## Outputs

| Output | Description |
|--------|-------------|
| src/index.ts | Entry point |
| src/routes/ | API routes |
| src/db/schema.ts | Database schema (if db enabled) |
| Dockerfile | Container configuration |
```

### Test Generator

```markdown
## Inputs

| Input | Required | Description |
|-------|----------|-------------|
| file | Yes | Source file to generate tests for |
| framework | No | vitest, jest, pytest (auto-detect) |
| coverage | No | Minimum coverage target (default: 80%) |

## Outputs

| Output | Description |
|--------|-------------|
| *.test.ts | Test file in same directory |
| coverage report | Via test runner |
```

## Best Practices

1. **Specify defaults** for optional inputs
2. **Use consistent types** (Yes/No for required)
3. **Describe format** when not obvious
4. **List all artifacts** produced

## Relationship to Prerequisites

- **Prerequisites**: What must exist BEFORE skill runs
- **Inputs**: What the skill needs FROM the user
- **Outputs**: What the skill PRODUCES

```
Prerequisites → Inputs → [Skill Execution] → Outputs
```
