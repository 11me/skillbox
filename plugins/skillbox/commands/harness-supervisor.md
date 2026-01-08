---
name: harness-supervisor
description: Start supervised feature development with automatic orchestration
allowed-tools: [Bash, Read, Task]
---

# /harness-supervisor

Start automated feature development with the Supervisor-Worker pattern.

## What This Does

1. **Loads harness state** from `.claude/features.json` and `.claude/harness.json`
2. **Selects next feature** (in_progress > pending > implemented)
3. **Orchestrates development** via feature-dev agents (code-explorer, code-architect, code-reviewer)
4. **Runs verification** with Root Cause Analysis and retry (up to 3 attempts)
5. **Ends session** cleanly with git commit or failure report

## Prerequisites

**Required plugins:**
```bash
npx claude-plugins install @anthropics/claude-code-plugins/feature-dev
```

**Initialized harness:**
```bash
/harness-init
```

## Usage

```bash
# Full supervised workflow (one feature per session)
/harness-supervisor

# Just show status without starting work
/harness-supervisor --status

# Continue current in_progress feature
/harness-supervisor --continue

# Only run verification on implemented features
/harness-supervisor --verify-only
```

## Workflow

```
/harness-supervisor
        │
        ▼
┌─────────────────────────┐
│ Load State              │
│ features.json           │
│ harness.json            │
└─────────────────────────┘
        │
        ▼
┌─────────────────────────┐
│ Select Feature          │
│ Priority: in_progress   │
│ > pending > implemented │
└─────────────────────────┘
        │
        ▼
┌─────────────────────────┐
│ Delegate to Agents      │
│ 1. code-explorer        │
│ 2. code-architect       │
│ 3. [implement]          │
│ 4. code-reviewer        │
└─────────────────────────┘
        │
        ▼
┌─────────────────────────┐
│ Verification            │
│ Run tests               │
│ RCA if fails            │
│ Retry up to 3x          │
└─────────────────────────┘
        │
        ▼
┌─────────────────────────┐
│ Session End             │
│ Git commit (if passed)  │
│ Report status           │
└─────────────────────────┘
```

## Session Output

### On Success

```
┌─────────────────────────────────────────┐
│ Session #4 | Project: Go               │
│ Features: 3/5 verified                  │
│                                         │
│ This session: user-profile (pending)   │
└─────────────────────────────────────────┘

Exploring codebase...
Designing implementation...
Implementing feature...
Reviewing code quality...

Running verification (attempt 1/3)...
✅ All tests passed

✅ Feature user-profile verified!
   Session complete. Remaining: 1 feature (settings)

   Git commit: feat(user): add profile editing
```

### On Failure

```
┌─────────────────────────────────────────┐
│ Session #4 | Project: Go               │
│ Features: 3/5 verified                  │
│                                         │
│ This session: user-profile (pending)   │
└─────────────────────────────────────────┘

Running verification (attempt 1/3)...
❌ Test failed: TestUserProfile/update

Root Cause Analysis:
  Error: nil pointer in user.Update()
  Fix: Add nil check

Applying fix and retrying (attempt 2/3)...
❌ Test failed: TestUserProfile/update

Retrying (attempt 3/3)...
❌ Test failed: TestUserProfile/update

❌ Feature user-profile failed verification

   Last error: nil pointer dereference
   Attempts: 3/3
   Status: NEEDS_HUMAN

   Manual intervention required.
   Run /harness-update user-profile pending to reset.
```

## One Feature Per Session

**Critical constraint:** This command works on exactly ONE feature.

When the feature is verified (or fails after 3 attempts), the session ends. The next session will pick up the next feature automatically.

This ensures:
- Clean git history
- Reduced blast radius
- Clear session boundaries
- Easy rollback

## Integration

| Command | Purpose |
|---------|---------|
| `/harness-init` | Initialize harness (run first) |
| `/harness-supervisor` | Automated feature development |
| `/harness-status` | Manual status check |
| `/harness-verify` | Manual verification |
| `/harness-update` | Manual status update |

## Agents Used

| Agent | Source | Role |
|-------|--------|------|
| feature-supervisor | skillbox | Orchestrates workflow |
| code-explorer | feature-dev | Analyzes codebase |
| code-architect | feature-dev | Designs solution |
| code-reviewer | feature-dev | Reviews quality |
| verification-worker | skillbox | Runs tests + RCA |

## See Also

- `skills/core/agent-harness/SKILL.md` — Full documentation
- Anthropic article: "Effective Harnesses for Long-Running Agents"
