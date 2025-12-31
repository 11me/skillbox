#!/usr/bin/env python3
"""PreToolUse hook: blocks editing .golangci.yml files."""

import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from lib.response import allow, block


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

    # Block .golangci.yml and .golangci.yaml
    if Path(file_path).name in (".golangci.yml", ".golangci.yaml"):
        block(
            reason="Cannot modify .golangci.yml — linting config is protected",
            event="PreToolUse",
            context=(
                "Linting configuration should not be modified by the agent.\n"
                "To change linting rules, edit the file manually.\n"
                "See: go-development skill → references/linting-pattern.md"
            ),
        )
    else:
        allow("PreToolUse")


if __name__ == "__main__":
    main()
