#!/usr/bin/env bash
set -euo pipefail

# SessionStart hook: inject context at session start
# Provides: date, project type detection, beads tasks, tool checks

output=""

# 1. Inject current date (critical for YAML generation)
output+="**Today:** $(date +%Y-%m-%d)"$'\n\n'

# 2. Detect project type and inject relevant context
PROJECT_TYPE=""

if [ -f "Chart.yaml" ]; then
    PROJECT_TYPE="helm-chart"
    output+="**Project type:** Helm chart"$'\n'
    output+="**Active skills:** helm-chart-developer, conventional-commit"$'\n\n'
elif [ -d "charts/" ] && [ -d "apps/" ]; then
    PROJECT_TYPE="gitops"
    output+="**Project type:** GitOps repository"$'\n'
    output+="**Commands:** /helm-scaffold, /helm-validate, /checkpoint"$'\n\n'
    output+="**Layout:**"$'\n'
    output+='```'$'\n'
    output+="gitops/"$'\n'
    output+="├── charts/app/              # Universal Helm chart"$'\n'
    output+="├── apps/"$'\n'
    output+="│   ├── base/                # Base HelmRelease"$'\n'
    output+="│   ├── dev/                 # Dev overlay: values + patches + ExternalSecret"$'\n'
    output+="│   └── prod/                # Prod overlay: values + patches + ExternalSecret"$'\n'
    output+="└── infra/                   # external-secrets-operator, cert-manager, etc."$'\n'
    output+='```'$'\n\n'
elif [ -f "go.mod" ]; then
    PROJECT_TYPE="go"
    output+="**Project type:** Go project"$'\n\n'
elif [ -f "pyproject.toml" ] || [ -f "requirements.txt" ]; then
    PROJECT_TYPE="python"
    output+="**Project type:** Python project"$'\n\n'
elif [ -f "package.json" ]; then
    PROJECT_TYPE="node"
    output+="**Project type:** Node.js project"$'\n\n'
fi

# Serena Detection
if [ -d ".serena" ]; then
    PROJECT_TYPE="${PROJECT_TYPE:-serena}"
    output+="**Serena enabled** — semantic navigation available"$'\n'
    output+="   Tools: find_symbol, get_symbols_overview, write_memory"$'\n\n'
fi

# 3. Check critical tools
MISSING_TOOLS=""
if [ "$PROJECT_TYPE" = "helm-chart" ] || [ "$PROJECT_TYPE" = "gitops" ]; then
    if ! command -v helm &> /dev/null; then
        MISSING_TOOLS+="helm "
    fi
    if ! command -v kubectl &> /dev/null; then
        MISSING_TOOLS+="kubectl "
    fi
fi

if [ -n "$MISSING_TOOLS" ]; then
    output+="⚠️ **Missing tools:** $MISSING_TOOLS"$'\n\n'
fi

# 4. Beads integration (if available)
if command -v bd &> /dev/null; then
    BD_READY=$(bd ready 2>/dev/null || true)
    if [ -n "$BD_READY" ] && [ "$BD_READY" != "No ready tasks" ]; then
        output+="**Ready tasks:**"$'\n'
        output+='```'$'\n'
        output+="$BD_READY"$'\n'
        output+='```'$'\n\n'
    fi
fi

# 5. GitOps rules reminder (if applicable)
if [ "$PROJECT_TYPE" = "helm-chart" ] || [ "$PROJECT_TYPE" = "gitops" ]; then
    output+="**Rules:**"$'\n'
    output+="- No literal secrets in values.yaml (use ExternalSecret)"$'\n'
    output+="- Use refreshPolicy: OnChange for deterministic ESO updates"$'\n'
    output+="- Validate with /helm-validate before completing work"$'\n'
fi

echo -e "$output"
