"""Hook response builders for Claude Code hooks."""

import json

from lib.notifier import notify


def session_output(message: str) -> None:
    """Output message for SessionStart hooks.

    Args:
        message: The message to output (supports markdown).
    """
    if message:
        print(
            json.dumps(
                {
                    "hookSpecificOutput": {
                        "hookEventName": "SessionStart",
                        "additionalContext": message,
                    }
                }
            )
        )


def block(reason: str, event: str = "PreToolUse", context: str | None = None) -> None:
    """Block action with reason.

    Args:
        reason: The reason for blocking.
        event: The hook event name (PreToolUse, Stop, SubagentStop, etc.).
        context: Additional context/guidance for the user.
    """
    # Send desktop notification
    notify("Claude Blocked", reason, urgency="critical")

    # Stop and SubagentStop hooks use different output format
    if event in ("Stop", "SubagentStop"):
        output: dict = {
            "decision": "block",
            "reason": reason,
        }
        if context:
            output["reason"] = f"{reason}\n\n{context}"
    else:
        # PreToolUse, PermissionRequest use hookSpecificOutput format
        output = {
            "hookSpecificOutput": {
                "hookEventName": event,
                "permissionDecision": "deny",
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
    # Send desktop notification
    notify("Claude Needs Input", reason, urgency="normal")

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
        # Stop and SubagentStop hooks don't support hookSpecificOutput
        # Just output empty JSON to allow
        if event in ("Stop", "SubagentStop"):
            print(json.dumps({}))
        else:
            print(json.dumps({"hookSpecificOutput": {"hookEventName": event}}))
