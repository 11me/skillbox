#!/bin/bash
# Validate Flux manifest before writing
# Used by PreToolUse hook to catch common issues

set -euo pipefail

# Read input from stdin (hook provides JSON)
input=$(cat)

# Extract file path and content from tool input
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
content=$(echo "$input" | jq -r '.tool_input.content // empty')

# Skip if not a YAML file or no content
if [[ -z "$file_path" ]] || [[ -z "$content" ]]; then
    exit 0
fi

# Only validate Flux-related paths
if [[ ! "$file_path" =~ (clusters|apps|infra).*\.yaml$ ]]; then
    exit 0
fi

errors=()

# Check for deprecated API versions
if echo "$content" | grep -q "helm.toolkit.fluxcd.io/v2beta1"; then
    errors+=("Deprecated API: Use helm.toolkit.fluxcd.io/v2 instead of v2beta1")
fi

if echo "$content" | grep -q "kustomize.toolkit.fluxcd.io/v1beta2"; then
    errors+=("Deprecated API: Use kustomize.toolkit.fluxcd.io/v1 instead of v1beta2")
fi

if echo "$content" | grep -q "source.toolkit.fluxcd.io/v1beta2"; then
    errors+=("Deprecated API: Use source.toolkit.fluxcd.io/v1 instead of v1beta2")
fi

if echo "$content" | grep -q "image.toolkit.fluxcd.io/v1beta"; then
    errors+=("Deprecated API: Use image.toolkit.fluxcd.io/v1 instead of v1beta*")
fi

if echo "$content" | grep -q "external-secrets.io/v1beta1"; then
    errors+=("Deprecated API: Use external-secrets.io/v1 instead of v1beta1")
fi

# Check HelmRelease patterns
if echo "$content" | grep -q "kind: HelmRelease"; then
    # Check for missing interval
    if ! echo "$content" | grep -q "interval:"; then
        errors+=("HelmRelease missing 'interval' field")
    fi

    # Check for sourceRef
    if ! echo "$content" | grep -q "sourceRef:"; then
        errors+=("HelmRelease missing 'sourceRef' in chart.spec")
    fi

    # Check for version in external charts
    if echo "$content" | grep -q "kind: HelmRepository" && ! echo "$content" | grep -q "version:"; then
        errors+=("HelmRelease referencing HelmRepository should specify chart version")
    fi
fi

# Check Kustomization patterns
if echo "$content" | grep -q "kind: Kustomization" && echo "$content" | grep -q "kustomize.toolkit.fluxcd.io"; then
    # Check for sourceRef
    if ! echo "$content" | grep -q "sourceRef:"; then
        errors+=("Flux Kustomization missing 'sourceRef'")
    fi

    # Check for path
    if ! echo "$content" | grep -q "path:"; then
        errors+=("Flux Kustomization missing 'path'")
    fi

    # Check CRD Kustomizations have prune: false
    if echo "$content" | grep -q "name:.*-crds" && ! echo "$content" | grep -q "prune: false"; then
        errors+=("CRD Kustomization should have 'prune: false' to prevent deletion")
    fi
fi

# Check for common mistakes in image automation
if echo "$content" | grep -q "kind: ImagePolicy"; then
    if ! echo "$content" | grep -q "imageRepositoryRef:"; then
        errors+=("ImagePolicy missing 'imageRepositoryRef'")
    fi
fi

# Output result
if [[ ${#errors[@]} -gt 0 ]]; then
    error_msg=$(printf '%s\n' "${errors[@]}")
    printf '{"decision": "deny", "reason": "Flux manifest validation failed: %s"}\n' "$error_msg" >&2
    exit 2
fi

# All checks passed
exit 0
