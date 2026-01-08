#!/usr/bin/env python3
"""PreToolUse hook: Guard features.json from direct modification.

Enforces that features.json can only be modified through harness commands,
not direct Write/Edit operations. This prevents accidental modification
of the feature tracking state (key principle from Anthropic's article:
"model is less likely to inappropriately change or overwrite JSON files").

We enforce this at the hook level to ensure the guardrail is always active.
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.response import block


def main() -> None:
    """Handle PreToolUse event for Write/Edit tools."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # No input, allow
        return

    tool_input = data.get("tool_input", {})
    file_path = tool_input.get("file_path", "")

    # Check if targeting features.json in .claude directory
    if "features.json" in file_path and ".claude" in file_path:
        block(
            reason="Direct modification of features.json is not allowed",
            context=(
                "Use harness commands to update features:\n"
                "- `/harness-verify <feature-id>` — Run verification and update status\n"
                "- `/harness-update <feature-id> <status>` — Manual status update\n"
                "- `/harness-init` — Add new features\n"
                "\n"
                "This ensures proper verification tracking and beads integration."
            ),
        )
        return

    # Check if targeting harness.json (read-only state)
    if "harness.json" in file_path and ".claude" in file_path:
        block(
            reason="Direct modification of harness.json is not allowed",
            context=(
                "harness.json is managed automatically by the harness system.\n"
                "Session state is updated on each session start.\n"
                "\n"
                "To reinitialize: `/harness-init --force`"
            ),
        )


if __name__ == "__main__":
    main()
