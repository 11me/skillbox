#!/usr/bin/env python3
"""Notification hook: send desktop notification when Claude needs attention."""

import json
import logging
import os
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.notifier import notify  # noqa: E402

# Configure logging to stderr
logging.basicConfig(
    level=logging.DEBUG if os.environ.get("SKILLBOX_DEBUG") else logging.WARNING,
    format="[notification] %(levelname)s: %(message)s",
)
logger = logging.getLogger(__name__)


def main() -> None:
    """Handle Notification event."""
    # DEBUG: log to file to understand what happens in real hook call
    debug_file = Path("/tmp/notification-debug.log")
    try:
        from lib import tmux_state

        debug_info = {
            "XDG_RUNTIME_DIR": os.environ.get("XDG_RUNTIME_DIR"),
            "TMUX": os.environ.get("TMUX"),
            "state": tmux_state.load_state(),
            "current_pane": tmux_state.get_current_pane_id(),
            "get_window_id": tmux_state.get_window_id(),
        }
        debug_file.write_text(json.dumps(debug_info, indent=2))
    except Exception as e:
        debug_file.write_text(f"Debug error: {e}")

    try:
        data = json.load(sys.stdin)
        notification_type = data.get("notification_type", "")
        message = data.get("message", "Claude needs attention")

        # Config: notification_type -> (title, emoji)
        # ⏳ = needs user action, 💤 = idle/sleeping
        notification_config = {
            "permission_prompt": ("Permission Required", "⏳"),
            "idle_prompt": ("Claude Waiting", "💤"),
            "auth_success": ("Auth Success", None),
            "elicitation_dialog": ("Input Required", "⏳"),
        }

        title, emoji = notification_config.get(notification_type, ("Claude Notification", None))
        success = notify(title, message, urgency="normal", emoji=emoji)
        logger.debug(
            "Notification %s: type=%s, title=%s, emoji=%s",
            "sent" if success else "skipped",
            notification_type,
            title,
            emoji,
        )

    except json.JSONDecodeError as e:
        logger.warning("Failed to parse notification data: %s", e)
    except KeyError as e:
        logger.warning("Missing required field in notification data: %s", e)
    except Exception as e:
        logger.error("Unexpected error in notification hook: %s", e)

    # Always allow notification to proceed
    print(json.dumps({}))


if __name__ == "__main__":
    main()
