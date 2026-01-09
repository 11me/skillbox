#!/usr/bin/env python3
"""SessionStart hook: inject TDD guidelines when TDD mode is enabled.

TDD mode is detected via:
1. Explicit config in .claude/tdd-enforcer.local.md
2. Auto-detect by presence of test files
"""

import json
from pathlib import Path


def has_tests(cwd: Path) -> bool:
    """Check if project has tests.

    Note: Keep in sync with workflow/scripts/hooks/lib/detector.py:has_tests()
    """
    for test_dir in ["tests", "test", "__tests__", "spec"]:
        if (cwd / test_dir).is_dir():
            return True

    # Use recursive glob to find test files in subdirectories
    patterns = [
        "**/*_test.go",
        "**/*_test.py",
        "**/*.test.ts",
        "**/*.spec.ts",
        "**/test_*.py",
    ]
    for pattern in patterns:
        # Use next() with default to short-circuit on first match
        if next(cwd.glob(pattern), None) is not None:
            return True

    return False


def detect_tdd_mode(cwd: Path) -> dict[str, bool]:
    """Detect TDD mode status."""
    result = {"enabled": False, "strict": False}

    config_path = cwd / ".claude" / "tdd-enforcer.local.md"
    if config_path.exists():
        try:
            content = config_path.read_text(errors="ignore")
            if "enabled: false" in content:
                return result
            if "enabled: true" in content:
                result["enabled"] = True
            if "strictMode: true" in content:
                result["strict"] = True
            return result
        except OSError:
            pass

    if has_tests(cwd):
        result["enabled"] = True

    return result


def main() -> None:
    cwd = Path.cwd()
    output_lines: list[str] = []

    tdd_status = detect_tdd_mode(cwd)
    if not tdd_status["enabled"]:
        return

    # Try to load TDD-GUIDELINES.md from plugin
    guidelines_path = Path(__file__).parent.parent.parent / "skills/tdd-enforcer/TDD-GUIDELINES.md"

    mode_label = "STRICT" if tdd_status["strict"] else "ACTIVE"
    output_lines.append(f"## TDD Mode ({mode_label})")
    output_lines.append("")

    if guidelines_path.exists():
        guidelines = guidelines_path.read_text().strip()
        output_lines.append(guidelines)
    else:
        output_lines.append("**Cycle:** RED -> GREEN -> REFACTOR")
        output_lines.append("1. Write failing test FIRST")
        output_lines.append("2. Minimal implementation to pass")
        output_lines.append("3. Refactor with tests passing")

    output_lines.append("")

    if output_lines:
        print(
            json.dumps(
                {
                    "hookSpecificOutput": {
                        "hookEventName": "SessionStart",
                        "additionalContext": "\n".join(output_lines),
                    }
                }
            )
        )


if __name__ == "__main__":
    main()
