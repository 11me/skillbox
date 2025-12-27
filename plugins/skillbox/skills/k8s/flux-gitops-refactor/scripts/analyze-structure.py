#!/usr/bin/env python3
"""Analyze Flux GitOps repository structure for refactoring.

Usage:
    python analyze-structure.py /path/to/gitops-repo
    python analyze-structure.py /path/to/gitops-repo --format=markdown
"""

import argparse
import json
import re
import sys
from pathlib import Path


def find_yaml_files(repo_path: Path) -> list[Path]:
    """Find all YAML files in repository."""
    yaml_files = []
    for pattern in ["**/*.yaml", "**/*.yml"]:
        yaml_files.extend(repo_path.glob(pattern))
    return [f for f in yaml_files if ".git" not in str(f)]


def parse_yaml_simple(content: str) -> list[dict]:
    """Simple YAML parser for Kubernetes manifests (no dependencies)."""
    documents = []
    current_doc = {}
    current_key = None

    for line in content.split("\n"):
        stripped = line.strip()

        # Document separator
        if stripped == "---":
            if current_doc:
                documents.append(current_doc)
            current_doc = {}
            continue

        # Skip empty lines and comments
        if not stripped or stripped.startswith("#"):
            continue

        # Parse key-value pairs
        if ":" in line:
            indent = len(line) - len(line.lstrip())
            key_part = stripped.split(":")[0].strip()
            value_part = ":".join(stripped.split(":")[1:]).strip()

            if indent == 0:
                current_key = key_part
                if value_part:
                    current_doc[current_key] = value_part
                else:
                    current_doc[current_key] = {}

    if current_doc:
        documents.append(current_doc)

    return documents


def analyze_helm_release(file_path: Path, content: str) -> dict | None:
    """Analyze HelmRelease for refactoring issues."""
    if "kind: HelmRelease" not in content:
        return None

    # Extract name
    name_match = re.search(r"name:\s*(\S+)", content)
    name = name_match.group(1) if name_match else "unknown"

    # Check for valuesFrom
    has_values_from = "valuesFrom:" in content

    # Check for crds: Skip
    has_crds_skip = "crds: Skip" in content

    # Check for inline values
    has_inline_values = bool(re.search(r"\n\s+values:\s*\n", content))

    # Extract chart name
    chart_match = re.search(r"chart:\s*(\S+)", content)
    chart = chart_match.group(1) if chart_match else None

    # Extract version
    version_match = re.search(r'version:\s*["\']?([^"\'\s]+)', content)
    version = version_match.group(1) if version_match else None

    return {
        "name": name,
        "path": str(file_path),
        "chart": chart,
        "version": version,
        "has_values_from": has_values_from,
        "has_crds_skip": has_crds_skip,
        "has_inline_values": has_inline_values,
    }


def detect_structure_type(repo_path: Path) -> str:
    """Detect the structure type of the repository."""
    has_base = (repo_path / "infra" / "base").exists() or any(
        repo_path.glob("**/base/**/kustomization.yaml")
    )
    has_overlays = any((repo_path / "infra" / env).exists() for env in ["dev", "staging", "prod"])
    has_clusters = (repo_path / "clusters").exists()
    has_kustomize = any(repo_path.glob("**/kustomization.yaml"))

    if has_base and has_overlays and has_clusters:
        return "partial"  # Has structure but may need improvements
    elif has_base or has_overlays:
        return "partial"
    elif has_kustomize:
        return "custom"
    else:
        return "flat"


def detect_environments(repo_path: Path) -> list[str]:
    """Detect environments in the repository."""
    envs = []
    for env in ["dev", "staging", "prod", "production", "test", "qa"]:
        if any(repo_path.glob(f"**/{env}")):
            envs.append(env)
        if any(repo_path.glob(f"**/clusters/{env}")):
            if env not in envs:
                envs.append(env)
    return envs or ["default"]


def categorize_component(name: str, path: str) -> str:
    """Categorize component as controller, config, service, or app."""
    controllers = [
        "cert-manager",
        "ingress-nginx",
        "external-secrets",
        "external-dns",
        "prometheus",
        "grafana",
        "loki",
        "tempo",
        "jaeger",
        "kube-state-metrics",
        "metrics-server",
    ]

    configs = ["cluster-issuer", "cluster-secret-store", "secrets-store"]

    services = ["redis", "postgres", "postgresql", "mysql", "mongodb", "rabbitmq"]

    name_lower = name.lower()

    for c in controllers:
        if c in name_lower:
            return "controllers"

    for c in configs:
        if c in name_lower:
            return "configs"

    for s in services:
        if s in name_lower:
            return "services"

    if "apps/" in path or "/app/" in path:
        return "apps"

    return "apps"  # Default to apps


