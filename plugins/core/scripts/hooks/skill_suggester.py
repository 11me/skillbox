#!/usr/bin/env python3
"""SessionStart hook: auto-detect project context and suggest skills.

Detects:
- Helm charts and GitOps repositories
- Kustomize overlays
- Flux GitOps
- Go, Python, Node.js, Rust projects
- Beads and Serena integration
"""

import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.detector import detect_flux, detect_project_types, detect_python_framework
from lib.response import session_output


def main() -> None:
    cwd = Path.cwd()
    output_lines: list[str] = []
    suggested_skills: list[str] = []

    types = detect_project_types(cwd)

    # Helm Chart Detection
    if types["helm"]:
        suggested_skills.append("skillbox-k8s:helm-chart-developer")
        output_lines.append("**Helm Chart detected**")

    # GitOps Repository Detection
    if types["gitops"]:
        if "skillbox-k8s:helm-chart-developer" not in suggested_skills:
            suggested_skills.append("skillbox-k8s:helm-chart-developer")
        output_lines.append("**GitOps repository detected**")
        output_lines.append("   Use: /helm-scaffold, /helm-validate")

    # Kustomize Detection
    if types["kustomize"]:
        if "skillbox-k8s:helm-chart-developer" not in suggested_skills:
            suggested_skills.append("skillbox-k8s:helm-chart-developer")
        output_lines.append("**Kustomize overlay detected**")

    # Flux Detection
    if detect_flux(cwd):
        output_lines.append("**Flux GitOps detected**")

    # Go Project Detection
    if types["go"]:
        output_lines.append("**Go project detected**")
        # Future: suggested_skills.append("skillbox-golang:go-conventions")

    # Python Project Detection
    if types["python"]:
        output_lines.append("**Python project detected**")
        # Future: suggested_skills.append("skillbox-python:python-conventions")

        framework = detect_python_framework(cwd)
        if framework:
            output_lines.append(f"   Framework: {framework}")

    # Node.js Detection
    if types["node"]:
        output_lines.append("**Node.js project detected**")
        # Future: suggested_skills.append("skillbox-node:node-conventions")

    # Rust Detection
    if types["rust"]:
        output_lines.append("**Rust project detected**")
        # Future: suggested_skills.append("skillbox-rust:rust-conventions")

    # Beads Detection
    if types["beads"]:
        suggested_skills.append("skillbox:beads-workflow")

    # Serena Detection
    if types["serena"]:
        suggested_skills.append("skillbox:serena-navigation")
        output_lines.append("**Serena project detected**")
        output_lines.append("   Use semantic tools: find_symbol, get_symbols_overview")

    # Output suggestions
    if output_lines:
        output_lines.append("")

    if suggested_skills:
        output_lines.append(f"**Suggested skills:** {', '.join(suggested_skills)}")

    if output_lines:
        session_output("\n".join(output_lines))


if __name__ == "__main__":
    main()
