#!/usr/bin/env python3
"""Stop hook: ensure checkpoint exists before session end.

This hook fires when Claude is about to stop responding.
If no recent checkpoint exists (within 1 hour), it creates one
as a safety net for session continuity.

This complements precompact-checkpoint.py which handles context compaction.
Together they ensure session state is always preserved.
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.checkpoint import (
    find_recent_checkpoint,
    get_beads_current,
    get_modified_files,
    update_beads_comment,
    write_auto_checkpoint,
)
from lib.notifier import notify
from lib.response import allow


def main() -> None:
    """Handle Stop event."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        data = {}

    transcript = data.get("transcript", [])

    # 1. Check for recent checkpoint (within 1 hour)
    recent = find_recent_checkpoint(max_age_hours=1)

    if recent:
        # Recent checkpoint exists, just allow stop
        allow("Stop")
        return

    # 2. No recent checkpoint - create auto-checkpoint
    beads_task = get_beads_current()
    modified_files = get_modified_files(transcript)

    # Only create checkpoint if there was actual work done
    if not modified_files and not beads_task:
        # No work detected, skip checkpoint
        allow("Stop")
        return

    checkpoint_path = write_auto_checkpoint(
        checkpoint_type="SessionEnd",
        modified_files=modified_files,
        beads_task=beads_task,
    )

    # 3. Update beads task with session end note
    if beads_task:
        files_count = len(modified_files)
        message = (
            f"Session ended. {files_count} file(s) modified. Checkpoint: {checkpoint_path.name}"
        )
        update_beads_comment(beads_task["id"], message)

    # 4. Send notification
    notify(
        title="Checkpoint Saved",
        message=f"Auto-saved: {checkpoint_path.name}",
        urgency="low",
        emoji=None,  # Use default (✅)
    )

    allow("Stop")


if __name__ == "__main__":
    main()
