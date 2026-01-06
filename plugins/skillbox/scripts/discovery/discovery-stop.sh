#!/usr/bin/env bash
#
# discovery-stop.sh - Stop hook for discovery loop
# Based on Ralph Wiggum pattern
#
# This hook intercepts session exit and continues the discovery loop
# until the completion signal is found or max iterations reached.
#
# Requirements: jq, perl (for regex extraction)
#

set -euo pipefail

STATE_FILE=".claude/discovery-loop.local.md"
FINDINGS_FILE=".claude/discovery-findings.md"

# If no state file, allow normal exit
if [[ ! -f "$STATE_FILE" ]]; then
    exit 0
fi

# Read state from frontmatter
get_frontmatter_value() {
    local key="$1"
    grep "^${key}:" "$STATE_FILE" | head -1 | sed "s/^${key}: *//" | tr -d '"'
}

ACTIVE=$(get_frontmatter_value "active")
ITERATION=$(get_frontmatter_value "iteration")
MAX_ITERATIONS=$(get_frontmatter_value "max_iterations")
COMPLETION_SIGNAL=$(get_frontmatter_value "completion_signal")

# If not active, allow normal exit
if [[ "$ACTIVE" != "true" ]]; then
    exit 0
fi

# Validate iteration is numeric
if ! [[ "$ITERATION" =~ ^[0-9]+$ ]]; then
    echo "Warning: Invalid iteration value in state file: $ITERATION" >&2
    rm -f "$STATE_FILE"
    exit 0
fi

# Validate max_iterations is numeric
if ! [[ "$MAX_ITERATIONS" =~ ^[0-9]+$ ]]; then
    MAX_ITERATIONS=10
fi

# Read transcript from stdin (JSON with transcript_path)
INPUT=$(cat)
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path // empty' 2>/dev/null || true)

# Check for completion signal in last assistant message
if [[ -n "$TRANSCRIPT_PATH" ]] && [[ -f "$TRANSCRIPT_PATH" ]]; then
    # Get last assistant message from JSONL transcript
    LAST_ASSISTANT=$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" | tail -1 || true)

    if [[ -n "$LAST_ASSISTANT" ]]; then
        # Extract text content
        ASSISTANT_TEXT=$(echo "$LAST_ASSISTANT" | jq -r '.content[]? | select(.type == "text") | .text' 2>/dev/null | tr '\n' ' ' || true)

        # Check for completion signal (extract text between <discovered> tags)
        if echo "$ASSISTANT_TEXT" | grep -q '<discovered>.*</discovered>'; then
            FOUND_SIGNAL=$(echo "$ASSISTANT_TEXT" | perl -0777 -pe 's/.*?<discovered>(.*?)<\/discovered>.*/$1/s' 2>/dev/null || true)

            # Extract expected signal text (without tags)
            EXPECTED_SIGNAL=$(echo "$COMPLETION_SIGNAL" | sed 's/<discovered>//' | sed 's/<\/discovered>//')

            # Compare (literal match)
            if [[ "$FOUND_SIGNAL" = "$EXPECTED_SIGNAL" ]]; then
                # Discovery complete! Clean up and allow exit
                rm -f "$STATE_FILE"

                echo ""
                echo "Discovery loop complete after $ITERATION iteration(s)."
                echo "Findings saved to: $FINDINGS_FILE"
                exit 0
            fi
        fi
    fi
fi

# Check iteration limit
if [[ "$MAX_ITERATIONS" -gt 0 ]] && [[ "$ITERATION" -ge "$MAX_ITERATIONS" ]]; then
    rm -f "$STATE_FILE"

    echo ""
    echo "Discovery loop reached maximum iterations ($MAX_ITERATIONS)."
    echo "Findings saved to: $FINDINGS_FILE"
    echo ""
    echo "Consider:"
    echo "  - Reviewing findings and narrowing the research question"
    echo "  - Restarting with /discover-loop and a more focused question"
    exit 0
fi

# Continue loop: increment iteration and re-inject prompt
NEXT_ITERATION=$((ITERATION + 1))

# Update iteration in state file (atomic)
TEMP_FILE="${STATE_FILE}.tmp.$$"
sed "s/^iteration: .*/iteration: $NEXT_ITERATION/" "$STATE_FILE" > "$TEMP_FILE"
mv "$TEMP_FILE" "$STATE_FILE"

# Also update the "Current Iteration" section if present
if grep -q "This is iteration" "$STATE_FILE"; then
    TEMP_FILE="${STATE_FILE}.tmp.$$"
    sed "s/This is iteration [0-9]* of/This is iteration $NEXT_ITERATION of/" "$STATE_FILE" > "$TEMP_FILE"
    mv "$TEMP_FILE" "$STATE_FILE"
fi

# Extract research question from state file (everything after the frontmatter)
RESEARCH_QUESTION=$(awk '/^---$/{p++} p==2{if(!/^---$/)print}' "$STATE_FILE" | grep -A1 "## Research Question" | tail -1 || true)

if [[ -z "$RESEARCH_QUESTION" ]]; then
    RESEARCH_QUESTION="Continue your discovery exploration."
fi

# Build system message with context
SYSTEM_MESSAGE="Discovery Loop - Iteration $NEXT_ITERATION of $MAX_ITERATIONS

Review your previous findings in .claude/discovery-findings.md and continue exploring.

When you find a significant insight, output: $COMPLETION_SIGNAL"

# Output JSON to continue loop (block exit, re-inject prompt)
cat << EOF
{
  "decision": "deny",
  "reason": "$RESEARCH_QUESTION",
  "systemMessage": "$SYSTEM_MESSAGE"
}
EOF

exit 0
