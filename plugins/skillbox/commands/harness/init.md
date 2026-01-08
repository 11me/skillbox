---
name: init
description: Initialize long-running agent harness with bootstrap and feature tracking
allowed-tools:
  - Bash
  - Read
  - mcp__plugin_serena_serena__write_memory
---

# /harness-init

Initialize the long-running agent harness for this project.

## What This Does

1. **Bootstrap Environment** (unless --skip-setup)
   - Runs project-appropriate setup commands
   - Verifies dependencies are installed
   - Confirms build/tests pass

2. **Create Feature Tracker**
   - Creates `.claude/features.json` for immutable feature tracking
   - Prompts for initial feature list
   - Sets up verification commands

3. **Record Harness State**
   - Creates `.claude/harness.json` with session history
   - Links to beads tasks if available
   - Generates `.claude/init-session.sh` for future sessions

## Usage

```bash
# Full initialization
/harness-init

# Skip setup commands (already done)
/harness-init --skip-setup

# Only create features.json
/harness-init --features-only
```

## Initialization Flow

### Step 1: Detect Project Type

Read project files to determine type (Go, Node, Python, Rust).

### Step 2: Run Bootstrap (unless skipped)

Execute project-specific setup:

**Go:**
```bash
go mod download
go build ./...
```

**Node/pnpm:**
```bash
pnpm install
pnpm run build
```

**Python/uv:**
```bash
uv sync
```

### Step 3: Define Features

Ask user for features to track. For each feature:

1. Generate unique ID from description (e.g., "User login" → "auth-login")
2. Determine verification command (project-specific test pattern)

### Step 4: Create Files via Script

Instead of using Write tool (blocked by guard hook), call initialization script:

```bash
python3 ${CLAUDE_PLUGIN_ROOT}/scripts/hooks/harness_init.py \
  --project-dir "$(pwd)" \
  --features '[
    {"id": "auth-login", "description": "User login with JWT"},
    {"id": "user-profile", "description": "User profile editing"}
  ]'
```

Use `--no-beads` flag to skip automatic beads task creation.

The script:
- Creates `.claude/harness.json` with session tracking
- Creates `.claude/features.json` with initial pending statuses
- Auto-links beads tasks if available (unless --no-beads)
- Uses direct Python I/O to bypass the PreToolUse guard hook

### Step 5: Generate init-session.sh

Create `.claude/init-session.sh` for quick future bootstrap.

## Feature List Template

When prompted, list features like:

```
1. User login with JWT
2. User profile editing
3. Password reset via email
```

Or provide JSON directly:

```json
[
  {"id": "auth-login", "description": "User login with JWT"},
  {"id": "user-profile", "description": "User profile editing"}
]
```

## After Initialization

The harness will:
- Track feature completion status
- Block session end until verification passes
- Persist state across context resets
- Generate init.sh for future sessions

## Integration

Works with existing tools:

| Tool | Role |
|------|------|
| `/init-workflow` | Base setup (beads + serena) — run first |
| `/harness-init` | Feature tracking layer on top |

## See Also

- `/harness-status` — Show feature progress
- `/harness-verify` — Run verification suite
- `/harness-update` — Manual status update
