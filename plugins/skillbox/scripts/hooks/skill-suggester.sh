#!/usr/bin/env bash
set -euo pipefail

# skill-suggester.sh — SessionStart hook
# Auto-detects project context and suggests relevant skills

output=""
suggested_skills=""

add_skill() {
    if [ -n "$suggested_skills" ]; then
        suggested_skills+=", "
    fi
    suggested_skills+="$1"
}

# Helm Chart Detection
if [ -f "Chart.yaml" ]; then
    add_skill "skillbox-k8s:helm-chart-developer"
    output+="📦 **Helm Chart detected**"$'\n'
fi

# GitOps Repository Detection
if [ -d "apps/" ] && [ -d "charts/" ]; then
    add_skill "skillbox-k8s:helm-chart-developer"
    output+="🔄 **GitOps repository detected**"$'\n'
    output+="   Use: /helm-scaffold, /helm-validate"$'\n'
fi

# Kustomize Detection
if [ -f "kustomization.yaml" ] || [ -f "kustomization.yml" ]; then
    add_skill "skillbox-k8s:helm-chart-developer"
    output+="🎯 **Kustomize overlay detected**"$'\n'
fi

# Flux Detection
if compgen -G "**/helmrelease*.yaml" > /dev/null 2>&1 || \
   compgen -G "**/kustomization.yaml" > /dev/null 2>&1; then
    if grep -rq "helm.toolkit.fluxcd.io" . 2>/dev/null; then
        output+="⚡ **Flux GitOps detected**"$'\n'
    fi
fi

# Go Project Detection
if [ -f "go.mod" ]; then
    output+="🐹 **Go project detected**"$'\n'
    # Future: add_skill "skillbox-golang:go-conventions"
fi

# Python Project Detection
if [ -f "pyproject.toml" ] || [ -f "requirements.txt" ] || [ -f "setup.py" ]; then
    output+="🐍 **Python project detected**"$'\n'
    # Future: add_skill "skillbox-python:python-conventions"

    # FastAPI detection
    if grep -q "fastapi" pyproject.toml 2>/dev/null || \
       grep -q "fastapi" requirements.txt 2>/dev/null; then
        output+="   Framework: FastAPI"$'\n'
    fi
fi

# Node.js Detection
if [ -f "package.json" ]; then
    output+="📦 **Node.js project detected**"$'\n'
    # Future: add_skill "skillbox-node:node-conventions"
fi

# Rust Detection
if [ -f "Cargo.toml" ]; then
    output+="🦀 **Rust project detected**"$'\n'
    # Future: add_skill "skillbox-rust:rust-conventions"
fi

# Beads Detection
if [ -d ".beads" ]; then
    add_skill "skillbox:beads-workflow"
fi

# Output suggestions
if [ -n "$output" ]; then
    echo -e "$output"
fi

if [ -n "$suggested_skills" ]; then
    echo -e "**Suggested skills:** $suggested_skills"$'\n'
fi
