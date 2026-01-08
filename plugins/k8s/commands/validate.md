---
name: validate
description: Validate Helm chart with all required checks
argument-hint: "[chart-path]"
---

# /helm-validate

Validate Helm chart with all required checks.

## Checks Performed

1. **helm lint** - Static analysis
2. **helm template (defaults)** - Render with default values
3. **helm template (GitOps mode)** - Render with existingSecretName
4. **helm template (ESO mode)** - Render with externalSecret.enabled
5. **helm install --dry-run** - Server-side validation

## Usage

```
/helm-validate [chart-path]
```

Default: current directory

## Script

Runs `scripts/validate-helm.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

CHART_DIR="${1:-.}"
RELEASE_NAME="${RELEASE_NAME:-test-release}"

cd "${CHART_DIR}"

echo "==> helm lint"
helm lint .

echo "==> helm template (defaults)"
helm template "${RELEASE_NAME}" . >/dev/null

echo "==> helm template (GitOps mode: existingSecretName)"
helm template "${RELEASE_NAME}" . \
  --set secrets.existingSecretName=fake-secret \
  --set secrets.inject.envFrom=true >/dev/null

echo "==> helm template (Chart-managed ExternalSecret)"
helm template "${RELEASE_NAME}" . \
  --set secrets.externalSecret.enabled=true \
  --set secrets.externalSecret.secretStoreRef.name=aws-secrets-manager \
  --set secrets.externalSecret.dataFrom.extractKey=fake/path >/dev/null

echo "==> helm install --dry-run --debug"
helm install "${RELEASE_NAME}" . --dry-run --debug >/dev/null

echo "✅ Chart validations passed"
```

## Exit Codes

- `0` - All validations passed
- `1` - Validation failed (check output for details)

## Prerequisites

- `helm` CLI installed
- Kubernetes context configured (for dry-run)
