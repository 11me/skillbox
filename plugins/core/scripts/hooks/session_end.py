#!/usr/bin/env python3
"""SessionEnd hook: cleanup when Claude Code session ends.

Handles:
- Clearing tmux window emoji
- Cleaning up tmux state file
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib import tmux_state  # noqa: E402
from lib.notifier import clear_tmux_window_emoji  # noqa: E402


def main() -> None:
    # Clear emoji from tmux window
    clear_tmux_window_emoji()

    # Clear saved tmux state (session is ending)
    tmux_state.clear_state()

    # SessionEnd hooks cannot block - just return empty response
    print(json.dumps({}))


if __name__ == "__main__":
    main()
