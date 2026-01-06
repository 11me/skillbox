#!/usr/bin/env python3
"""TDD Enforcement Stop Hook.

Analyzes session transcript to check if code was modified without running tests.
By default, issues a warning (advisory mode). Can be configured to block.

Config: .claude/tdd-enforcer.local.md
"""

import json
import os
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.notifier import notify  # noqa: E402
from lib.transcript_analyzer import analyze_transcript, read_transcript  # noqa: E402


def load_config() -> dict:
    """Load TDD enforcer configuration from .claude/tdd-enforcer.local.md."""
    config = {
        "enabled": True,
        "strictMode": False,
        "testCommand": None,
    }

    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    config_path = Path(project_dir) / ".claude" / "tdd-enforcer.local.md"

    if not config_path.exists():
        # TDD not explicitly enabled, check for auto-detect
        config["enabled"] = False
        return config

    try:
        content = config_path.read_text()
        # Simple YAML frontmatter parsing
        if "enabled: false" in content:
            config["enabled"] = False
        if "strictMode: true" in content:
            config["strictMode"] = True
    except IOError:
        pass

    return config


def main() -> None:
    """Main entry point for Stop hook."""
    try:
        # Read input from stdin (Claude Code hook protocol)
        try:
            _input_data = json.load(sys.stdin)
        except json.JSONDecodeError:
            pass

        # Load config
        config = load_config()

        # If TDD not enabled, allow stop silently
        if not config["enabled"]:
            print(json.dumps({}))
            sys.exit(0)

        # Get transcript path
        transcript_path = os.environ.get("TRANSCRIPT_PATH")
        if not transcript_path:
            print(json.dumps({}))
            sys.exit(0)

        # Read and analyze transcript
        transcript_content = read_transcript(transcript_path)
        if not transcript_content:
            print(json.dumps({}))
            sys.exit(0)

        # Analyze TDD compliance
        result = analyze_transcript(transcript_content)

        # Determine response
        if result.is_compliant:
            if result.tests_executed:
                notify("Claude Done", "TDD compliant - tests passed", urgency="low")
                output = {"systemMessage": f"✅ TDD Check: {result.message}"}
            else:
                output = {}
        else:
            # TDD violation detected
            modified_list = "\n".join(f"  - {f}" for f in result.modified_files[:5])
            if len(result.modified_files) > 5:
                modified_list += f"\n  ... and {len(result.modified_files) - 5} more"

            warning_msg = f"""## ⚠️ TDD Enforcement Warning

Code was modified but tests were not executed.

**Modified files:**
{modified_list}

**Recommended action:**
Run your test suite to verify changes.

**Commands:**
- Go: `go test ./...`
- TypeScript: `pnpm vitest run`
- Python: `pytest`
- Rust: `cargo test`

**TDD Reminder:** RED → GREEN → REFACTOR
"""

            if config["strictMode"]:
                notify("Claude Blocked", "Tests must be run (TDD strict mode)", urgency="critical")
                output = {
                    "decision": "deny",
                    "reason": "Tests must be run after code modifications (TDD strict mode)",
                    "systemMessage": warning_msg,
                }
            else:
                # Warning mode - allow but advise
                output = {"systemMessage": warning_msg}

        print(json.dumps(output))

    except Exception as e:
        # On any error, allow the operation with a warning
        error_output = {"systemMessage": f"TDD Enforcer error: {e!s}"}
        print(json.dumps(error_output))

    finally:
        sys.exit(0)


if __name__ == "__main__":
    main()
