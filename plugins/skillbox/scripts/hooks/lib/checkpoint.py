"""Auto-checkpoint utilities for hooks.

Provides functions to:
- Extract modified files from session transcript
- Get current beads task info
- Write auto-checkpoint files to serena memories
- Find recent checkpoints
"""

import re
import subprocess
import time
from datetime import datetime
from pathlib import Path


def get_beads_current() -> dict | None:
    """Get current beads task info.

    Returns:
        Dict with 'id', 'title', 'status' or None if no task/beads unavailable.
    """
    try:
        # Get current task ID
        result = subprocess.run(
            ["bd", "current", "-q"],
            capture_output=True,
            text=True,
            timeout=2,
        )
        if result.returncode != 0 or not result.stdout.strip():
            return None

        task_id = result.stdout.strip()

        # Get task details
        result = subprocess.run(
            ["bd", "show", task_id],
            capture_output=True,
            text=True,
            timeout=2,
        )
        if result.returncode != 0:
            return {"id": task_id, "title": "", "status": "unknown"}

        # Parse output (format: "ID: X\nTitle: Y\nStatus: Z\n...")
        output = result.stdout
        title_match = re.search(r"Title:\s*(.+)", output)
        status_match = re.search(r"Status:\s*(\w+)", output)

        return {
            "id": task_id,
            "title": title_match.group(1).strip() if title_match else "",
            "status": status_match.group(1).strip() if status_match else "unknown",
        }

    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return None


def get_modified_files(transcript: list[dict]) -> list[tuple[str, str]]:
    """Extract modified files from session transcript.

    Args:
        transcript: List of message dicts from hook input.

    Returns:
        List of (file_path, action) tuples where action is 'created' or 'modified'.
    """
    modified: dict[str, str] = {}  # path -> action

    # Patterns to detect file modifications
    # Tool use patterns: Write, Edit tools
    write_pattern = re.compile(r'Write.*?file_path["\s:]+([^\s"]+)')
    edit_pattern = re.compile(r'Edit.*?file_path["\s:]+([^\s"]+)')
    created_pattern = re.compile(r"(?:created|wrote|Created)\s+[`'\"]?([^\s`'\"]+)[`'\"]?")
    modified_pattern = re.compile(r"(?:modified|updated|edited)\s+[`'\"]?([^\s`'\"]+)[`'\"]?")

    for msg in transcript:
        content = str(msg.get("content", ""))

        # Check for Write tool
        for match in write_pattern.finditer(content):
            path = match.group(1)
            if path not in modified:
                modified[path] = "created"

        # Check for Edit tool
        for match in edit_pattern.finditer(content):
            path = match.group(1)
            if path not in modified:
                modified[path] = "modified"

        # Check natural language patterns
        for match in created_pattern.finditer(content):
            path = match.group(1)
            if _is_likely_file_path(path) and path not in modified:
                modified[path] = "created"

        for match in modified_pattern.finditer(content):
            path = match.group(1)
            if _is_likely_file_path(path) and path not in modified:
                modified[path] = "modified"

    return [(path, action) for path, action in modified.items()]


def _is_likely_file_path(s: str) -> bool:
    """Check if string looks like a file path."""
    # Must have extension or be in a directory
    if "." not in s and "/" not in s:
        return False
    # Filter out common false positives
    if s.startswith("http") or s.startswith("git@"):
        return False
    if len(s) > 200:  # Too long
        return False
    return True


def write_auto_checkpoint(
    checkpoint_type: str,
    modified_files: list[tuple[str, str]],
    beads_task: dict | None,
) -> Path:
    """Write auto-checkpoint to serena memories.

    Args:
        checkpoint_type: "PreCompact" or "SessionEnd"
        modified_files: List of (path, action) tuples
        beads_task: Current beads task info or None

    Returns:
        Path to created checkpoint file.
    """
    memories_dir = Path.cwd() / ".serena" / "memories"
    memories_dir.mkdir(parents=True, exist_ok=True)

    timestamp = datetime.now().strftime("%Y-%m-%d-%H%M")
    filename = f"auto-checkpoint-{timestamp}.md"

    # Build content
    lines = [
        "# Auto Checkpoint",
        "",
        f"**Date:** {datetime.now().strftime('%Y-%m-%d %H:%M')}",
        f"**Type:** {checkpoint_type}",
        "",
    ]

    # Active task section
    if beads_task:
        lines.extend(
            [
                "## Active Task",
                f"{beads_task['id']}: {beads_task['title']}",
                f"Status: {beads_task['status']}",
                "",
            ]
        )

    # Modified files section
    lines.append("## Modified Files")
    if modified_files:
        for path, action in modified_files:
            lines.append(f"- {path} ({action})")
    else:
        lines.append("- (no files detected)")
    lines.append("")

    # Placeholder sections for Claude to fill
    lines.extend(
        [
            "## Session Summary",
            "(Add summary here)",
            "",
            "## Next Steps",
            "(Add next steps here)",
            "",
        ]
    )

    content = "\n".join(lines)
    checkpoint_path = memories_dir / filename
    checkpoint_path.write_text(content)

    return checkpoint_path


def find_recent_checkpoint(max_age_hours: int = 1) -> Path | None:
    """Find checkpoint created within max_age_hours.

    Looks for both auto-checkpoints and manual checkpoints.

    Args:
        max_age_hours: Maximum age in hours.

    Returns:
        Path to recent checkpoint or None.
    """
    memories_dir = Path.cwd() / ".serena" / "memories"
    if not memories_dir.exists():
        return None

    now = time.time()
    max_age_seconds = max_age_hours * 3600

    # Check auto-checkpoints first
    for pattern in ["auto-checkpoint-*.md", "checkpoint-*.md"]:
        for checkpoint in memories_dir.glob(pattern):
            try:
                mtime = checkpoint.stat().st_mtime
                if now - mtime < max_age_seconds:
                    return checkpoint
            except OSError:
                continue

    return None


def update_beads_comment(task_id: str, message: str) -> bool:
    """Add comment to beads task.

    Args:
        task_id: Beads task ID.
        message: Comment message.

    Returns:
        True if comment was added successfully.
    """
    try:
        result = subprocess.run(
            ["bd", "comments", "add", task_id, message],
            capture_output=True,
            timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False
