#!/usr/bin/env python3
"""PreCompact hook: auto-save checkpoint before context compaction.

This hook fires when Claude's context is about to be compacted.
It immediately saves an auto-checkpoint with session state,
then prompts Claude to add a summary before compaction completes.

The checkpoint is written FIRST (synchronously) to guarantee
at least basic information is preserved even if Claude runs out of context.
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.checkpoint import (
    get_beads_current,
    get_modified_files,
    write_auto_checkpoint,
)
from lib.response import session_output


def main() -> None:
    """Handle PreCompact event."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # No input, still try to save checkpoint
        data = {}

    transcript = data.get("transcript", [])

    # 1. Get context
    beads_task = get_beads_current()
    modified_files = get_modified_files(transcript)

    # 2. Write checkpoint IMMEDIATELY (before returning control to Claude)
    checkpoint_path = write_auto_checkpoint(
        checkpoint_type="PreCompact",
        modified_files=modified_files,
        beads_task=beads_task,
    )

    # 3. Return prompt for Claude to enrich the checkpoint
    task_info = ""
    if beads_task:
        task_info = f"Task: {beads_task['id']} - {beads_task['title']}\n"

    files_info = ""
    if modified_files:
        files_list = ", ".join(f[0] for f in modified_files[:5])
        if len(modified_files) > 5:
            files_list += f" (+{len(modified_files) - 5} more)"
        files_info = f"Files: {files_list}\n"

    session_output(f"""## Context Compaction

Auto-checkpoint saved: `{checkpoint_path.name}`
{task_info}{files_info}
**Before compaction, add summary:**

```python
mcp__plugin_serena_serena__edit_memory(
    memory_file_name="{checkpoint_path.name}",
    needle="(Add summary here)",
    repl="<your summary of what was accomplished>",
    mode="literal"
)
```

Include:
- What was accomplished
- Current state / blockers
- Next steps
""")


if __name__ == "__main__":
    main()
