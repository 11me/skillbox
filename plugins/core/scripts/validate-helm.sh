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
