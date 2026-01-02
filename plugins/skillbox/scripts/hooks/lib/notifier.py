"""Desktop notification utility using notify-send."""

import os
import re
import shutil
import subprocess
from pathlib import Path


def _is_enabled() -> bool:
    """Check if notifications are enabled in .claude/skillbox.local.md."""
    config_path = Path.cwd() / ".claude" / "skillbox.local.md"
    if not config_path.exists():
        return True  # enabled by default

    try:
        content = config_path.read_text()
        # Parse YAML frontmatter
        match = re.match(r"^---\s*\n(.*?)\n---", content, re.DOTALL)
        if match:
            frontmatter = match.group(1)
            # Look for notifications: false
            if re.search(r"^\s*notifications:\s*false\s*$", frontmatter, re.MULTILINE):
                return False
        return True
    except OSError:
        return True


def _get_tmux_context() -> str | None:
    """Get tmux session:window context if running in tmux."""
    if "TMUX" not in os.environ:
        return None
    try:
        result = subprocess.run(
            [
                "tmux",
                "display-message",
                "-p",
                "[#{session_name}:#{window_index}] #{window_name}",
            ],
            capture_output=True,
            text=True,
            timeout=1,
        )
        if result.returncode == 0:
            return result.stdout.strip()
    except (subprocess.TimeoutExpired, OSError):
        pass
    return None


def _get_beads_task() -> str | None:
    """Get current beads task if available."""
    try:
        result = subprocess.run(
            ["bd", "current", "-q"],
            capture_output=True,
            text=True,
            timeout=2,
        )
        if result.returncode == 0 and result.stdout.strip():
            return f"#{result.stdout.strip()}"
    except (subprocess.TimeoutExpired, OSError, FileNotFoundError):
        pass
    return None


def notify(title: str, message: str, urgency: str = "normal") -> bool:
    """Send desktop notification via notify-send.

    Args:
        title: Notification title
        message: Notification body
        urgency: low, normal, critical

    Returns:
        True if notification sent successfully
    """
    if not _is_enabled():
        return False

    if not shutil.which("notify-send"):
        return False

    # Build enriched title with context
    context_parts = [title]
    if tmux := _get_tmux_context():
        context_parts.append(tmux)
    if task := _get_beads_task():
        context_parts.append(task)
    enriched_title = " ".join(context_parts)

    # Expire time based on urgency (ms): critical=persistent, normal=10s, low=5s
    expire_map = {"critical": 0, "normal": 10000, "low": 5000}
    expire_time = expire_map.get(urgency, 10000)

    try:
        subprocess.run(
            [
                "notify-send",
                "--urgency",
                urgency,
                "--expire-time",
                str(expire_time),
                "--app-name",
                "Claude Code",
                enriched_title,
                message,
            ],
            timeout=5,
            check=False,
        )
        return True
    except (subprocess.TimeoutExpired, OSError):
        return False
