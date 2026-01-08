"""Bootstrap detection and generation for long-running agent harness.

This module provides:
- First-session detection for projects
- Project-type-aware startup command generation
- Harness state file management
- init-session.sh script generation

Key concept from Anthropic's article: "Initializer Agent" sets up environment
on first session, subsequent sessions use cached state for quick startup.
"""

import json
from datetime import datetime
from pathlib import Path


def detect_first_session(project_dir: Path | None = None) -> bool:
    """Detect if this is the first Claude session for this project.

    Checks for presence of harness state file. Returns True if this is
    the first session (no harness initialized yet).

    Args:
        project_dir: Project root directory (defaults to cwd)

    Returns:
        True if this is the first session (harness not initialized)
    """
    if project_dir is None:
        project_dir = Path.cwd()

    harness_file = project_dir / ".claude" / "harness.json"
    return not harness_file.exists()


def get_project_type(project_dir: Path) -> str | None:
    """Detect project type from files.

    Returns:
        Project type string or None if unknown
    """
    if (project_dir / "go.mod").exists():
        return "go"
    if (project_dir / "Cargo.toml").exists():
        return "rust"
    if (project_dir / "package.json").exists():
        if (project_dir / "pnpm-lock.yaml").exists():
            return "node-pnpm"
        if (project_dir / "yarn.lock").exists():
            return "node-yarn"
        return "node-npm"
    if (project_dir / "pyproject.toml").exists():
        return "python-uv"
    if (project_dir / "requirements.txt").exists():
        return "python-pip"
    return None


def get_project_startup_commands(project_dir: Path) -> list[str]:
    """Generate init.sh equivalent commands for project.

    Returns list of shell commands to bootstrap the project.
    """
    project_type = get_project_type(project_dir)
    commands: list[str] = []

    if project_type == "go":
        commands.extend(
            [
                "go mod download",
                "go build ./...",
            ]
        )
        if (project_dir / "Makefile").exists():
            commands.append("make setup 2>/dev/null || true")

    elif project_type == "rust":
        commands.append("cargo build")

    elif project_type and project_type.startswith("node"):
        pm = "pnpm" if "pnpm" in project_type else ("yarn" if "yarn" in project_type else "npm")
        commands.extend(
            [
                f"{pm} install",
                f"{pm} run build 2>/dev/null || true",
            ]
        )

    elif project_type == "python-uv":
        commands.append("uv sync")

    elif project_type == "python-pip":
        commands.append("pip install -e . 2>/dev/null || pip install -r requirements.txt")

    return commands


def get_default_verification_command(project_dir: Path, feature_id: str) -> str | None:
    """Get default verification command for project type.

    Args:
        project_dir: Project root
        feature_id: Feature ID to use in test pattern

    Returns:
        Default test command or None
    """
    project_type = get_project_type(project_dir)

    if project_type == "go":
        return f"go test ./... -run {_to_test_pattern(feature_id)}"

    if project_type == "rust":
        return f"cargo test {feature_id.replace('-', '_')}"

    if project_type and project_type.startswith("node"):
        pm = "pnpm" if "pnpm" in project_type else ("yarn" if "yarn" in project_type else "npx")
        return (
            f"{pm} vitest run --grep '{feature_id}' 2>/dev/null || "
            f"{pm} jest --testNamePattern '{feature_id}'"
        )

    if project_type and project_type.startswith("python"):
        return f"pytest -k {feature_id.replace('-', '_')}"

    return None


def _to_test_pattern(feature_id: str) -> str:
    """Convert feature ID to test pattern.

    auth-login → AuthLogin (for Go TestAuthLogin)
    """
    parts = feature_id.split("-")
    return "".join(part.capitalize() for part in parts)


def create_harness_state(project_dir: Path) -> Path:
    """Create initial harness.json state file.

    Returns path to created file.
    """
    harness = {
        "version": "1.0.0",
        "created": datetime.now().isoformat(),
        "project_type": get_project_type(project_dir),
        "initialized": True,
        "sessions": [
            {
                "id": 1,
                "started": datetime.now().isoformat(),
                "type": "initializer",
            }
        ],
    }

    claude_dir = project_dir / ".claude"
    claude_dir.mkdir(parents=True, exist_ok=True)

    harness_file = claude_dir / "harness.json"
    harness_file.write_text(json.dumps(harness, indent=2))

    return harness_file


def increment_session(project_dir: Path) -> int:
    """Increment session counter in harness.json.

    Returns new session number.
    """
    harness_file = project_dir / ".claude" / "harness.json"
    if not harness_file.exists():
        return 1

    try:
        harness = json.loads(harness_file.read_text())
    except (json.JSONDecodeError, OSError):
        return 1

    sessions = harness.get("sessions", [])
    new_session_id = len(sessions) + 1

    sessions.append(
        {
            "id": new_session_id,
            "started": datetime.now().isoformat(),
            "type": "coding",
        }
    )

    harness["sessions"] = sessions
    harness_file.write_text(json.dumps(harness, indent=2))

    return new_session_id


def get_session_count(project_dir: Path) -> int:
    """Get current session count from harness.json."""
    harness_file = project_dir / ".claude" / "harness.json"
    if not harness_file.exists():
        return 0

    try:
        harness = json.loads(harness_file.read_text())
        return len(harness.get("sessions", []))
    except (json.JSONDecodeError, OSError):
        return 0


def generate_init_script(project_dir: Path) -> Path:
    """Generate init-session.sh bootstrap script.

    Returns path to generated script.
    """
    commands = get_project_startup_commands(project_dir)

    script_lines = [
        "#!/bin/bash",
        "# Auto-generated by skillbox harness",
        "# Regenerate with: /harness-init --regenerate",
        "",
        "set -e",
        "",
        'PROJECT_DIR="$(dirname "$0")/.."',
        "",
        "# 1. Change to project directory",
        'cd "$PROJECT_DIR"',
        "",
        "# 2. Load environment if available",
        "source .env 2>/dev/null || true",
        "",
    ]

    if commands:
        script_lines.extend(
            [
                "# 3. Dependencies and build",
            ]
        )
        for cmd in commands:
            script_lines.append(cmd)
        script_lines.append("")

    script_lines.extend(
        [
            'echo "\u2713 Session ready"',
        ]
    )

    script_path = project_dir / ".claude" / "init-session.sh"
    script_path.parent.mkdir(parents=True, exist_ok=True)
    script_path.write_text("\n".join(script_lines))
    script_path.chmod(0o755)

    return script_path


def is_harness_initialized(project_dir: Path | None = None) -> bool:
    """Check if harness is initialized for this project."""
    if project_dir is None:
        project_dir = Path.cwd()
    return (project_dir / ".claude" / "harness.json").exists()
