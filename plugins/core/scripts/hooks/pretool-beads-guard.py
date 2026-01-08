#!/usr/bin/env python3
"""PreToolUse hook: blocks Write/Edit without active beads task.

Enforces task tracking discipline when workflow mode is active
(harness initialized OR .beads/ exists). This prevents agents from
making code changes without proper task context.

Activation criteria:
- .claude/harness.json exists (harness mode), OR
- .beads/ directory exists (beads workflow)

When active, blocks Write/Edit unless:
- There's an in_progress beads task, OR
- The file being modified is in allowed paths (docs, configs)
"""

import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from lib.bootstrap import is_harness_initialized
from lib.response import allow, block

# Paths that are always allowed (don't require active task)
ALLOWED_PATHS = [
    ".claude/",
    "CLAUDE.md",
    ".beads/",
    ".gitignore",
    "README.md",
    "CHANGELOG.md",
    "docs/",
    ".pre-commit-config.yaml",
    "plugin.json",
    ".claude-plugin/",
]


def is_workflow_active(project_dir: Path) -> bool:
    """Check if workflow mode is active (harness or beads)."""
    if is_harness_initialized(project_dir):
        return True
    if (project_dir / ".beads").is_dir():
        return True
    return False


def is_allowed_path(file_path: str) -> bool:
    """Check if file path is in allowed list."""
    for allowed in ALLOWED_PATHS:
        if allowed in file_path:
            return True
    return False


def has_active_beads_task() -> tuple[bool, str | None]:
    """Check if there's an in_progress beads task.

    Returns:
        tuple: (has_task: bool, error: str | None)
        - has_task: True if active task exists, False otherwise
        - error: Error message if beads command failed, None on success

    Note: This function implements FAIL-SAFE behavior - if we cannot verify
    task status, we return (False, error) to BLOCK operations rather than
    allowing them. This prevents bypassing task tracking on errors.
    """
    try:
        result = subprocess.run(
            ["bd", "list", "--status", "in_progress", "--json"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.returncode != 0:
            # Fail safe: cannot verify task status, block operation
            stderr = result.stderr.strip() if result.stderr else ""
            return False, f"bd command failed (exit {result.returncode}): {stderr}"

        tasks = json.loads(result.stdout)
        return len(tasks) > 0, None
    except subprocess.TimeoutExpired:
        return False, "bd command timed out (5s)"
    except FileNotFoundError:
        return False, "bd command not found - is beads installed?"
    except json.JSONDecodeError as e:
        return False, f"bd returned invalid JSON: {e}"
    except OSError as e:
        return False, f"bd command error: {e}"


def main() -> None:
    """Handle PreToolUse event for Write/Edit tools."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow("PreToolUse")
        return

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})

    # Only check Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        allow("PreToolUse")
        return

    cwd = Path.cwd()

    # Skip if workflow not active
    if not is_workflow_active(cwd):
        allow("PreToolUse")
        return

    file_path = tool_input.get("file_path", "")

    # Allow modifications to allowed paths
    if is_allowed_path(file_path):
        allow("PreToolUse")
        return

    # Check for active beads task
    has_task, error = has_active_beads_task()
    if has_task:
        allow("PreToolUse")
        return

    # Block: no active task or error verifying task status
    if error:
        # Beads command failed - block with error details
        block(
            reason=(
                f"Cannot verify beads task status: {error}. "
                "Fix beads or create task: "
                "bd create --title 'desc' -p 2 && bd update ID --status in_progress"
            ),
            event="PreToolUse",
            context=(
                "**The beads command failed.** Fix the issue or create a task manually:\n\n"
                f"**Error:** `{error}`\n\n"
                "**Possible fixes:**\n"
                "- Ensure beads is installed: `pip install beads` or `uv tool install beads`\n"
                "- Initialize beads: `bd init`\n"
                "- Check bd is in PATH\n\n"
                "**Or create a task:**\n"
                "```bash\n"
                'bd create --title "Description" -p 2\n'
                "bd update <id> --status in_progress\n"
                "```"
            ),
        )
    else:
        # No error, just no active task - include command in reason so Claude sees it
        block(
            reason=(
                "No active beads task. "
                "Run: bd create --title 'desc' -p 2 "
                "&& bd update <ID> --status in_progress"
            ),
            event="PreToolUse",
            context=(
                "Workflow mode is active. Create or start a task before modifying code:\n\n"
                "**Create new task:**\n"
                "```bash\n"
                'bd create --title "Description" -p 2\n'
                "bd update <id> --status in_progress\n"
                "```\n\n"
                "**Start existing task:**\n"
                "```bash\n"
                "bd ready  # List available tasks\n"
                "bd update <id> --status in_progress\n"
                "```\n\n"
                "**Priority levels:** 0=Critical, 1=High, 2=Medium (default), 3=Low, 4=Someday\n"
                "(Use `-p 0-4` or `-p P0-P4`, NOT words like 'high' or 'low')"
            ),
        )


if __name__ == "__main__":
    main()
