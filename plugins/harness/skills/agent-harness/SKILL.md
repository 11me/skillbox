---
name: agent-harness
description: This skill should be used when the user asks about "multi-session features", "verification tracking", "long-running agents", or mentions "harness", "feature supervisor", "prevent premature completion".
context: fork
allowed-tools:
  - Read
  - Write
  - Bash
---

# Long-Running Agent Harness

Based on Anthropic's [Effective Harnesses for Long-Running Agents](https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents) article.

## Purpose / When to Use

Use this skill when:
- Starting a multi-session feature implementation
- Need to track feature verification status
- Want to prevent premature "victory" declarations
- Resuming work after context reset

## The Supervisor-Worker Pattern

Based on Anthropic's Two-Agent pattern, enhanced with subagent orchestration.

```
Session 1: INITIALIZER
├── Bootstrap environment
├── Create features.json
├── Set up verification commands
└── Hand off to coding sessions

Sessions 2+: SUPERVISED DEVELOPMENT
├── Feature Supervisor (orchestrator)
│   ├── Load harness state
│   ├── Select next feature
│   ├── Delegate to feature-dev agents
│   └── Enforce verification gate
│
├── Anthropic feature-dev (implementation)
│   ├── code-explorer → Analyze codebase
│   ├── code-architect → Design solution
│   └── code-reviewer → Quality check
│
└── Verification Worker (validation)
    ├── Run verification command
    ├── Root Cause Analysis on failure
    └── Retry up to 3 times
```

### Prerequisites

```bash
# Install Anthropic's official feature-dev plugin
npx claude-plugins install @anthropics/claude-code-plugins/feature-dev
```

## Quick Start

### First Session (Initializer)

```bash
# 1. Basic setup (if not done)
/init-workflow

# 2. Initialize harness
/harness-init

# Define features when prompted
```

### Subsequent Sessions (Supervised)

**Automated (recommended):**
```bash
# Start supervised workflow
/harness-supervisor

# Supervisor handles:
# 1. Selects next feature
# 2. Delegates to code-explorer, code-architect, code-reviewer
# 3. Implements feature
# 4. Runs verification with RCA
# 5. Retries up to 3 times on failure
# 6. Ends session with git commit
```

**Manual (alternative):**
```bash
# Check status
/harness-status

# Implement feature manually
/harness-update auth-login implemented

# Run verification
/harness-verify auth-login

# Session end (blocked if unverified)
```

## Feature Lifecycle

```
pending → in_progress → implemented → verified
                              ↓
                           failed
                              ↓
                        (fix & retry)
```

## Key Concepts

### JSON Over Markdown

Features are tracked in `.claude/features.json` (not markdown) because:
- "Model is less likely to inappropriately change or overwrite JSON files"
- Clear schema prevents ambiguous states
- Direct file modification is blocked by hook

### Mandatory Verification

Session end is blocked if features are implemented but not verified:
- Prevents premature victory declarations
- Ensures all features are tested before completion
- Linked beads tasks auto-close on verification

### Auto-Beads Integration

When adding features:
1. Beads task created automatically
2. Status synced with feature status
3. Task closed when feature verified

## Commands

| Command | Purpose |
|---------|---------|
| `/harness-init` | Initialize harness, create features |
| `/harness-supervisor` | **Automated workflow** — orchestrates feature-dev + verification |
| `/harness-status` | Show feature progress |
| `/harness-verify <id>` | Run verification, update status |
| `/harness-update <id> <status>` | Manual status update |

## Agents

| Agent | Model | Role |
|-------|-------|------|
| `feature-supervisor` | haiku | Orchestrates workflow, one feature per session |
| `verification-worker` | sonnet | Runs tests, RCA, retry logic |
| `code-explorer` | (feature-dev) | Analyzes codebase |
| `code-architect` | (feature-dev) | Designs solution |
| `code-reviewer` | (feature-dev) | Reviews quality |

## Integration with Existing Tools

| Tool | Role in Harness |
|------|-----------------|
| **Beads** | High-level task tracking (auto-linked) |
| **Serena** | Code memory, architectural decisions |
| **Checkpoints** | Session state snapshots |
| **TDD Enforcer** | Code-level test coverage |
| **Harness** | Feature-level verification tracking |

**Layer Hierarchy:**
- `/init-workflow` = Base layer (beads + serena)
- `/harness-init` = Feature tracking layer on top

## File Structure

```
.claude/
├── harness.json      # Session history, project type
├── features.json     # Feature list with status
└── init-session.sh   # Bootstrap script
```

### features.json Schema

```json
{
  "version": "1.0.0",
  "features": [
    {
      "id": "auth-login",
      "description": "User login with JWT",
      "status": "verified",
      "verification": "go test ./... -run TestAuthLogin",
      "beads_id": "skills-abc123",
      "last_verified": "2026-01-05T15:30:00",
      "verification_output": "PASS"
    }
  ]
}
```

## Patterns

### Pattern: Multi-Session Feature

**Session 1 (Initializer):**
```bash
/harness-init
# Define: auth-login, auth-logout, user-profile
```

**Session 2:**
```bash
# Harness shows: Session #2, 0/3 verified
/harness-update auth-login in_progress
# Implement login...
/harness-update auth-login implemented
/harness-verify auth-login
# ✅ verified
```

**Session 3:**
```bash
# Harness shows: Session #3, 1/3 verified
# Continue with next feature...
```

### Pattern: Recovery After Failure

```bash
/harness-verify user-profile
# ❌ failed: TestProfileUpdate assertion failed

# Fix the bug...
/harness-update user-profile in_progress
# ... fix code ...
/harness-update user-profile implemented
/harness-verify user-profile
# ✅ verified
```

## Guardrails

**MUST:**
- Initialize harness before multi-session work
- Verify features before marking complete
- Run `/harness-verify --implemented` before session end

**NEVER:**
- Skip verification step
- Directly modify features.json (blocked by hook)
- Declare victory without all features verified

## Parallel Execution

When working with independent tasks or sub-features:

**ALWAYS run Task agents in PARALLEL** when:
- Multiple independent code searches needed
- Analyzing different parts of codebase
- Running independent verifications
- Exploring multiple approaches simultaneously

**Example - Sequential (WRONG):**
```
Task(code-explorer) → wait → Task(code-architect) → wait → ...
```

**Example - Parallel (CORRECT):**
```
Task(code-explorer) ─┬─► Results
Task(code-architect) ─┤
Task(code-reviewer)  ─┘
```

**Implementation:**
Send a SINGLE message with MULTIPLE Task tool calls - they execute in parallel.

**Keep sequential when:**
- Task B depends on result of Task A
- Sequential ordering is required for correctness

## Related Skills

- **workflow-orchestration** — Task-to-code traceability
- **beads-workflow** — Task tracking details
- **tdd-enforcer** — Code-level testing
- **context-engineering** — Context management
- **reliable-execution** — Persistence patterns

## Version History

- 1.2.0 — Add Parallel Execution guidelines for independent tasks
- 1.1.0 — Add Supervisor-Worker pattern with feature-dev integration
- 1.0.0 — Initial release (based on Anthropic article)
