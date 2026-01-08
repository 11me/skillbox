"""Shared utilities for Claude Code hooks."""

from lib.detector import detect_project_types
from lib.response import allow, ask, block, session_output

__all__ = [
    "detect_project_types",
    "session_output",
    "block",
    "ask",
    "allow",
]
