---
name: workflow-orchestration
description: Unified workflow integrating beads task tracking, serena code memory, and conventional commits. Use when starting features, tracking work across sessions, or maintaining task-to-code traceability.
---

# Workflow Orchestration

## Purpose / When to Use

Use this skill when:
- Starting a new feature or task
- Need to track work across multiple sessions
- Want traceability from task → code → commit
- Resuming work after context reset

## The Unified Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                    WORKFLOW CYCLE                           │
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌───────┐ │
│  │  Task    │───►│  Code    │───►│  Commit  │───►│ Close │ │
│  │ (beads)  │    │ (serena) │    │ (conv)   │    │ (bd)  │ │
│  └──────────┘    └──────────┘    └──────────┘    └───────┘ │
│       │               │                                     │
│       └───────────────┴── Memory persists across sessions ──┘
└─────────────────────────────────────────────────────────────┘
```

## Phase 1: Task Creation (Beads)

Start every feature with a tracked task:

```bash
# Create task
bd create --title "Implement user authentication" -t feature

# Output: Created BEADS-42
```

This creates a persistent record that survives context resets.

## Phase 2: Code Discovery (Serena)

Use serena to explore and document the codebase:

```
# Find relevant symbols
find_symbol "authenticate"

# Explore references
find_referencing_symbols "UserService"

# Save discoveries to memory
write_memory "auth-architecture.md" "Authentication flow: ..."
```

**Why this matters:** Serena memories persist across sessions. When you resume, read the memory instead of re-exploring.

## Phase 3: Implementation

Write code with task context in mind:
- Keep changes focused on the task
- Update serena memory with architectural decisions
- Track subtasks in beads if needed

```bash
# Create subtask if scope grows
bd create --title "Add password hashing" -t task --parent BEADS-42
```

## Phase 4: Commit with Reference

Use conventional commit with task reference:

```bash
git commit -m "feat(auth): implement user authentication

Refs: BEADS-42"
```

This creates an audit trail linking code changes to tasks.

## Phase 5: Close Task

When feature is complete:

```bash
bd close BEADS-42 --reason "Implemented with tests, merged in PR #123"
```

## Patterns

### Pattern: Resume After Context Reset

When starting a new session on existing work:

1. Check active tasks: `bd list --state open`
2. Read relevant memories: `read_memory "feature-name.md"`
3. Find where you left off: `find_symbol "last_function_modified"`
4. Continue from there

### Pattern: Feature Branch Workflow

```bash
# 1. Create task
bd create --title "Add dark mode" -t feature  # → BEADS-55

# 2. Create branch (optional: include task ID)
git checkout -b feature/BEADS-55-dark-mode

# 3. Work with serena for discovery
find_symbol "ThemeProvider"
write_memory "dark-mode-impl.md" "Found ThemeProvider at..."

# 4. Commit with references
git commit -m "feat(ui): add dark mode toggle

Refs: BEADS-55"

# 5. Close on merge
bd close BEADS-55
```

### Pattern: Multi-Session Feature

For features spanning multiple days:

**Session 1:**
```bash
bd create --title "API redesign" -t epic  # → BEADS-100
# Explore architecture
write_memory "api-redesign-plan.md" "Current state: ..."
bd update BEADS-100 --state in_progress
```

**Session 2 (after context reset):**
```bash
bd list --state in_progress  # Find BEADS-100
read_memory "api-redesign-plan.md"  # Resume context
# Continue work
```

**Session N:**
```bash
bd close BEADS-100 --reason "Completed in 5 sessions"
```

## Integration Points

| Tool | Role | Key Commands |
|------|------|--------------|
| **beads** | Task lifecycle | `bd create`, `bd list`, `bd close` |
| **serena** | Code memory | `find_symbol`, `write_memory`, `read_memory` |
| **conventional-commit** | Audit trail | Commit message with `Refs: BEADS-ID` |

## Guardrails

**MUST:**
- Create a beads task before starting significant work
- Reference task ID in commit messages
- Write serena memory for architectural decisions

**NEVER:**
- Lose context by not using memories
- Create orphan commits without task references
- Skip closing tasks when work is complete

## Examples

Trigger prompts:
- "Start working on feature X"
- "Track this task"
- "How do I resume my previous work?"
- "Create a beads task for this"
- "What's the workflow for new features?"

## Related Skills

- **beads-workflow** — Detailed beads CLI reference
- **serena-navigation** — Semantic code navigation
- **conventional-commit** — Commit message format
- **context-engineering** — Long session management

## Version History

- 1.0.0 — Initial release: unified workflow pattern
