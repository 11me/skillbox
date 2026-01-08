---
name: harness-verify
description: Run verification for features and update status
allowed-tools: [Bash, Read, Write]
---

# /harness-verify

Run verification commands for features and update their status.

## Usage

```bash
# Verify specific feature
/harness-verify auth-login

# Verify all implemented features
/harness-verify --implemented

# Verify all features (full test suite)
/harness-verify --all
```

## Process

1. Read feature's verification command from `features.json`
2. Execute verification (test command, e2e, etc.)
3. Update feature status based on result:
   - Pass → `verified`
   - Fail → `failed` (with output captured)
4. Close linked beads task if verified
5. Report results

## Example

```bash
/harness-verify auth-login
```

Output:
```
## Verifying: auth-login

**Description:** User login with JWT
**Command:** go test ./internal/auth/... -run TestAuthLogin

Running verification...

✅ PASSED

TestAuthLogin/valid_credentials PASS
TestAuthLogin/invalid_password PASS
TestAuthLogin/user_not_found PASS

Updated status: implemented → verified
Closed beads task: skills-abc123
```

## Verification Commands

Features should specify verification in features.json:

```json
{
  "id": "auth-login",
  "description": "User login with JWT",
  "verification": "go test ./internal/auth/... -run TestAuthLogin"
}
```

If no verification command specified, uses project default:

| Project | Default Command |
|---------|-----------------|
| Go | `go test ./... -run <FeatureId>` |
| Node | `pnpm vitest run --grep '<feature>'` |
| Python | `pytest -k <feature>` |
| Rust | `cargo test <feature>` |

## Failure Handling

On failure:
1. Status set to `failed`
2. Error output captured in `verification_output`
3. Beads task NOT closed
4. Suggestion to fix and retry

```
## Verifying: user-profile

Running: go test ./internal/user/... -run TestProfile

❌ FAILED

--- FAIL: TestProfile/update_email
    user_test.go:45: expected email to be updated

Updated status: implemented → failed

**Next steps:**
1. Fix the failing test
2. Run `/harness-verify user-profile` again
```

## Batch Verification

```bash
# Verify only implemented features
/harness-verify --implemented

# Verify everything
/harness-verify --all
```

Output:
```
## Batch Verification

| Feature       | Result |
|---------------|--------|
| auth-login    | ✅ PASS |
| auth-logout   | ✅ PASS |
| user-profile  | ❌ FAIL |

2/3 passed
```

## See Also

- `/harness-status` — View current status
- `/harness-update` — Manual status update
