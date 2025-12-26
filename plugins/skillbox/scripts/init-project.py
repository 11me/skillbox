#!/usr/bin/env python3
"""Initialize project for production workflow.

Creates:
- .beads/ (task tracking)
- .serena/ (code memory)
- CLAUDE.md (AI quick reference)
- .pre-commit-config.yaml (optional)
- tests/ directory (optional)
"""

import argparse
import subprocess
import sys
from pathlib import Path


def command_exists(cmd: str) -> bool:
    """Check if a command exists in PATH."""
    try:
        subprocess.run(
            ["which", cmd],
            capture_output=True,
            check=True,
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def init_beads(cwd: Path) -> None:
    """Initialize beads task tracking."""
    print("[1/5] Setting up beads...")

    if (cwd / ".beads").is_dir():
        print("  ✓ .beads/ already exists")
        return

    if not command_exists("bd"):
        print("  ⚠ beads CLI not found. Install: cargo install beads")
        return

    result = subprocess.run(["bd", "init"], capture_output=True, text=True)
    if result.returncode == 0:
        print("  ✓ beads initialized")
        # Run doctor --fix silently
        subprocess.run(["bd", "doctor", "--fix"], capture_output=True)
    else:
        print(f"  ⚠ beads init had issues: {result.stderr.strip()}")


def init_serena(cwd: Path, name: str) -> None:
    """Initialize serena code memory."""
    print("[2/5] Setting up serena...")

    serena_dir = cwd / ".serena"
    serena_dir.mkdir(exist_ok=True)
    (serena_dir / "memories").mkdir(exist_ok=True)

    # project.yml
    project_yml = serena_dir / "project.yml"
    if not project_yml.exists():
        project_yml.write_text(f"""version: "1.0"
name: "{name}"
description: "Project initialized with skillbox"
""")
        print("  ✓ .serena/project.yml created")
    else:
        print("  ✓ .serena/project.yml already exists")

    # overview.md
    overview = serena_dir / "memories" / "overview.md"
    if not overview.exists():
        # Get directory listing for structure
        try:
            files = sorted([f.name for f in cwd.iterdir() if not f.name.startswith(".")])[:15]
            structure = "\n".join(files) if files else "[empty]"
        except OSError:
            structure = "[unable to list]"

        overview.write_text(f"""# {name} Overview

## Structure
```
{structure}
```

## Key Files
- README.md — documentation
- CLAUDE.md — AI quick reference

## Patterns
[Add patterns used in this project]

## Notes
[Add notes for future sessions]
""")
        print("  ✓ .serena/memories/overview.md created")
    else:
        print("  ✓ overview.md already exists")


def detect_stack(cwd: Path) -> str:
    """Detect project stack and return markdown section."""
    if (cwd / "package.json").exists():
        return """## Stack
- Node.js / TypeScript
- Run: `npm install && npm run dev`
- Test: `npm test`

"""
    if (cwd / "pyproject.toml").exists() or (cwd / "requirements.txt").exists():
        return """## Stack
- Python
- Run: `uv sync && uv run python main.py`
- Test: `pytest`

"""
    if (cwd / "go.mod").exists():
        return """## Stack
- Go
- Run: `go run .`
- Test: `go test ./...`

"""
    if (cwd / "Cargo.toml").exists():
        return """## Stack
- Rust
- Run: `cargo run`
- Test: `cargo test`

"""
    return ""


def init_claude_md(cwd: Path, name: str) -> None:
    """Create CLAUDE.md with language-aware content."""
    print("[3/5] Creating CLAUDE.md...")

    claude_md = cwd / "CLAUDE.md"
    if claude_md.exists():
        print("  ✓ CLAUDE.md already exists")
        return

    stack_section = detect_stack(cwd)

    claude_md.write_text(f"""# {name}

> **New session?** Run: `bd prime` · `bd ready`

## Key Commands

| Command | Purpose |
|---------|---------|
| `bd ready` | Available tasks |
| `bd update <id> --status in_progress` | Start task |
| `bd close <id>` | Complete task |
| `/commit` | Create commit |

{stack_section}## Architecture

[Describe key components]

## Rules

- Update README.md for user-facing changes
- Update CLAUDE.md for workflow changes
""")
    print("  ✓ CLAUDE.md created")


def init_precommit(cwd: Path) -> None:
    """Set up pre-commit hooks."""
    print("[4/5] Setting up pre-commit...")

    config = cwd / ".pre-commit-config.yaml"
    if config.exists():
        print("  ✓ .pre-commit-config.yaml already exists")
    else:
        config.write_text("""repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v6.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-added-large-files
      - id: detect-private-key

  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.30.0
    hooks:
      - id: gitleaks

  - repo: https://github.com/compilerla/conventional-pre-commit
    rev: v3.0.0
    hooks:
      - id: conventional-pre-commit
        stages: [commit-msg]
""")
        print("  ✓ .pre-commit-config.yaml created")

    # Install hooks if pre-commit available
    if command_exists("pre-commit"):
        result = subprocess.run(
            ["pre-commit", "install", "--install-hooks", "-t", "pre-commit", "-t", "commit-msg"],
            capture_output=True,
        )
        if result.returncode == 0:
            print("  ✓ pre-commit hooks installed")
        else:
            print("  ⚠ pre-commit install failed")
    else:
        print("  ⚠ pre-commit not found. Run: pip install pre-commit && pre-commit install")


def init_tests(cwd: Path) -> None:
    """Create tests directory based on stack."""
    print("[5/5] Checking tests directory...")

    # Check if tests exist
    for test_dir in ["tests", "test", "__tests__"]:
        if (cwd / test_dir).is_dir():
            print(f"  ✓ {test_dir}/ already exists")
            return

    # Check for Go/Rust inline tests
    if (cwd / "go.mod").exists():
        print("  ℹ Go uses *_test.go files (no separate dir needed)")
        return

    # Create appropriate tests dir
    if (cwd / "package.json").exists():
        (cwd / "__tests__").mkdir()
        print("  ✓ Created __tests__/ (Jest/Vitest convention)")
    elif (cwd / "pyproject.toml").exists() or (cwd / "requirements.txt").exists():
        tests_dir = cwd / "tests"
        tests_dir.mkdir()
        (tests_dir / "__init__.py").touch()
        print("  ✓ Created tests/ with __init__.py")
    elif (cwd / "Cargo.toml").exists():
        (cwd / "tests").mkdir()
        print("  ✓ Created tests/ for integration tests")
    else:
        (cwd / "tests").mkdir()
        print("  ✓ Created tests/")


def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Initialize project for production workflow")
    parser.add_argument("--name", help="Project name (default: directory name)")
    parser.add_argument(
        "--minimal",
        action="store_true",
        help="Only beads + CLAUDE.md (no serena, no pre-commit, no tests)",
    )
    parser.add_argument(
        "--skip-precommit",
        action="store_true",
        help="Skip pre-commit setup",
    )
    args = parser.parse_args()

    cwd = Path.cwd()
    name = args.name or cwd.name

    # Check git repo
    if not (cwd / ".git").is_dir():
        print("❌ Not a git repository. Run: git init")
        sys.exit(1)

    print(f"🚀 Initializing project: {name}\n")

    # Step 1: Beads (always)
    init_beads(cwd)

    # Step 2: Serena (unless minimal)
    if not args.minimal:
        init_serena(cwd, name)
    else:
        print("[2/5] Skipping serena (--minimal mode)")

    # Step 3: CLAUDE.md (always)
    init_claude_md(cwd, name)

    # Step 4: Pre-commit (unless minimal or skipped)
    if not args.minimal and not args.skip_precommit:
        init_precommit(cwd)
    else:
        print("[4/5] Skipping pre-commit")

    # Step 5: Tests (unless minimal)
    if not args.minimal:
        init_tests(cwd)
    else:
        print("[5/5] Skipping tests directory (--minimal mode)")

    print("\n✅ Project initialized!")
    print("\nNext steps:")
    print("  1. Customize CLAUDE.md and .serena/memories/overview.md")
    print("  2. git add -A && git commit -m 'chore: init production workflow'")
    print("  3. bd create --title '...' -t task")


if __name__ == "__main__":
    main()
