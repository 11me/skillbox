# Do/Verify/Repair Pattern

Three-phase workflow ensuring skills don't just execute — they validate and fix.

## Pattern Structure

```markdown
## Workflow

### Do (Execute)
1. Perform the main task
2. Create/modify artifacts
3. Complete primary work

### Verify (Validate)
Run these checks:
- `lint command` — check syntax
- `test command` — verify behavior

Acceptance criteria:
- [ ] All checks pass
- [ ] Output is correct

### Repair (If Verify Fails)
1. Read error output
2. Identify root cause
3. Apply minimal fix
4. Re-run Verify
5. Repeat until green
```

## When to Use

- Skills that produce artifacts requiring validation
- Operations where failures can be detected programmatically
- Complex multi-step workflows with checkpoints

## Examples

### Helm Chart Development

```markdown
### Do
1. Create Chart.yaml with metadata
2. Write values.yaml with defaults
3. Generate templates/

### Verify
Run:
- `helm lint .` — syntax check
- `helm template .` — render test
- `helm template . | kubectl apply --dry-run=client -f -` — K8s validation

### Repair
1. Read lint/template errors
2. Fix YAML syntax or missing values
3. Re-run verification
```

### TDD Workflow

The Red-Green-Refactor pattern is a variant of Do/Verify/Repair:

```markdown
### Red (Do)
Write a failing test

### Green (Verify + Repair)
Make the test pass with minimal code

### Refactor (Do again)
Clean up while keeping tests green
```

## Anti-patterns

**Don't use when:**
- Validation is purely manual
- No clear success criteria exist
- Simple operations with no failure modes
