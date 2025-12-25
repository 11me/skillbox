---
name: beads-workflow
description: Task tracking with beads CLI (bd). Use when managing tasks, tracking progress, creating issues, or when user mentions "task", "issue", "todo", "beads", "bd", "next task", "create task", "close task".
globs: ["**/.beads/**", "**/beads.db"]
allowed-tools: Bash, Read, Grep, Glob
---

# Beads Workflow

Proactive task management for projects using beads issue tracker.

## Quick Reference

### Task Lifecycle

```bash
# Create task
bd create --title "Implement feature" -t task -p 1

# Start work (pick from ready)
bd update <id> --status in_progress

# Add progress note
bd comments add <id> "Completed base template"

# Complete task
bd close <id> --reason "Implemented with tests"

# Quick capture (returns ID only)
bd q "Fix login bug"
```

### Priority Levels

| Priority | Value | Usage |
|----------|-------|-------|
| P0 | 0 | Critical, drop everything |
| P1 | 1 | High, do next |
| P2 | 2 | Medium, this sprint |
| P3 | 3 | Low, backlog |
| P4 | 4 | Someday/maybe |

### Common Commands

```bash
# View tasks
bd list                          # All tasks
bd ready                         # Ready to work
bd list --status in_progress     # Currently active

# Task details
bd show <id>                     # Full task info
bd prime                         # Current task context (for injection)

# Subtasks
bd create --title "Subtask" --parent <parent-id>

# Sync state
bd sync                          # Save to storage
```

## Session Protocol

### At Session Start

1. Check for beads: `ls -d .beads 2>/dev/null`
2. View ready tasks: `bd ready`
3. Pick task: `bd update <id> --status in_progress`

**The SessionStart hook automatically injects beads context** if `bd` is available.

### During Work

- Track current task ID in conversation
- Use `bd comments add <id> "note"` for milestones
- Create subtasks for discovered work: `bd create --title "..." --parent <id>`

### At Session End

1. Update task status (complete or leave in_progress)
2. Run `bd sync` to save state
3. Add handoff notes: `bd comments add <id> "Stopped at: ..."`

## Beads vs TodoWrite

| TodoWrite | Beads |
|-----------|-------|
| Session-only visibility | Persists across sessions |
| No history | Full audit trail |
| No dependencies | Supports parent/child and blockers |
| Visible in Claude Code UI | Git-tracked in .beads/ |

**Recommendation:**
- Use **beads** for main tasks and features
- Use **TodoWrite** for session subtasks and quick checklists

## Task Creation Best Practices

### Good Task Titles

```bash
# Action + Object pattern
bd create --title "Implement user authentication API" -t task -p 1
bd create --title "Fix login timeout on slow networks" -t task -p 1
bd create --title "Add dark mode toggle to settings" -t task -p 2
bd create --title "Write tests for payment service" -t task -p 2
```

### With Labels

```bash
# Add labels for categorization
bd create --title "Implement webhook handler" -t task -p 1 -l "backend"
bd create --title "Fix Helm chart ingress" -t task -p 1 -l "k8s"
```

## Task Completion

When user says "done", "готово", "close task":

1. Confirm which task (if ambiguous)
2. Ask for completion reason:
   - "Implemented" — feature complete
   - "Fixed" — bug resolved
   - "Not relevant" — cancelled
   - Custom reason
3. Close: `bd close <id> --reason "<reason>"`
4. Offer next task from ready list

## Integration with Skills

Beads integrates with other skillbox skills:

- **SessionStart hook** automatically shows ready tasks
- **helm-chart-developer** can track chart work
- **conventional-commit** can reference task ID in commits

Example commit with task reference:
```
feat(api): implement user auth endpoint

Refs: BEADS-123
```

## Anti-Patterns

| ❌ Avoid | ✅ Instead |
|----------|-----------|
| Creating tasks without priority | Always set `-p` flag |
| Vague titles like "Fix bug" | Be specific: "Fix login timeout" |
| Leaving tasks in_progress forever | Close or defer when blocked |
| Not syncing at session end | Run `bd sync` before exit |
| Creating many small tasks | Use parent/child for subtasks |

## Version History

- 1.0.0 — Initial release with core workflow
