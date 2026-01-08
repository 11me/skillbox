#!/usr/bin/env python3
"""PostToolUse hook: Check HelmRelease chart versions.

Analyzes written/edited files to detect HelmRelease manifests and
reminds to verify chart versions are current.
"""

import json
import re
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))


def is_helmrelease(content: str) -> bool:
    """Check if content is a HelmRelease manifest."""
    return bool(
        re.search(r"apiVersion:\s*helm\.toolkit\.fluxcd\.io/v2", content)
        and re.search(r"kind:\s*HelmRelease", content)
    )


def extract_chart_info(content: str) -> dict | None:
    """Extract chart name and version from HelmRelease."""
    chart_match = re.search(r"chart:\s*\n\s+spec:\s*\n\s+chart:\s*(\S+)", content)
    version_match = re.search(r"version:\s*[\"']?([^\"'\s]+)[\"']?", content)
    source_match = re.search(r"sourceRef:\s*\n\s+kind:\s*(\S+)", content)

    if not chart_match:
        return None

    return {
        "chart": chart_match.group(1),
        "version": version_match.group(1) if version_match else "unknown",
        "source_kind": source_match.group(1) if source_match else "unknown",
    }


def main() -> None:
    """Handle PostToolUse event."""
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # No valid input, allow silently
        print(json.dumps({}))
        sys.exit(0)

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {})

    # Only check Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        print(json.dumps({}))
        sys.exit(0)

    # Get file path from tool input
    file_path = tool_input.get("file_path", "")
    if not file_path:
        print(json.dumps({}))
        sys.exit(0)

    # Skip non-YAML files
    if not file_path.endswith((".yaml", ".yml")):
        print(json.dumps({}))
        sys.exit(0)

    # Try to read the file content
    try:
        content = Path(file_path).read_text()
    except (OSError, IOError):
        print(json.dumps({}))
        sys.exit(0)

    # Check if it's a HelmRelease
    if not is_helmrelease(content):
        print(json.dumps({}))
        sys.exit(0)

    # Extract chart info
    chart_info = extract_chart_info(content)
    if not chart_info:
        print(json.dumps({}))
        sys.exit(0)

    # Build reminder message
    if chart_info["source_kind"] == "HelmRepository":
        message = (
            f"HelmRelease detected: {chart_info['chart']} v{chart_info['version']}\n"
            f"Consider verifying this is the latest version via Context7."
        )
    else:
        message = f"HelmRelease detected: {chart_info['chart']} v{chart_info['version']}"

    # Output context for Claude
    output = {
        "hookSpecificOutput": {
            "hookEventName": "PostToolUse",
            "additionalContext": message,
        }
    }
    print(json.dumps(output))
    sys.exit(0)


if __name__ == "__main__":
    main()
