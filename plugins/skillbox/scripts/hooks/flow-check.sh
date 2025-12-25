#!/usr/bin/env bash
set -euo pipefail

# flow-check.sh — SessionStart hook
# Checks production workflow compliance and suggests improvements

output=""
missing_count=0

# Check CLAUDE.md presence
if [ -f "CLAUDE.md" ] || [ -f ".claude/CLAUDE.md" ]; then
    : # Present, good
else
    output+="⚠️ **Missing:** CLAUDE.md (AI context file)"$'\n'
    ((missing_count++)) || true
fi

# Check pre-commit hooks
if [ -f ".pre-commit-config.yaml" ]; then
    if [ -f ".git/hooks/pre-commit" ]; then
        : # Installed
    else
        output+="⚠️ **Pre-commit not installed:** Run \`pre-commit install\`"$'\n'
        ((missing_count++)) || true
    fi
else
    output+="⚠️ **Missing:** .pre-commit-config.yaml"$'\n'
    ((missing_count++)) || true
fi

# Check beads initialization
if [ -d ".beads" ]; then
    : # Present
else
    output+="ℹ️ **Optional:** No .beads/ directory (task tracking)"$'\n'
fi

# Check tests directory
TESTS_FOUND=false
for test_dir in tests test __tests__ spec; do
    if [ -d "$test_dir" ]; then
        TESTS_FOUND=true
        break
    fi
done

# Check for test files in common patterns
if [ "$TESTS_FOUND" = false ]; then
    if compgen -G "*_test.go" > /dev/null 2>&1 || \
       compgen -G "*_test.py" > /dev/null 2>&1 || \
       compgen -G "*.test.ts" > /dev/null 2>&1 || \
       compgen -G "*.spec.ts" > /dev/null 2>&1; then
        TESTS_FOUND=true
    fi
fi

if [ "$TESTS_FOUND" = false ]; then
    output+="ℹ️ **No tests found:** Consider adding tests/"$'\n'
fi

# Summary
if [ "$missing_count" -ge 2 ]; then
    output+=$'\n'"💡 **Suggestion:** Multiple workflow components missing."$'\n'
    output+="   Consider running project initialization to set up:"$'\n'
    output+="   - CLAUDE.md for AI context"$'\n'
    output+="   - Pre-commit hooks for quality gates"$'\n'
    output+="   - Beads for task tracking"$'\n'
fi

# Only output if there's something to report
if [ -n "$output" ]; then
    echo -e "**Workflow Check:**"$'\n'"$output"
fi
