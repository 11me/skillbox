# Scope Boundaries Pattern

Define what a skill handles and what it explicitly does not.

## Pattern Structure

```markdown
## Scope

**In Scope:**
- Task A
- Task B
- Task C

**Out of Scope:**
- Task D → use `other-skill` instead
- Task E → requires manual intervention
- Task F → handled by external tool
```

## Purpose

1. **Prevent overreach** — skill stays focused
2. **Guide users** — know where to go for other tasks
3. **Avoid conflicts** — multiple skills don't overlap

## Examples

### Helm Chart Developer

```markdown
## Scope

**In Scope:**
- Creating new Helm charts
- Validating chart syntax and templates
- GitOps overlay structure (Flux, Kustomize)
- ExternalSecret resources

**Out of Scope:**
- CI/CD pipeline setup → use deployment automation
- Cluster provisioning → requires manual infrastructure
- Secret management backends → use external-secrets docs
- Container image building → use docker/buildah directly
```

### TypeScript API

```markdown
## Scope

**In Scope:**
- API route scaffolding
- Request/response validation (Zod)
- Database schema (Drizzle)
- Authentication middleware

**Out of Scope:**
- Frontend components → use frontend-skill
- Deployment configuration → use k8s-skill
- CI/CD setup → use github-actions-skill
- Performance tuning → requires profiling
```

### Code Review

```markdown
## Scope

**In Scope:**
- Syntax and style issues
- Type safety problems
- Security vulnerabilities (OWASP)
- Performance anti-patterns

**Out of Scope:**
- Business logic correctness → requires domain knowledge
- Architecture decisions → use architecture-skill
- Test coverage → use test-analyzer agent
```

## Writing Effective Scope

### Good Practices

1. **Be specific** about what's included
2. **Provide alternatives** for out-of-scope items
3. **Explain why** something is out of scope
4. **Keep lists balanced** (not everything out of scope)

### Anti-patterns

| Bad | Better |
|-----|--------|
| Everything related to K8s | Helm charts, validation, GitOps overlays |
| Not database stuff | Database schema design → use db-skill |
| Complex things | Multi-cluster federation → requires manual setup |

## Relationship to Guardrails

- **Scope**: What the skill CAN do
- **Guardrails**: What the skill MUST NOT do

```
Scope: Define boundaries
Guardrails: Enforce safety within those boundaries
```
