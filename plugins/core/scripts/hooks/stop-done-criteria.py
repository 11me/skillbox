#!/usr/bin/env python3
"""Stop hook: blocks session end until validation is complete.

Enforces that validation commands are run before ending session:
- For Helm/GitOps: /helm-validate and /checkpoint
- For Go: golangci-lint if Go files were modified
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.detector import detect_project_types
from lib.notifier import notify
from lib.response import allow, block


def find_checkpoint_files() -> list[str]:
    """Find CHECKPOINT.md files in apps directory."""
    checkpoints: list[str] = []
    apps_dir = Path("apps")
    if apps_dir.exists():
        for checkpoint in apps_dir.rglob("CHECKPOINT.md"):
            checkpoints.append(str(checkpoint))
    return checkpoints


def check_go_files_modified(transcript: list[dict]) -> bool:
    """Check if any .go files were written/edited in the session."""
    for msg in transcript:
        content = str(msg.get("content", ""))
        # Check for tool uses that modified .go files
        if ".go" in content and any(
            tool in content for tool in ["Write", "Edit", "created", "updated", "modified"]
        ):
            return True
    return False


def check_golangci_lint_ran(transcript: list[dict]) -> bool:
    """Check if golangci-lint was run in the session."""
    for msg in transcript:
        content = str(msg.get("content", ""))
        if "golangci-lint" in content and any(
            indicator in content for indicator in ["run", "passed", "no issues", "exit code 0"]
        ):
            return True
    return False


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow("Stop")
        return

    # Get session context
    transcript = data.get("transcript", [])
    cwd = Path.cwd()
    types = detect_project_types(cwd)

    warnings: list[str] = []

    # Helm/GitOps checks
    if types.get("helm") or types.get("gitops"):
        validate_ran = any("/helm-validate" in str(msg.get("content", "")) for msg in transcript)
        checkpoint_ran = any("/checkpoint" in str(msg.get("content", "")) for msg in transcript)
        checkpoints = find_checkpoint_files()

        if not validate_ran:
            warnings.append("- /helm-validate was not run")

        if not checkpoint_ran and not checkpoints:
            warnings.append("- No checkpoint created (run /checkpoint)")

    # Go project checks
    if types.get("go"):
        go_files_modified = check_go_files_modified(transcript)
        lint_ran = check_golangci_lint_ran(transcript)

        if go_files_modified and not lint_ran:
            warnings.append("- Go files were modified but golangci-lint was not run")
            warnings.append("  → Run: golangci-lint run ./...")

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
        # Notify user that Claude finished
        notify("Claude Done", "Task completed", urgency="low")
        allow("Stop")


if __name__ == "__main__":
    main()
