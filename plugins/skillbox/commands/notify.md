---
name: notify
description: Toggle desktop notifications for Claude Code events
arguments:
  - name: action
    description: "on, off, or status (default: status)"
    required: false
---

# Desktop Notifications Toggle

Action: **${action:-status}**

## Instructions

### If action is "on":
Create or update `.claude/skillbox.local.md` with:
```yaml
---
notifications: true
---
```

### If action is "off":
Create or update `.claude/skillbox.local.md` with:
```yaml
---
notifications: false
---
```

### If action is "status" (default):
1. Check if `.claude/skillbox.local.md` exists
2. Read the `notifications` field from YAML frontmatter
3. Report current state:
   - If file exists and `notifications: false` → "Notifications are **disabled**"
   - Otherwise → "Notifications are **enabled** (default)"

## Notes

- Notifications are enabled by default
- Requires `notify-send` command (Linux)
- Notification types:
  - **Block**: When Claude is blocked by a hook (critical urgency)
  - **Ask**: When Claude needs user input (normal urgency)
  - **Done**: When task completes (low urgency)
  - **Permission**: When Claude needs permission (normal urgency)
