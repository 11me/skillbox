#!/usr/bin/env python3
"""
PreToolUse hook: blocks writing secrets to values.yaml.
"""

import json
import re
import sys

SECRET_PATTERNS = [
    r"password\s*[:=]\s*['\"][^'\"]+['\"]",
    r"token\s*[:=]\s*['\"][^'\"]+['\"]",
    r"secret\s*[:=]\s*['\"][^'\"]+['\"]",
    r"api[_-]?key\s*[:=]\s*['\"][^'\"]+['\"]",
    r"access[_-]?key\s*[:=]\s*['\"][^'\"]+['\"]",
    r"private[_-]?key\s*[:=]\s*['\"][^'\"]+['\"]",
    r"credentials?\s*[:=]\s*['\"][^'\"]+['\"]",
]


def looks_like_secret(content: str) -> bool:
    """Check if content appears to contain hardcoded secrets."""
    content_lower = content.lower()
    for pattern in SECRET_PATTERNS:
        if re.search(pattern, content_lower):
            return True
    return False


def main():
    data = json.load(sys.stdin)

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})

    # Only check Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "PreToolUse"}}))
        return

    file_path = tool_input.get("file_path", "")
    content = tool_input.get("content", "") or tool_input.get("new_string", "")

    # Only check values.yaml files
    if "values.yaml" not in file_path and "values.yml" not in file_path:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "PreToolUse"}}))
        return

    # Check for secrets
    if looks_like_secret(content):
        out = {
            "decision": "block",
            "reason": "Detected potential secret in values.yaml",
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "additionalContext": (
                    "Do not hardcode secrets in values.yaml.\n"
                    "Use ExternalSecret to fetch secrets from AWS Secrets Manager:\n"
                    "1. Create ExternalSecret in overlay (apps/dev/<app>/secrets/)\n"
                    "2. Reference with secrets.existingSecretName in values.yaml"
                ),
            },
        }
        print(json.dumps(out))
    else:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "PreToolUse"}}))


if __name__ == "__main__":
    main()
