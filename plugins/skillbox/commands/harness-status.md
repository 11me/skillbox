---
name: harness-status
description: Show harness feature progress and verification status
allowed-tools: [Read, Bash]
---

# /harness-status

Display current harness state and feature progress.

## Output

Shows:
1. Session number and project type
2. Feature list with status (pending/in_progress/implemented/verified/failed)
3. Verification results for each feature
4. Overall progress percentage
5. Recommendations for next steps

## Example Output

```
## Harness Status

**Session:** #4 | **Project:** Go

### Features (2/5 verified)

| ID            | Status         | Last Verified    |
|---------------|----------------|------------------|
| auth-login    | ✅ verified     | 2026-01-04 15:30 |
| auth-logout   | ✅ verified     | 2026-01-04 16:00 |
| user-profile  | ⚙️ implemented  | —                |
| user-avatar   | 🔄 in_progress | —                |
| user-settings | ⏳ pending      | —                |

### Next Steps

1. Verify `user-profile`:
   ```bash
   /harness-verify user-profile
   ```

2. Continue implementing `user-avatar`
```

## Usage

```bash
# Show full status
/harness-status

# JSON output
/harness-status --json
```

## Implementation

1. Read `.claude/harness.json` for session info
2. Read `.claude/features.json` for features
3. Format status table with emoji indicators
4. Suggest next action based on priorities

## Status Icons

| Status | Icon |
|--------|------|
| verified | ✅ |
| implemented | ⚙️ |
| in_progress | 🔄 |
| pending | ⏳ |
| failed | ❌ |

## See Also

- `/harness-init` — Initialize harness
- `/harness-verify` — Run verification
- `/harness-update` — Update status
