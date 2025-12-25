#!/usr/bin/env python3
"""
PreToolUse hook: requires user confirmation for git push.
Deterministic - pure regex matching.
"""

import json
import re
import sys


def main():
    data = json.load(sys.stdin)

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})
    command = tool_input.get("command", "")

    if tool_name != "Bash":
        sys.exit(0)

    # Match git push (with optional flags/args)
    if re.search(r"\bgit\s+push\b", command):
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "ask",
                "permissionDecisionReason": f"Git push requires confirmation: {command}",
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    sys.exit(0)


if __name__ == "__main__":
    main()
