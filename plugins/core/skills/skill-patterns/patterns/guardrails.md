# Guardrails Pattern

Non-negotiable NEVER/MUST rules that define safety boundaries.

## Pattern Structure

```markdown
## Guardrails

**NEVER:**
- Never do X (risk: explanation)
- Never modify Y without Z
- Never skip validation

**MUST:**
- Always do Z before completing
- Always check W
- Always document changes
```

## Key Principles

1. **Guardrails are not suggestions** — they're enforced rules
2. **Include risk explanation** — why the rule exists
3. **Be specific** — vague rules are ignored
4. **Keep list short** — 3-5 rules per category max

## Examples

### Helm Charts

```markdown
## Guardrails

**NEVER:**
- Never put secrets in values.yaml (use ExternalSecret instead)
- Never use `latest` tag in production images
- Never skip `helm lint` before completing

**MUST:**
- Always run `helm template` to verify rendering
- Always use semantic versioning for chart version
- Always document breaking changes in CHANGELOG
```

### Git Operations

```markdown
## Guardrails

**NEVER:**
- Never force-push to main/master
- Never commit `.env` files or credentials
- Never skip pre-commit hooks

**MUST:**
- Always write descriptive commit messages
- Always reference issue/task in commit
- Always run tests before pushing
```

### Database Migrations

```markdown
## Guardrails

**NEVER:**
- Never DROP TABLE without backup confirmation
- Never run migrations without testing in staging
- Never modify production data directly

**MUST:**
- Always create reversible migrations
- Always test rollback procedure
- Always backup before destructive changes
```

## Writing Effective Guardrails

| Good | Bad |
|------|-----|
| Never put AWS keys in code | Be careful with secrets |
| Always run lint before commit | Check the code |
| Never deploy on Friday after 4pm | Be cautious with deployments |

## Integration with Hooks

Guardrails can be enforced via hooks:

```python
# PreToolUse hook
if "password" in content and "values.yaml" in file_path:
    block("Secrets detected in values.yaml — use ExternalSecret")
```

See [../../hooks/](../../hooks/) for hook implementations.
