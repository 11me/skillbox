---
name: checkpoint
description: Save session progress to serena memory
---

# /checkpoint

Save current session state to serena memory for recovery after context reset.

## Usage

### Quick Checkpoint
```bash
/checkpoint
```
Creates: `checkpoint-YYYY-MM-DD-HHMM.md`

### Named Checkpoint
```bash
/checkpoint pre-refactor
/checkpoint feature-complete
```
Creates: `checkpoint-<name>.md`

### Restore
```bash
/checkpoint restore
```
Lists available checkpoints and restores selected one.

## What Gets Saved

1. **Beads task** — current task ID and status
2. **Completed work** — what was done this session
3. **Remaining work** — what's left
4. **Current state** — files modified, test status
5. **Next steps** — immediate actions to resume

## Checkpoint Template

Use this structure:

```markdown
# Session Checkpoint: <name>

## Task Context
- **Beads Task:** <id> - <title>
- **Goal:** <what we're working on>
- **Status:** <in progress / complete / blocked>

## Completed
- [x] What was done

## Remaining
- [ ] What's left

## Current State
- **Last modified:** <file>
- **Tests passing:** <yes/no>
- **Blockers:** <any issues>

## Key Files
- `path/to/file` — <description>

## Next Steps
1. <immediate next action>

## Recovery
To continue: read this checkpoint, open key files, proceed with next steps.
```

## Integration

- Works with `beads-workflow` for task context
- Works with `serena-navigation` for memory storage
- Works with `workflow-orchestration` pattern

## See Also

- `/init-workflow` — Initialize workflow tools
- `session-checkpoint` agent — Automated checkpointing
