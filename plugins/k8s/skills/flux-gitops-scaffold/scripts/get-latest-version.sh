#!/bin/bash
# Fetch latest Helm chart version from repository
# Usage: ./get-latest-version.sh <repo-url> <chart-name>
#
# Examples:
#   ./get-latest-version.sh https://charts.jetstack.io cert-manager
#   ./get-latest-version.sh https://kubernetes.github.io/ingress-nginx ingress-nginx

set -euo pipefail

REPO_URL="${1:-}"
CHART_NAME="${2:-}"

if [[ -z "$REPO_URL" ]] || [[ -z "$CHART_NAME" ]]; then
    echo "Usage: $0 <repo-url> <chart-name>" >&2
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo "Error: helm is not installed" >&2
    exit 1
fi

# Create temp repo name
TEMP_REPO="temp-$(date +%s)"

# Add repository temporarily
helm repo add "$TEMP_REPO" "$REPO_URL" --force-update > /dev/null 2>&1

# Search for chart and get latest version
VERSION=$(helm search repo "$TEMP_REPO/$CHART_NAME" --output json 2>/dev/null | \
    jq -r '.[0].version // empty')

# Remove temp repository
helm repo remove "$TEMP_REPO" > /dev/null 2>&1

if [[ -z "$VERSION" ]]; then
    echo "Error: Could not find chart $CHART_NAME in $REPO_URL" >&2
    exit 1
fi

echo "$VERSION"
