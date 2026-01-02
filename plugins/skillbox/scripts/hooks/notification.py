#!/usr/bin/env python3
"""Notification hook: send desktop notification when Claude needs attention."""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.notifier import notify  # noqa: E402


def main() -> None:
    """Handle Notification event."""
    try:
        data = json.load(sys.stdin)
        notification_type = data.get("notification_type", "")
        message = data.get("message", "Claude needs attention")

        title_map = {
            "permission_prompt": "Permission Required",
            "idle_prompt": "Claude Waiting",
            "auth_success": "Auth Success",
            "elicitation_dialog": "Input Required",
        }

        title = title_map.get(notification_type, "Claude Notification")
        notify(title, message, urgency="normal")

    except (json.JSONDecodeError, KeyError):
        pass  # Silent fail

    # Always allow notification to proceed
    print(json.dumps({}))


if __name__ == "__main__":
    main()
