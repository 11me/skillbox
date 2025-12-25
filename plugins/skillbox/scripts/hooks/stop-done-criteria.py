#!/usr/bin/env python3
"""
Stop hook: blocks session end until validation is complete.
"""

import json
import sys
from pathlib import Path


def find_checkpoint_files() -> list[str]:
    """Find CHECKPOINT.md files in apps directory."""
    checkpoints = []
    apps_dir = Path("apps")
    if apps_dir.exists():
        for checkpoint in apps_dir.rglob("CHECKPOINT.md"):
            checkpoints.append(str(checkpoint))
    return checkpoints


def check_validation_ran() -> bool:
    """Check if helm-validate was likely run (heuristic)."""
    # This is a simple heuristic - in practice you might check
    # for specific markers or use session state
    return True  # Allow for now, can be made stricter


def main():
    data = json.load(sys.stdin)

    # Get session context
    transcript = data.get("transcript", [])

    # Check if /helm-validate was run in this session
    validate_ran = any("/helm-validate" in str(msg.get("content", "")) for msg in transcript)

    # Check if /checkpoint was run
    checkpoint_ran = any("/checkpoint" in str(msg.get("content", "")) for msg in transcript)

    # Check for existing checkpoint files
    checkpoints = find_checkpoint_files()

    warnings = []

    if not validate_ran:
        warnings.append("- /helm-validate was not run")

    if not checkpoint_ran and not checkpoints:
        warnings.append("- No checkpoint created (run /checkpoint)")

    if warnings:
        out = {
            "decision": "block",
            "reason": "Session completion criteria not met",
            "hookSpecificOutput": {
                "hookEventName": "Stop",
                "additionalContext": (
                    "Before ending session:\n"
                    + "\n".join(warnings)
                    + "\n\nRun the missing commands to complete the session."
                ),
            },
        }
        print(json.dumps(out))
    else:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "Stop"}}))


if __name__ == "__main__":
    main()
