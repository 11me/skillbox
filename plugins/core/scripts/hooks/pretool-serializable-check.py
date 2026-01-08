#!/usr/bin/env python3
"""PreToolUse hook: Check Go serializable transactions for Serialize() call.

Only activates for Go files with pgx.Serializable usage.
Warns if Serializable is used without advisory lock pattern.
"""

import json
import re
import sys


def main() -> None:
    """Handle PreToolUse event for Write/Edit commands."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        sys.exit(0)

    tool_input = data.get("tool_input", {})
    file_path = tool_input.get("file_path", "")

    # Only check Go files
    if not file_path.endswith(".go"):
        sys.exit(0)

    # Get content being written
    # For Write: content is in tool_input
    # For Edit: we need to check the file after edit (use new_string)
    content = tool_input.get("content", "")
    new_string = tool_input.get("new_string", "")

    # Combine for checking
    check_content = content + new_string

    # Check for serializable transaction patterns
    serializable_patterns = [
        r"pgx\.Serializable",
        r"ExecSerializable",
        r"WithTx.*Serializable",
        r"sql\.LevelSerializable",
    ]

    has_serializable = any(re.search(pattern, check_content) for pattern in serializable_patterns)

    if not has_serializable:
        # No serializable usage - exit silently
        sys.exit(0)

    # Check if Serialize() is present
    if re.search(r"\.Serialize\s*\(", check_content):
        # Advisory lock pattern is used - exit silently
        sys.exit(0)

    # Serializable without Serialize() - warn
    output = {
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "additionalContext": (
                "⚠️ Serializable transaction detected without Serialize() call.\n"
                "Consider using advisory lock to prevent serialization conflicts.\n"
                'Pattern: repo.Serialize(ctx, "OperationName:resourceID")\n'
                "See: advisory-lock-pattern.md"
            ),
        }
    }
    print(json.dumps(output))
    sys.exit(0)


if __name__ == "__main__":
    main()
