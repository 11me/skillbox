#!/usr/bin/env bash
#
# cancel-loop.sh - Cancel active discovery loop
#

set -euo pipefail

STATE_FILE=".claude/discovery-loop.local.md"
FINDINGS_FILE=".claude/discovery-findings.md"

if [[ -f "$STATE_FILE" ]]; then
    rm -f "$STATE_FILE"
    echo "Discovery loop cancelled."
    echo ""
    if [[ -f "$FINDINGS_FILE" ]]; then
        echo "Findings preserved in: $FINDINGS_FILE"
    fi
else
    echo "No active discovery loop found."
fi
