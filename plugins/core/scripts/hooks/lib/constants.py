"""Shared constants for emoji/notification hooks.

Single source of truth for emoji patterns and configuration.
"""

import re

# Emoji definitions - used in tmux window names and notifications
EMOJIS = {
    "working": "⏳",
    "attention": "🔴",
    "done": "✅",
    "idle": "💤",
    "permission": "🔐",
    "question": "❓",
}

# All emoji characters concatenated for pattern matching
EMOJI_CHARS = "".join(EMOJIS.values())  # "⏳🔴✅💤🔐❓"

# Python regex pattern for removing emoji prefix from window names
EMOJI_PREFIX_PATTERN = re.compile(rf"^[{EMOJI_CHARS}]\s*")

# State file TTL in seconds (24 hours)
# Files older than this will be cleaned up
STATE_FILE_TTL_SECONDS = 86400
