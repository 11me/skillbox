"""Desktop notification utility using notify-send."""

import logging
import os
import re
import shutil
import subprocess
from pathlib import Path

from . import tmux_state

# Configure logging to stderr (won't interfere with JSON output)
logging.basicConfig(
    level=logging.DEBUG if os.environ.get("SKILLBOX_DEBUG") else logging.WARNING,
    format="[notifier] %(levelname)s: %(message)s",
)
logger = logging.getLogger(__name__)

# Emoji pattern for window name cleanup
EMOJI_PREFIX_PATTERN = re.compile(r"^[🔴⏳✅💤]\s*")


def _find_config_file() -> Path | None:
    """Find skillbox.local.md config file.

    Searches in:
    1. CLAUDE_PROJECT_ROOT/.claude/skillbox.local.md
    2. Current working directory/.claude/skillbox.local.md
    3. Walk up directory tree looking for .claude/skillbox.local.md
    """
    # Try CLAUDE_PROJECT_ROOT first
    project_root = os.environ.get("CLAUDE_PROJECT_ROOT")
    if project_root:
        config = Path(project_root) / ".claude" / "skillbox.local.md"
        if config.exists():
            return config

    # Try current directory
    cwd = Path.cwd()
    config = cwd / ".claude" / "skillbox.local.md"
    if config.exists():
        return config

    # Walk up directory tree (max 10 levels)
    current = cwd
    for _ in range(10):
        config = current / ".claude" / "skillbox.local.md"
        if config.exists():
            return config
        parent = current.parent
        if parent == current:  # Reached root
            break
        current = parent

    return None


def _is_enabled() -> bool:
    """Check if notifications are enabled in config."""
    config_path = _find_config_file()
    if not config_path:
        logger.debug("No config file found, notifications enabled by default")
        return True

    try:
        content = config_path.read_text()
        # Parse YAML frontmatter
        match = re.match(r"^---\s*\n(.*?)\n---", content, re.DOTALL)
        if match:
            frontmatter = match.group(1)
            # Look for notifications: false
            if re.search(r"^\s*notifications:\s*false\s*$", frontmatter, re.MULTILINE):
                logger.debug("Notifications disabled in config")
                return False
        return True
    except OSError as e:
        logger.warning("Failed to read config: %s", e)
        return True


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


def _clean_window_name(name: str) -> str:
    """Remove emoji prefix from window name."""
    return EMOJI_PREFIX_PATTERN.sub("", name)


def _set_tmux_window_emoji(emoji: str) -> bool:
    """Add emoji prefix to tmux window name.

    Uses saved tmux state for consistent targeting across panes.

    Args:
        emoji: Emoji to prefix (🔴, ⏳, ✅)

    Returns:
        True if emoji was set successfully
    """
    if "TMUX" not in os.environ:
        return False

    # Get current window name using saved target
    current_name = tmux_state.get_window_name()
    if not current_name:
        logger.warning("Failed to get tmux window name")
        return False

    # Clean existing emoji and add new one
    clean_name = _clean_window_name(current_name)
    new_name = f"{emoji} {clean_name}"

    # Rename with explicit target
    if tmux_state.rename_window(new_name):
        logger.debug("Set window emoji: %s -> %s", current_name, new_name)
        return True
    else:
        logger.warning("Failed to rename tmux window")
        return False


def clear_tmux_window_emoji() -> bool:
    """Remove emoji prefix from tmux window name.

    Uses saved tmux state for consistent targeting.

    Returns:
        True if emoji was cleared successfully
    """
    if "TMUX" not in os.environ:
        return False

    current_name = tmux_state.get_window_name()
    if not current_name:
        return False

    clean_name = _clean_window_name(current_name)
    if clean_name != current_name:
        if tmux_state.rename_window(clean_name):
            logger.debug("Cleared window emoji: %s -> %s", current_name, clean_name)
            return True
        else:
            logger.warning("Failed to clear window emoji")
            return False

    return True  # No emoji to clear


def notify(title: str, message: str, urgency: str = "normal", emoji: str | None = None) -> bool:
    """Send desktop notification via notify-send.

    Args:
        title: Notification title
        message: Notification body
        urgency: low, normal, critical
        emoji: Optional explicit emoji to set on tmux window.
               If None, derives from urgency (🔴=critical, ⏳=normal, ✅=low).

    Returns:
        True if notification sent successfully
    """
    if not _is_enabled():
        return False

    if not shutil.which("notify-send"):
        logger.warning("notify-send not found in PATH")
        return False

    # Build enriched title with context
    context_parts = [title]
    if tmux_context := tmux_state.get_context_string():
        context_parts.append(tmux_context)
    if task := _get_beads_task():
        context_parts.append(task)
    enriched_title = " ".join(context_parts)

    # Expire time based on urgency (ms): critical=persistent, normal=10s, low=5s
    expire_map = {"critical": 0, "normal": 10000, "low": 5000}
    expire_time = expire_map.get(urgency, 10000)

    # Set emoji indicator on tmux window
    # Use explicit emoji if provided, otherwise derive from urgency
    if emoji is None:
        emoji_map = {"critical": "🔴", "normal": "⏳", "low": "✅"}
        emoji = emoji_map.get(urgency)

    if emoji:
        _set_tmux_window_emoji(emoji)

    try:
        result = subprocess.run(
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
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            logger.warning("notify-send failed: %s", result.stderr)
            return False
        logger.debug("Notification sent: %s", enriched_title)
        return True
    except subprocess.TimeoutExpired:
        logger.warning("notify-send timed out")
        return False
    except OSError as e:
        logger.warning("notify-send error: %s", e)
        return False
