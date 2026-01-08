---
name: update
description: Manually update feature status in harness
argument-hint: "<feature-id> <status>"
allowed-tools:
  - Read
  - Write
---

# /harness-update

Manually update a feature's status without running verification.

## Usage

```bash
# Start working on a feature
/harness-update auth-login in_progress

# Mark as implemented (ready for verification)
/harness-update auth-login implemented

# Force mark as verified (skip verification)
/harness-update auth-login verified --force

# Mark as pending (reset)
/harness-update auth-login pending
```

## Valid Statuses

| Status | Description | Beads Action |
|--------|-------------|--------------|
| `pending` | Not started | — |
| `in_progress` | Being worked on | Update to in_progress |
| `implemented` | Code complete, needs verification | — |
| `verified` | Passed verification | Close task |
| `failed` | Failed verification | — |

## Workflow

Typical status progression:

```
pending → in_progress → implemented → verified
                              ↓
                           failed
                              ↓
                        in_progress (retry)
```

## Examples

### Start Feature

```bash
/harness-update user-profile in_progress
```
```
Updated user-profile: pending → in_progress
Updated beads task skills-abc123 status
```

### Mark Implemented

```bash
/harness-update user-profile implemented
```
```
Updated user-profile: in_progress → implemented

Next: Run /harness-verify user-profile
```

### Force Verify

Use `--force` to skip verification (for features tested manually):

```bash
/harness-update user-profile verified --force
```
```
⚠️ Forcing status to verified without running verification

Updated user-profile: implemented → verified
Closed beads task: skills-abc123
```

### Reset Feature

```bash
/harness-update user-profile pending
```
```
Reset user-profile to pending
```

## Notes

- Prefer `/harness-verify` over manual `verified` status
- `--force` flag required to set `verified` without running tests
- Beads task auto-updated when status changes to `in_progress` or `verified`

## See Also

- `/harness-verify` — Run verification (recommended)
- `/harness-status` — View all features
