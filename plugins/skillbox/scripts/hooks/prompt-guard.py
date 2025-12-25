#!/usr/bin/env python3
"""
UserPromptSubmit hook: blocks scaffold commands without required parameters.
"""

import json
import re
import sys


def main():
    data = json.load(sys.stdin)
    prompt = (data.get("prompt") or "").lower()

    # Check if this is a helm scaffold request
    if not ("helm" in prompt and ("scaffold" in prompt or "chart" in prompt or "create" in prompt)):
        # Not a scaffold request, allow
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "UserPromptSubmit"}}))
        return

    missing = []

    # Check for required parameters
    if not re.search(r"\bapp(name)?\b", prompt) and not re.search(r"\bname[=:]\s*\w+", prompt):
        missing.append("appName")

    if "namespace" not in prompt and not re.search(r"\b(dev|prod|stage|staging)\b", prompt):
        missing.append("namespace/env")

    if "image" not in prompt and "repository" not in prompt:
        missing.append("image repository")

    if "secret" not in prompt and "external" not in prompt:
        missing.append("secretPath (AWS Secrets Manager key)")

    if missing:
        out = {
            "decision": "block",
            "reason": f"Missing required inputs: {', '.join(missing)}",
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": (
                    "Please provide:\n"
                    "- appName: Application name (kebab-case)\n"
                    "- namespace/env: Target environment (dev, prod)\n"
                    "- image: Container image repository\n"
                    "- secretPath: AWS Secrets Manager path (e.g., project/dev/app)\n"
                    "- ingressHost: Ingress hostname (optional)"
                ),
            },
        }
        print(json.dumps(out))
    else:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "UserPromptSubmit"}}))


if __name__ == "__main__":
    main()
