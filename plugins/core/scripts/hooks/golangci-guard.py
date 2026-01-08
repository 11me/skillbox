#!/usr/bin/env python3
"""PreToolUse hook: protects .golangci.yml files from unintended edits."""

import json
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from lib.response import allow, ask


def is_inside_plugin_dir(file_path: str) -> bool:
    """Check if file is inside the skillbox plugin directory.

    Uses CLAUDE_PLUGIN_ROOT env var provided by Claude Code.
    """
    plugin_root = os.environ.get("CLAUDE_PLUGIN_ROOT", "")
    if not plugin_root:
        return False

    file_abs = os.path.abspath(file_path)
    plugin_abs = os.path.abspath(plugin_root)

    return file_abs.startswith(plugin_abs + os.sep)


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow("PreToolUse")
        return

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})
    file_path = tool_input.get("file_path", "")

    if tool_name not in ("Write", "Edit"):
        allow("PreToolUse")
        return

    # Not a golangci config file
    if Path(file_path).name not in (".golangci.yml", ".golangci.yaml"):
        allow("PreToolUse")
        return

    # Allow editing files inside the skillbox plugin directory
    if is_inside_plugin_dir(file_path):
        allow("PreToolUse")
        return

    # For real project configs - ask for confirmation
    ask(
        reason="Modifying .golangci.yml — confirm this is intentional",
        event="PreToolUse",
        context=(
            "Linting config is a protected file.\nApprove if you explicitly requested this change."
        ),
    )


if __name__ == "__main__":
    main()
