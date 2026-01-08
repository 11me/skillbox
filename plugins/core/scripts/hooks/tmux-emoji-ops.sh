#!/bin/bash
# Unified tmux emoji operations - ultra-fast bash implementation
#
# Usage:
#   tmux-emoji-ops.sh clear          # Remove emoji prefix from window name
#   tmux-emoji-ops.sh set <emoji>    # Set emoji prefix on window name
#
# Performance: ~22ms (vs ~77ms for Python equivalent)

set -euo pipefail

ACTION="${1:-}"
EMOJI="${2:-}"

# Exit if not in tmux
[[ -z "${TMUX:-}" ]] && echo '{}' && exit 0

# Emoji characters - SINGLE SOURCE OF TRUTH
# Must match EMOJI_CHARS in lib/constants.py
EMOJI_CHARS="⏳🔴✅💤🔐❓"

# Get pane-specific state file
PANE_ID="${TMUX_PANE:-default}"
SAFE_PANE_ID="${PANE_ID//%/_}"
STATE_DIR="${XDG_RUNTIME_DIR:-/tmp}/claude-skillbox"
STATE_FILE="$STATE_DIR/tmux-state$SAFE_PANE_ID.json"

# Extract pane_id from saved state using grep/sed (avoids Python overhead)
TARGET="$TMUX_PANE"
if [[ -f "$STATE_FILE" ]]; then
    SAVED_PANE=$(grep -o '"pane_id"[[:space:]]*:[[:space:]]*"[^"]*"' "$STATE_FILE" 2>/dev/null | sed 's/.*"\([^"]*\)"$/\1/')
    [[ -n "$SAVED_PANE" ]] && TARGET="$SAVED_PANE"
fi

# Get current window name
CURRENT_NAME=$(tmux display-message -t "${TARGET:-$TMUX_PANE}" -p '#{window_name}' 2>/dev/null)
[[ -z "$CURRENT_NAME" ]] && echo '{}' && exit 0

# Remove existing emoji prefix
CLEAN_NAME=$(printf '%s' "$CURRENT_NAME" | sed "s/^[$EMOJI_CHARS][[:space:]]*//")

case "$ACTION" in
    clear)
        if [[ "$CLEAN_NAME" != "$CURRENT_NAME" ]]; then
            tmux rename-window -t "${TARGET:-$TMUX_PANE}" "$CLEAN_NAME" 2>/dev/null
        fi
        ;;
    set)
        [[ -z "$EMOJI" ]] && echo '{"error": "emoji required"}' && exit 1
        NEW_NAME="$EMOJI $CLEAN_NAME"
        tmux rename-window -t "${TARGET:-$TMUX_PANE}" "$NEW_NAME" 2>/dev/null
        ;;
    *)
        echo "Usage: $0 {clear|set <emoji>}" >&2
        echo '{"error": "invalid action"}' && exit 1
        ;;
esac

echo '{}'
