#!/usr/bin/env python3
"""PreToolUse hook: Enforce Context7 usage for HelmRelease versions.

Checks:
1. Empty version: "" → Block and require Context7
2. Hardcoded version: v1.2.3 → Remind to verify via Context7
"""

import json
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent / "lib"))
from response import ask


def main() -> None:
    data = json.load(sys.stdin)
    tool_input = data.get("tool_input", {})

    # Only check Write/Edit to yaml files
    file_path = tool_input.get("file_path", "")
    if not file_path.endswith((".yaml", ".yml")):
        return  # Allow silently

    content = tool_input.get("content", "") or tool_input.get("new_string", "")
    if not content:
        return  # Allow silently

    # Only check HelmRelease files
    if "kind: HelmRelease" not in content and "helmrelease" not in file_path.lower():
        return  # Allow silently

    # Extract chart name for better error messages
    chart_match = re.search(r"chart:\s*(\S+)", content)
    chart_name = chart_match.group(1) if chart_match else "{chart-name}"

    # Check for version field in spec.chart.spec
    version_match = re.search(r'version:\s*["\']?([^"\'\s\n#]*)["\']?', content)

    if version_match:
        version = version_match.group(1).strip()

        # Case 1: Empty version - require Context7
        if not version:
            return ask(
                reason=(
                    "HelmRelease has empty version. Use Context7 first:\n\n"
                    f'1. resolve-library-id: libraryName="{chart_name}"\n'
                    '2. query-docs: topic="helm chart version"\n'
                    "3. Set version from documentation\n\n"
                    "This ensures you're using the current stable version."
                ),
                event="PreToolUse",
            )

        # Case 2: Version looks like hardcoded (v1.2.3 or 1.2.3)
        if re.match(r"^v?\d+\.\d+", version):
            return ask(
                reason=(
                    f"HelmRelease version '{version}' detected.\n\n"
                    "VERIFY this version was obtained from Context7:\n"
                    f'1. resolve-library-id: libraryName="{chart_name}"\n'
                    '2. query-docs: topic="helm chart latest version"\n\n'
                    "If version is from Context7 → proceed.\n"
                    "If version is hardcoded/copied → fetch from Context7 first."
                ),
                event="PreToolUse",
            )

    # Allow all other cases silently


if __name__ == "__main__":
    main()
