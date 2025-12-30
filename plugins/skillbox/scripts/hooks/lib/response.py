"""Hook response builders for Claude Code hooks."""

import json


def session_output(message: str) -> None:
    """Output message for SessionStart hooks.

    Args:
        message: The message to output (supports markdown).
    """
    if message:
        print(json.dumps({"systemMessage": message}))


def block(reason: str, event: str = "PreToolUse", context: str | None = None) -> None:
    """Block action with reason.

    Args:
        reason: The reason for blocking.
        event: The hook event name.
        context: Additional context/guidance for the user.
    """
    output: dict = {
        "hookSpecificOutput": {
            "hookEventName": event,
            "permissionDecision": "block",
            "permissionDecisionReason": reason,
        }
    }
    if context:
        output["hookSpecificOutput"]["additionalContext"] = context
    print(json.dumps(output))


def ask(reason: str, event: str = "PreToolUse", context: str | None = None) -> None:
    """Ask for user permission.

    Args:
        reason: The reason for asking.
        event: The hook event name.
        context: Additional context/guidance for the user.
    """
    output: dict = {
        "hookSpecificOutput": {
            "hookEventName": event,
            "permissionDecision": "ask",
            "permissionDecisionReason": reason,
        }
    }
    if context:
        output["hookSpecificOutput"]["additionalContext"] = context
    print(json.dumps(output))


def allow(event: str | None = None) -> None:
    """Allow action.

    Args:
        event: Optional event name. If provided, outputs JSON with event.
               If None, exits silently (for SessionStart hooks).
    """
    if event:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": event}}))
