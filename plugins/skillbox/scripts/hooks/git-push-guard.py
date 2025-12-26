#!/usr/bin/env python3
"""PreToolUse hook: requires user confirmation for git push.

Deterministic - pure regex matching.
"""

import json
import re
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.response import allow, ask


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow("PreToolUse")
        return

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})
    command = tool_input.get("command", "")

    if tool_name != "Bash":
        allow("PreToolUse")
        return

    # Match git push (with optional flags/args)
    if re.search(r"\bgit\s+push\b", command):
        ask(
            reason=f"Git push requires confirmation: {command}",
            event="PreToolUse",
        )
    else:
        allow("PreToolUse")


if __name__ == "__main__":
    main()
