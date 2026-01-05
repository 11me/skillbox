#!/usr/bin/env python3
"""PreToolUse hook: validates beads command parameters.

Intercepts bd commands and validates:
- Priority format (must be 0-4 or P0-P4, not words)
- Status format (must use underscore, not hyphen)

This prevents common errors that cause bd commands to fail.
"""

import json
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from lib.response import allow, block

# Invalid priority patterns (words instead of numbers)
INVALID_PRIORITY_PATTERNS = [
    r"(?:--priority|-p)\s+(high|low|medium|critical|urgent|normal)",
]

# Invalid status patterns
INVALID_STATUS_PATTERNS = [
    r"--status\s+in-progress",  # Should be in_progress (underscore)
]


def main() -> None:
    """Handle PreToolUse event for Bash tool."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow("PreToolUse")
        return

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})

    if tool_name != "Bash":
        allow("PreToolUse")
        return

    command = tool_input.get("command", "")

    # Only check bd commands
    if not re.search(r"\bbd\s+", command):
        allow("PreToolUse")
        return

    # Check for invalid priority format
    for pattern in INVALID_PRIORITY_PATTERNS:
        match = re.search(pattern, command, re.IGNORECASE)
        if match:
            block(
                reason=f"Invalid beads priority format: '{match.group()}'",
                event="PreToolUse",
                context=(
                    "**Priority must be numeric (0-4 or P0-P4):**\n"
                    "- `-p 0` or `-p P0` = Critical\n"
                    "- `-p 1` or `-p P1` = High\n"
                    "- `-p 2` or `-p P2` = Medium (default)\n"
                    "- `-p 3` or `-p P3` = Low\n"
                    "- `-p 4` or `-p P4` = Someday\n\n"
                    f"Fix: Replace `{match.group()}` with `-p <number>`"
                ),
            )
            return

    # Check for invalid status format
    for pattern in INVALID_STATUS_PATTERNS:
        match = re.search(pattern, command, re.IGNORECASE)
        if match:
            block(
                reason=f"Invalid beads status format: '{match.group()}'",
                event="PreToolUse",
                context=(
                    "**Status uses underscore, not hyphen:**\n"
                    "- `--status in_progress` (correct)\n"
                    "- `--status in-progress` (wrong)\n\n"
                    "Valid statuses: open, in_progress, blocked, closed"
                ),
            )
            return

    allow("PreToolUse")


if __name__ == "__main__":
    main()
