"""Project type detection utilities."""

from pathlib import Path


def find_beads_dir(cwd: Path | None = None) -> Path | None:
    """Find .beads directory by searching up the directory tree.

    This matches beads CLI behavior which auto-discovers the database
    by walking up from cwd until finding .beads/ or hitting a boundary.

    Returns:
        Path to .beads directory if found, None otherwise.
    """
    cwd = cwd or Path.cwd()

    for parent in [cwd, *cwd.parents]:
        beads_dir = parent / ".beads"
        if beads_dir.is_dir():
            return beads_dir
        # Stop at git root or home directory (boundaries)
        if (parent / ".git").is_dir() or parent == Path.home():
            break

    return None


def detect_project_types(cwd: Path | None = None) -> dict[str, bool]:
    """Detect project types based on marker files.

    Returns:
        Dictionary with project type as key and boolean as value.
    """
    cwd = cwd or Path.cwd()

    return {
        "helm": (cwd / "Chart.yaml").exists(),
        "gitops": (cwd / "clusters").is_dir()
        or ((cwd / "apps").is_dir() and (cwd / "charts").is_dir()),
        "kustomize": (cwd / "kustomization.yaml").exists() or (cwd / "kustomization.yml").exists(),
        "go": (cwd / "go.mod").exists(),
        "python": (cwd / "pyproject.toml").exists()
        or (cwd / "requirements.txt").exists()
        or (cwd / "setup.py").exists(),
        "node": (cwd / "package.json").exists(),
        "rust": (cwd / "Cargo.toml").exists(),
        "beads": find_beads_dir(cwd) is not None,
        "serena": (cwd / ".serena").is_dir(),
    }


def detect_flux(cwd: Path | None = None, max_depth: int = 3) -> bool:
    """Detect Flux GitOps by searching for Flux CRDs.

    Args:
        cwd: Working directory to search in.
        max_depth: Maximum directory depth to search (default 3 for performance).

    Returns:
        True if Flux CRDs are found.
    """
    cwd = cwd or Path.cwd()

    def iter_yaml_files(base: Path, depth: int = 0) -> list[Path]:
        """Iterate YAML files up to max_depth."""
        if depth > max_depth:
            return []
        files: list[Path] = []
        try:
            for item in base.iterdir():
                if item.is_file() and item.suffix in (".yaml", ".yml"):
                    files.append(item)
                elif item.is_dir() and not item.name.startswith("."):
                    files.extend(iter_yaml_files(item, depth + 1))
        except (OSError, PermissionError):
            pass
        return files

    for yaml_file in iter_yaml_files(cwd):
        try:
            content = yaml_file.read_text(errors="ignore")
            if "helm.toolkit.fluxcd.io" in content:
                return True
        except (OSError, PermissionError):
            continue

    return False


def has_tests(cwd: Path | None = None) -> bool:
    """Check if project has tests.

    Note: Keep in sync with tdd/scripts/hooks/session_context.py:has_tests()
    """
    cwd = cwd or Path.cwd()

    # Check test directories
    for test_dir in ["tests", "test", "__tests__", "spec"]:
        if (cwd / test_dir).is_dir():
            return True

    # Use recursive glob to find test files in subdirectories
    patterns = [
        "**/*_test.go",
        "**/*_test.py",
        "**/*.test.ts",
        "**/*.spec.ts",
        "**/test_*.py",
    ]
    for pattern in patterns:
        # Use next() with default to short-circuit on first match
        if next(cwd.glob(pattern), None) is not None:
            return True

    return False


def detect_python_framework(cwd: Path | None = None) -> str | None:
    """Detect Python framework if any."""
    cwd = cwd or Path.cwd()

    files_to_check = [cwd / "pyproject.toml", cwd / "requirements.txt"]

    for file_path in files_to_check:
        if file_path.exists():
            try:
                content = file_path.read_text(errors="ignore").lower()
                if "fastapi" in content:
                    return "FastAPI"
                if "django" in content:
                    return "Django"
                if "flask" in content:
                    return "Flask"
            except (OSError, PermissionError):
                continue

    return None


def detect_tdd_mode(cwd: Path | None = None) -> dict[str, bool]:
    """Detect TDD mode status.

    Priority:
    1. Explicit config in .claude/tdd-enforcer.local.md
    2. Auto-detect by presence of test files

    Returns:
        Dictionary with 'enabled' and 'strict' keys.
    """
    cwd = cwd or Path.cwd()
    result = {"enabled": False, "strict": False}

    # Check explicit config
    config_path = cwd / ".claude" / "tdd-enforcer.local.md"
    if config_path.exists():
        try:
            content = config_path.read_text(errors="ignore")
            # Simple frontmatter parsing
            if "enabled: false" in content:
                return result  # Explicitly disabled
            if "enabled: true" in content:
                result["enabled"] = True
            if "strictMode: true" in content:
                result["strict"] = True
            return result
        except (OSError, PermissionError):
            pass

    # Auto-detect by test files
    if has_tests(cwd):
        result["enabled"] = True

    return result
