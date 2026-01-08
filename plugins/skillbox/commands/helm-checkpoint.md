---
name: helm-checkpoint
description: Create a checkpoint summary of current Helm chart work
---

# /helm-checkpoint

Create a checkpoint summary of current Helm chart work.

## Purpose

Generate `CHECKPOINT.md` documenting:
- What was changed
- Which validations were run
- Secret configuration details
- Remaining work

## Usage

```
/helm-checkpoint [app-name] [env]
```

## Output

Creates `apps/<env>/<app>/CHECKPOINT.md`:

```markdown
# Checkpoint: <app-name>

## Date
YYYY-MM-DD HH:MM

## Changes Made
- [ ] Created base HelmRelease
- [ ] Created dev overlay
- [ ] Created ExternalSecret
- [ ] Added ingress configuration

## Validations Run
- [ ] helm lint
- [ ] helm template (defaults)
- [ ] helm template (existingSecretName mode)
- [ ] helm template (externalSecret mode)
- [ ] helm install --dry-run

## Secret Configuration
- **SecretStore**: aws-secrets-manager
- **Secret Path**: project/dev/app
- **Target Secret**: app-dev
- **Namespace**: dev

## Remaining Work
- [ ] Add HPA configuration
- [ ] Configure PDB
- [ ] Add resource limits

## Notes
Any additional context...
```

## When to Use

- Before ending a session
- After completing major changes
- Before handing off to another developer
- When `/helm-validate` passes

## Integration

The `stop-done-criteria` hook checks for checkpoint before allowing session end.
