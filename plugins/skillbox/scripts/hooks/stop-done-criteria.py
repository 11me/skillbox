#!/usr/bin/env python3
"""Stop hook: blocks session end until validation is complete.

Enforces that /helm-validate and /checkpoint are run before ending session.
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.response import allow, block


def find_checkpoint_files() -> list[str]:
    """Find CHECKPOINT.md files in apps directory."""
    checkpoints: list[str] = []
    apps_dir = Path("apps")
    if apps_dir.exists():
        for checkpoint in apps_dir.rglob("CHECKPOINT.md"):
            checkpoints.append(str(checkpoint))
    return checkpoints


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow("Stop")
        return

    # Get session context
    transcript = data.get("transcript", [])

    # Check if /helm-validate was run in this session
    validate_ran = any("/helm-validate" in str(msg.get("content", "")) for msg in transcript)

    # Check if /checkpoint was run
    checkpoint_ran = any("/checkpoint" in str(msg.get("content", "")) for msg in transcript)

    # Check for existing checkpoint files
    checkpoints = find_checkpoint_files()

    warnings: list[str] = []

    if not validate_ran:
        warnings.append("- /helm-validate was not run")

    if not checkpoint_ran and not checkpoints:
        warnings.append("- No checkpoint created (run /checkpoint)")

    if warnings:
        block(
            reason="Session completion criteria not met",
            event="Stop",
            context=(
                "Before ending session:\n"
                + "\n".join(warnings)
                + "\n\nRun the missing commands to complete the session."
            ),
        )
    else:
        allow("Stop")


if __name__ == "__main__":
    main()
