#!/usr/bin/env python3
"""PreToolUse hook: Check go get version specifications.

Only activates for 'go get' commands:
- No version → suggest adding @latest
- @latest → allow silently
- Specific version → remind to verify via Context7
"""

import json
import sys


def main() -> None:
    """Handle PreToolUse event for Bash commands."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # No valid input, allow silently
        sys.exit(0)

    tool_input = data.get("tool_input", {})
    command = tool_input.get("command", "")

    # Only check 'go get' commands
    if "go get " not in command:
        # Not a go get command - exit silently (allow)
        sys.exit(0)

    # Check for @latest
    if "@latest" in command:
        # Good practice - allow silently
        sys.exit(0)

    # Check for specific version (@v...)
    if "@v" in command or "@" in command:
        # Specific version - remind to verify
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "additionalContext": (
                    "Specific version detected in go get. "
                    "Consider verifying this is the latest version via Context7."
                ),
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    # No version specified - suggest @latest
    output = {
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "additionalContext": (
                "No version specified in go get. "
                "Consider adding @latest for explicit latest version."
            ),
        }
    }
    print(json.dumps(output))
    sys.exit(0)


if __name__ == "__main__":
    main()
