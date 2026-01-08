---
name: unified-workflow
description: This skill should be used when the user asks to "start feature", "implement task", "ship code", or wants a complete workflow combining task tracking, code memory, and session management.
allowed-tools: [Read, Write, Bash, Grep, Glob]
---

# Unified Development Workflow

Complete workflow from task creation to code delivery.

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────┐
│                    WORKFLOW STAGES                               │
├─────────────────────────────────────────────────────────────────┤
│  1. INIT      → pre-commit, beads, serena, CLAUDE.md            │
│  2. PLAN      → EnterPlanMode or feature planning               │
│  3. DEVELOP   → TDD (Red→Green→Refactor), convention skills     │
│  4. VERIFY    → pre-commit, type check, SAST, tests             │
│  5. SHIP      → /commit, PR, merge to main                      │
└─────────────────────────────────────────────────────────────────┘
```

## Workflow Diagram

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

## 4-Layer Persistence Stack

```
┌─────────────────────────────────────────────┐
│              Session Layer                   │
│  TodoWrite — visible progress tracking      │
│  (Volatile: lost on context reset)          │
└─────────────────────────────────────────────┘
                    ▼
┌─────────────────────────────────────────────┐
│              Task Layer                      │
│  Beads — high-level task lifecycle          │
│  (Persistent: survives sessions)            │
└─────────────────────────────────────────────┘
                    ▼
┌─────────────────────────────────────────────┐
│            Knowledge Layer                   │
│  Serena Memories — persistent discoveries   │
│  (Persistent: survives everything)          │
└─────────────────────────────────────────────┘
                    ▼
┌─────────────────────────────────────────────┐
│              Code Layer                      │
│  Git commits — permanent artifacts          │
│  (Permanent: version controlled)            │
└─────────────────────────────────────────────┘
```

**Key principle:** Each layer is a fallback for the one above.

---

## 1. Project Setup

### New Project

```bash
# Initialize
mkdir my-project && cd my-project
git init

# Language-specific (choose one)
go mod init github.com/user/project  # Go
uv init                               # Python
pnpm init                             # TypeScript

# Quality gates
pre-commit install
pre-commit install --hook-type commit-msg

# Task tracking
bd init
```

### Create CLAUDE.md

```bash
cat > CLAUDE.md << 'EOF'
# Project: my-project

## Tech Stack
- Language: [Go/TypeScript/Python]
- Framework: [if applicable]

## Development Commands
```bash
# Tests
[test command]

# Lint
[lint command]
```

## Conventions
- [Key convention 1]
- [Key convention 2]
EOF
```

---

## 2. Feature Development Flow

### Step 1: Create Task

```bash
bd create --title "Implement user authentication" -t feature -p 1
bd update <id> --status in_progress
```

### Step 2: Plan (optional)

For complex features, use EnterPlanMode for research and approval.

### Step 3: Develop with TDD

```
RED:    Write failing test → Commit: "test: add failing test for X"
GREEN:  Write minimal code → Commit: "feat: implement X"
REFACTOR: Improve code   → Commit: "refactor: clean up X"
```

### Step 4: Save Discoveries

```
write_memory(
  memory_file_name="auth-patterns.md",
  content="# Auth Patterns\n\nDiscovered that..."
)
```

### Step 5: Checkpoint Progress

```bash
/checkpoint  # After significant progress
```

### Step 6: Ship

```bash
pre-commit run --all-files  # Verify
/commit                      # Commit with conventional format
git push
bd close <id> --reason "Implemented with tests"
```

---

## 3. Session Recovery

### Pre-Flight Checklist

Before starting significant work, verify:

1. **Task exists?** `bd show` → If not: `bd create`
2. **Context loaded?** SessionStart hook injected `bd prime`?
3. **TodoWrite initialized?** Create subtasks for implementation

### Recovery Flow

When starting a new session:

```bash
# 1. Check for checkpoints
list_memories()

# 2. Read most recent
read_memory("checkpoint-<latest>.md")

# 3. Resume task
bd update <task-id> --status in_progress

# 4. Follow "Next Steps" from checkpoint
```

---

## 4. Checkpoint Protocol

### When to Checkpoint

| Trigger | What to Include |
|---------|-----------------|
| Major step completed | Completed work, next steps |
| Before refactoring | Current state, rollback plan |
| Context feeling full | Full state, all discoveries |
| Session ending | Summary, recovery instructions |

### Checkpoint Structure

```markdown
# Checkpoint: <task-name>

**Task ID:** <beads-id>
**Status:** in_progress

## Completed Work
- [x] Item 1
- [x] Item 2

## Remaining Work
- [ ] Item 3
- [ ] Item 4

## Current State
- Last file: <path>
- Test status: <passing/failing>
- Blockers: <none/description>

## Key Discoveries
1. <Important finding>

## Next Steps
1. <Exact next action>

## Recovery
1. `bd update <id> --status in_progress`
2. Read `<file>` lines X-Y
3. Continue implementing <feature>
```

---

## Quick Commands

| Command | Purpose |
|---------|---------|
| `/init-workflow` | Init beads + serena + CLAUDE.md |
| `/checkpoint` | Save session progress |
| `/commit` | Create conventional commit |
| `bd create` | Create beads task |
| `bd close <id>` | Complete task |

---

## Guardrails

**MUST:**
- Create beads task before significant work
- Reference task ID in commit messages
- Write serena memory for architectural decisions
- Checkpoint before context overflow

**NEVER:**
- Lose context by not using memories
- Create orphan commits without task references
- Skip closing tasks when work is complete
- Wait until end to checkpoint (risk of loss)

---

## Anti-Patterns

| Anti-Pattern | Why Bad | Do Instead |
|--------------|---------|------------|
| No task tracking | Lost context | Start with `bd create` |
| Discoveries only in chat | Lost on reset | Use `write_memory()` |
| No progress visibility | Hard to resume | Use `TodoWrite` |
| Waiting to checkpoint | Risk of loss | After each major step |
| Large uncommitted changes | Risk of loss | Commit incrementally |
| Skip tests | Bugs in prod | Always TDD |
| Giant commits | Hard to review | Small, focused |

---

## Related Skills

- **beads-workflow** — Detailed task tracking
- **serena-navigation** — Code memory and exploration
- **conventional-commit** — Commit message format
- **context-engineering** — Context budget management
- **tdd-enforcer** — Red-Green-Refactor patterns

## Version History

- 1.0.0 — Unified from workflow-orchestration, production-flow, reliable-execution
