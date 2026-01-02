"""Desktop notification utility using notify-send."""

import shutil
import subprocess


def notify(title: str, message: str, urgency: str = "normal") -> bool:
    """Send desktop notification via notify-send.

    Args:
        title: Notification title
        message: Notification body
        urgency: low, normal, critical

    Returns:
        True if notification sent successfully
    """
    if not shutil.which("notify-send"):
        return False

    try:
        subprocess.run(
            [
                "notify-send",
                "--urgency",
                urgency,
                "--app-name",
                "Claude Code",
                title,
                message,
            ],
            timeout=5,
            check=False,
        )
        return True
    except (subprocess.TimeoutExpired, OSError):
        return False
