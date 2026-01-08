#!/usr/bin/env bash
#
# setup-loop.sh - Initialize discovery loop state file
# Based on Ralph Wiggum pattern
#

set -euo pipefail

# Default values
MAX_ITERATIONS=10
COMPLETION_SIGNAL="INSIGHT_COMPLETE"
RESEARCH_QUESTION=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --max-iterations)
            MAX_ITERATIONS="$2"
            shift 2
            ;;
        --completion-promise|--completion-signal)
            COMPLETION_SIGNAL="$2"
            shift 2
            ;;
        *)
            # Accumulate non-option arguments as research question
            if [[ -n "$RESEARCH_QUESTION" ]]; then
                RESEARCH_QUESTION="$RESEARCH_QUESTION $1"
            else
                RESEARCH_QUESTION="$1"
            fi
            shift
            ;;
    esac
done

# Validate research question
if [[ -z "$RESEARCH_QUESTION" ]]; then
    echo "Error: Research question is required."
    echo ""
    echo "Usage: /discover-loop \"your research question\" [--max-iterations N] [--completion-signal TEXT]"
    echo ""
    echo "Options:"
    echo "  --max-iterations N     Maximum number of discovery iterations (default: 10)"
    echo "  --completion-signal    Text to signal discovery complete (default: INSIGHT_COMPLETE)"
    exit 1
fi

# Validate max_iterations is numeric
if ! [[ "$MAX_ITERATIONS" =~ ^[0-9]+$ ]]; then
    echo "Error: --max-iterations must be a positive integer, got: $MAX_ITERATIONS"
    exit 1
fi

# Create .claude directory if it doesn't exist
mkdir -p .claude

# Get current timestamp (ISO 8601)
STARTED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create state file with YAML frontmatter
STATE_FILE=".claude/discovery-loop.local.md"

cat > "$STATE_FILE" << EOF
---
active: true
iteration: 1
max_iterations: $MAX_ITERATIONS
question_chain: []
findings: []
completion_signal: "<discovered>${COMPLETION_SIGNAL}</discovered>"
started_at: "$STARTED_AT"
---

# Discovery Loop

## Research Question

$RESEARCH_QUESTION

## Instructions

You are in a discovery loop. Each iteration:

1. **Review** this file and any previous findings
2. **Generate** the next most important question to explore
3. **Explore** that question using available tools
4. **Document** your finding in .claude/discovery-findings.md
5. **Check** if you've found a significant insight

When you discover a key insight that answers the research question, output:

\`\`\`
<discovered>${COMPLETION_SIGNAL}</discovered>
\`\`\`

This will end the loop. Otherwise, the loop will continue to the next iteration.

## Question Patterns

- What am I assuming about this problem?
- What existing pattern solves similar problems?
- What would an expert in [other field] notice?
- What's the simplest version that could work?
- What would make this problem trivial?
- Who else has this problem, and what did they do?

## Current Iteration

This is iteration 1 of $MAX_ITERATIONS.

Focus on: **Decomposing the problem into sub-questions**

Good luck with your discovery!
EOF

# Create findings file
FINDINGS_FILE=".claude/discovery-findings.md"
if [[ ! -f "$FINDINGS_FILE" ]]; then
    cat > "$FINDINGS_FILE" << EOF
# Discovery Findings

Research question: $RESEARCH_QUESTION

Started: $STARTED_AT

---

EOF
fi

# Output confirmation
echo "Discovery loop initialized!"
echo ""
echo "Research question: $RESEARCH_QUESTION"
echo "Max iterations: $MAX_ITERATIONS"
echo "Completion signal: <discovered>${COMPLETION_SIGNAL}</discovered>"
echo ""
echo "State file: $STATE_FILE"
echo "Findings file: $FINDINGS_FILE"
echo ""
echo "The loop will continue until you output the completion signal or reach max iterations."
echo ""
echo "---"
echo ""
echo "## Begin Discovery"
echo ""
echo "$RESEARCH_QUESTION"
