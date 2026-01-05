---
description: Toggle automatic supervisor mode for harness
allowed-tools: Bash, Read, Write
argument-hint: [on|off|status]
---

# /harness-auto

Toggle automatic supervisor mode. When enabled, `/harness-supervisor` runs automatically at session start.

## Usage

```bash
# Enable auto mode
/harness-auto on

# Disable auto mode
/harness-auto off

# Check current status
/harness-auto status
```

## How It Works

Creates/updates `.claude/harness-config.json`:

```json
{
  "auto_supervisor": true
}
```

When `auto_supervisor: true` AND features exist that aren't all verified:
- SessionStart hook outputs strong instruction to invoke feature-supervisor
- Claude automatically starts supervised workflow
- No manual `/harness-supervisor` needed

## Session Flow (Auto Mode)

```
Session Start
    │
    ▼
session_bootstrap.py hook
    │
    ├─► Check harness-config.json
    │   auto_supervisor: true?
    │
    ├─► YES + unverified features exist
    │   → Output: "AUTO-START feature-supervisor NOW"
    │   → Claude invokes feature-supervisor agent
    │
    └─► NO or all verified
        → Normal status output
```

## When to Use

**Enable auto mode when:**
- Working on multi-session feature implementation
- Want hands-off operation
- Trust the verification commands

**Disable auto mode when:**
- Debugging or exploring
- Want manual control
- Need to check status first

## Implementation

When you run `/harness-auto on`:

```bash
# Create/update config
mkdir -p .claude
cat > .claude/harness-config.json << 'EOF'
{
  "auto_supervisor": true
}
EOF
```

When you run `/harness-auto off`:

```bash
# Update config
cat > .claude/harness-config.json << 'EOF'
{
  "auto_supervisor": false
}
EOF
```
