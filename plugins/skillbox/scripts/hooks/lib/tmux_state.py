"""Tmux state management for consistent pane/window targeting."""

import json
import os
import subprocess
import time
from pathlib import Path

from .constants import STATE_FILE_TTL_SECONDS

# State file location - use XDG or fallback to /tmp
_STATE_DIR = Path(os.environ.get("XDG_RUNTIME_DIR", "/tmp")) / "claude-skillbox"


def _get_state_file() -> Path:
    """Get state file path for current pane.

    Uses TMUX_PANE env var to create pane-specific state file.
    This ensures each Claude session has isolated state.
    """
    pane_id = os.environ.get("TMUX_PANE", "default")
    # Sanitize: %0 -> _0 for safe filename
    safe_pane_id = pane_id.replace("%", "_")
    return _STATE_DIR / f"tmux-state{safe_pane_id}.json"


def _run_tmux(args: list[str], timeout: int = 1) -> str | None:
    """Run tmux command and return stdout or None on failure."""
    if "TMUX" not in os.environ:
        return None
    try:
        result = subprocess.run(
            ["tmux", *args],
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        if result.returncode == 0:
            return result.stdout.strip()
    except (subprocess.TimeoutExpired, OSError):
        pass
    return None


def get_current_pane_id() -> str | None:
    """Get current pane ID from tmux (e.g., %5).

    Uses TMUX_PANE env var (reliable) with fallback to display-message.
    """
    # TMUX_PANE is set by tmux for shells - most reliable method
    if pane_id := os.environ.get("TMUX_PANE"):
        return pane_id
    # Fallback (may return wrong pane if user switched windows)
    return _run_tmux(["display-message", "-p", "#{pane_id}"])


def get_current_window_id() -> str | None:
    """Get current window ID from tmux (e.g., @2).

    Derives window ID from pane ID for accuracy.
    """
    # Get our pane ID first (reliable via TMUX_PANE)
    pane_id = get_current_pane_id()
    if pane_id:
        # Query window ID for our specific pane
        return _run_tmux(["display-message", "-t", pane_id, "-p", "#{window_id}"])
    # Fallback (may return wrong window if user switched)
    return _run_tmux(["display-message", "-p", "#{window_id}"])


def get_session_name() -> str | None:
    """Get current session name.

    Derives session name from pane ID for accuracy.
    """
    pane_id = get_current_pane_id()
    if pane_id:
        return _run_tmux(["display-message", "-t", pane_id, "-p", "#{session_name}"])
    return _run_tmux(["display-message", "-p", "#{session_name}"])


def save_state() -> bool:
    """Save current tmux pane/window IDs to state file.

    Should be called on SessionStart to capture the correct context.
    Returns True if state was saved successfully.
    """
    if "TMUX" not in os.environ:
        return False

    pane_id = get_current_pane_id()
    window_id = get_current_window_id()
    session_name = get_session_name()

    if not all([pane_id, window_id, session_name]):
        return False

    state = {
        "pane_id": pane_id,
        "window_id": window_id,
        "session_name": session_name,
        "pid": os.getpid(),
    }

    try:
        _STATE_DIR.mkdir(parents=True, exist_ok=True)
        _get_state_file().write_text(json.dumps(state))
        return True
    except OSError:
        return False


def load_state() -> dict | None:
    """Load saved tmux state with validation.

    Returns dict with pane_id, window_id, session_name or None if not available.
    Validates that loaded state belongs to current pane.
    """
    state_file = _get_state_file()
    try:
        if state_file.exists():
            state = json.loads(state_file.read_text())
            # Validate: pane_id must match current pane
            current_pane = os.environ.get("TMUX_PANE")
            if current_pane and state.get("pane_id") != current_pane:
                # State from another session - ignore
                return None
            return state
    except (OSError, json.JSONDecodeError):
        pass
    return None


def get_pane_id() -> str | None:
    """Get pane ID from saved state or current context.

    Prefers saved state (set at SessionStart) for consistency,
    falls back to current pane if state is unavailable.
    """
    state = load_state()
    if state and state.get("pane_id"):
        return state["pane_id"]
    return get_current_pane_id()


def get_window_id() -> str | None:
    """Get window ID from saved state or current context.

    Prefers saved state (set at SessionStart) for consistency,
    falls back to current window if state is unavailable.
    """
    state = load_state()
    if state and state.get("window_id"):
        return state["window_id"]
    return get_current_window_id()


def get_window_name(target: str | None = None) -> str | None:
    """Get window name, optionally for a specific target.

    Args:
        target: tmux target (pane_id, window_id, or None for saved/current)
    """
    if target is None:
        target = get_pane_id()

    if target:
        return _run_tmux(["display-message", "-t", target, "-p", "#{window_name}"])
    return _run_tmux(["display-message", "-p", "#{window_name}"])


def rename_window(new_name: str, target: str | None = None) -> bool:
    """Rename window with explicit target.

    Args:
        new_name: New window name
        target: tmux target (uses saved window_id if None)

    Returns:
        True if rename succeeded
    """
    if "TMUX" not in os.environ:
        return False

    if target is None:
        target = get_window_id()

    args = ["rename-window"]
    if target:
        args.extend(["-t", target])
    args.append(new_name)

    return _run_tmux(args) is not None


def get_context_string() -> str | None:
    """Get tmux context string for notification title.

    Returns string like '[session:0] window-name' or None if not in tmux.
    """
    if "TMUX" not in os.environ:
        return None

    state = load_state()
    if state:
        # Use saved state for consistency
        pane_id = state.get("pane_id")
        if pane_id:
            return _run_tmux(
                [
                    "display-message",
                    "-t",
                    pane_id,
                    "-p",
                    "[#{session_name}:#{window_index}] #{window_name}",
                ]
            )

    # Fallback to current context
    return _run_tmux(["display-message", "-p", "[#{session_name}:#{window_index}] #{window_name}"])


def clear_state() -> None:
    """Clear saved state file for current pane."""
    try:
        state_file = _get_state_file()
        if state_file.exists():
            state_file.unlink()
    except OSError:
        pass


def cleanup_stale_states() -> int:
    """Remove state files older than TTL.

    Cleans up orphaned state files from crashed or abandoned sessions.
    Called during SessionStart to prevent accumulation.

    Returns:
        Number of files cleaned up.
    """
    if not _STATE_DIR.exists():
        return 0

    cleaned = 0
    now = time.time()

    try:
        for state_file in _STATE_DIR.glob("tmux-state*.json"):
            try:
                mtime = state_file.stat().st_mtime
                if now - mtime > STATE_FILE_TTL_SECONDS:
                    state_file.unlink()
                    cleaned += 1
            except OSError:
                pass
    except OSError:
        pass

    return cleaned
