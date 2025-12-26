#!/usr/bin/env python3
"""PreToolUse hook: blocks writing secrets to values.yaml.

Detects hardcoded secrets in YAML values files and blocks the operation,
guiding users to use ExternalSecret instead.
"""

import json
import re
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.response import allow, block

SECRET_PATTERNS = [
    r"password\s*[:=]\s*['\"][^'\"]+['\"]",
    r"token\s*[:=]\s*['\"][^'\"]+['\"]",
    r"secret\s*[:=]\s*['\"][^'\"]+['\"]",
    r"api[_-]?key\s*[:=]\s*['\"][^'\"]+['\"]",
    r"access[_-]?key\s*[:=]\s*['\"][^'\"]+['\"]",
    r"private[_-]?key\s*[:=]\s*['\"][^'\"]+['\"]",
    r"credentials?\s*[:=]\s*['\"][^'\"]+['\"]",
    # AWS credentials
    r"aws_access_key_id\s*[:=]\s*['\"][^'\"]+['\"]",
    r"aws_secret_access_key\s*[:=]\s*['\"][^'\"]+['\"]",
    # Connection strings
    r"connection[_-]?string\s*[:=]\s*['\"][^'\"]+['\"]",
    # Bearer tokens
    r"bearer\s+[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+",
]


def looks_like_secret(content: str) -> bool:
    """Check if content appears to contain hardcoded secrets."""
    content_lower = content.lower()
    for pattern in SECRET_PATTERNS:
        if re.search(pattern, content_lower):
            return True
    return False


def main() -> None:
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

    file_path = tool_input.get("file_path", "")
    content = tool_input.get("content", "") or tool_input.get("new_string", "")

    # Only check values.yaml files
    if "values.yaml" not in file_path and "values.yml" not in file_path:
        allow("PreToolUse")
        return

    # Check for secrets
    if looks_like_secret(content):
        block(
            reason="Detected potential secret in values.yaml",
            event="PreToolUse",
            context=(
                "Do not hardcode secrets in values.yaml.\n"
                "Use ExternalSecret to fetch secrets from AWS Secrets Manager:\n"
                "1. Create ExternalSecret in overlay (apps/dev/<app>/secrets/)\n"
                "2. Reference with secrets.existingSecretName in values.yaml"
            ),
        )
    else:
        allow("PreToolUse")


if __name__ == "__main__":
    main()