def find_issues(repo_path: Path, helm_releases: list[dict], structure_type: str) -> list[dict]:
    """Find refactoring issues."""
    issues = []

    for hr in helm_releases:
        # Check for missing valuesFrom
        if not hr["has_values_from"]:
            issues.append(
                {
                    "type": "no_values_from",
                    "component": hr["name"],
                    "path": hr["path"],
                    "severity": "high",
                    "fix": "Add valuesFrom: ConfigMap reference",
                }
            )

        # Check for inline values
        if hr["has_inline_values"]:
            issues.append(
                {
                    "type": "inline_values",
                    "component": hr["name"],
                    "path": hr["path"],
                    "severity": "medium",
                    "fix": "Extract inline values to values.yaml",
                }
            )

        # Check for CRDs handling
        if (
            hr["name"] in ["cert-manager", "external-secrets", "prometheus"]
            and not hr["has_crds_skip"]
        ):
            issues.append(
                {
                    "type": "missing_crds_skip",
                    "component": hr["name"],
                    "path": hr["path"],
                    "severity": "high",
                    "fix": "Add crds: Skip and vendor CRDs separately",
                }
            )

    # Check for missing CRDs directory
    crds_path = repo_path / "infra" / "crds"
    if not crds_path.exists():
        crd_components = [
            hr["name"] for hr in helm_releases if hr["name"] in ["cert-manager", "external-secrets"]
        ]
        if crd_components:
            issues.append(
                {
                    "type": "missing_crds_dir",
                    "components": crd_components,
                    "severity": "high",
                    "fix": "Create infra/crds/ with vendored CRDs",
                }
            )

    # Check for missing aggregator files
    for env in ["dev", "staging", "prod"]:
        controllers_path = repo_path / "infra" / env / "cluster" / "controllers"
        if controllers_path.exists():
            if not (controllers_path / "kustomization.yaml").exists():
                issues.append(
                    {
                        "type": "missing_aggregator",
                        "path": str(controllers_path),
                        "severity": "medium",
                        "fix": "Create kustomization.yaml aggregator",
                    }
                )

    # Check for structure issues
    if structure_type == "flat":
        issues.append(
            {
                "type": "flat_structure",
                "severity": "high",
                "fix": "Create base/overlay directory structure",
            }
        )

    return issues


def analyze_repository(repo_path: Path) -> dict:
    """Analyze GitOps repository structure."""
    yaml_files = find_yaml_files(repo_path)

    # Analyze HelmReleases
    helm_releases = []
    for yaml_file in yaml_files:
        try:
            content = yaml_file.read_text()
            hr = analyze_helm_release(yaml_file.relative_to(repo_path), content)
            if hr:
                helm_releases.append(hr)
        except (UnicodeDecodeError, OSError):
            continue

    # Detect structure and environments
    structure_type = detect_structure_type(repo_path)
    environments = detect_environments(repo_path)

    # Categorize components
    components: dict[str, list[str]] = {
        "controllers": [],
        "configs": [],
        "services": [],
        "apps": [],
    }

    for hr in helm_releases:
        category = categorize_component(hr["name"], hr["path"])
        if hr["name"] not in components[category]:
            components[category].append(hr["name"])

    # Find issues
    issues = find_issues(repo_path, helm_releases, structure_type)

    return {
        "repository": str(repo_path),
        "structure_type": structure_type,
        "environments": environments,
        "components": components,
        "helm_releases": helm_releases,
        "issues": issues,
        "summary": {
            "total_helm_releases": len(helm_releases),
            "total_issues": len(issues),
            "high_severity": len([i for i in issues if i.get("severity") == "high"]),
            "migration_required": structure_type != "standard"
            or any(i.get("severity") == "high" for i in issues),
        },
    }


def format_markdown(report: dict) -> str:
    """Format report as Markdown."""
    lines = [
        "# Flux GitOps Refactoring Analysis Report",
        "",
        f"**Repository:** `{report['repository']}`",
        f"**Structure Type:** {report['structure_type']}",
        f"**Environments:** {', '.join(report['environments'])}",
        "",
        "## Summary",
        "",
        f"- Total HelmReleases: {report['summary']['total_helm_releases']}",
        f"- Total Issues: {report['summary']['total_issues']}",
        f"- High Severity: {report['summary']['high_severity']}",
        f"- Migration Required: {'Yes' if report['summary']['migration_required'] else 'No'}",
        "",
        "## Components",
        "",
    ]

    for category, items in report["components"].items():
        if items:
            lines.append(f"### {category.title()}")
            for item in items:
                lines.append(f"- {item}")
            lines.append("")

    if report["issues"]:
        lines.append("## Issues")
        lines.append("")
        for issue in report["issues"]:
            severity = issue.get("severity", "medium")
            icon = "🔴" if severity == "high" else "🟡"
            lines.append(f"{icon} **{issue['type']}**")
            if "component" in issue:
                lines.append(f"   - Component: {issue['component']}")
            if "path" in issue:
                lines.append(f"   - Path: `{issue['path']}`")
            if "fix" in issue:
                lines.append(f"   - Fix: {issue['fix']}")
            lines.append("")

    if report["helm_releases"]:
        lines.append("## HelmReleases")
        lines.append("")
        lines.append("| Name | Path | valuesFrom | CRDs Skip |")
        lines.append("|------|------|------------|-----------|")
        for hr in report["helm_releases"]:
            vf = "✅" if hr["has_values_from"] else "❌"
            crds = "✅" if hr["has_crds_skip"] else "❌"
            lines.append(f"| {hr['name']} | `{hr['path']}` | {vf} | {crds} |")

    return "\n".join(lines)


def main() -> None:
    parser = argparse.ArgumentParser(description="Analyze Flux GitOps repository structure")
    parser.add_argument("repo_path", help="Path to GitOps repository")
    parser.add_argument(
        "--format",
        choices=["json", "markdown"],
        default="json",
        help="Output format (default: json)",
    )

    args = parser.parse_args()

    repo_path = Path(args.repo_path).resolve()
    if not repo_path.exists():
        print(f"Error: Repository path does not exist: {repo_path}", file=sys.stderr)
        sys.exit(1)

    report = analyze_repository(repo_path)

    if args.format == "markdown":
        print(format_markdown(report))
    else:
        print(json.dumps(report, indent=2))


if __name__ == "__main__":
    main()
