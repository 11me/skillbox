---
name: secrets-check
description: Scan project for secrets and credentials
---

# Secrets Check

You are scanning the current project for secrets and credentials.

## Workflow

### Step 1: Check Protection Setup

First, verify if the project has secrets protection:

```bash
ls -la .pre-commit-config.yaml .secrets.baseline .gitignore 2>/dev/null
```

If `.pre-commit-config.yaml` is missing, warn the user and offer to set it up.

### Step 2: Scan for Secrets

Run gitleaks scan (if available):

```bash
# Check if gitleaks is installed
which gitleaks && gitleaks detect --no-git -v
```

If gitleaks is not installed, suggest:
- `brew install gitleaks` (macOS)
- `pip install detect-secrets` (alternative)

### Step 3: Check .gitignore

Verify critical patterns are in .gitignore:

```bash
grep -E "\.env|\.key|credentials|secret" .gitignore
```

If patterns are missing, suggest adding them.

### Step 4: Report

Provide a summary:

1. **Protection Status:**
   - pre-commit hooks: installed / missing
   - .secrets.baseline: present / missing
   - .gitignore patterns: complete / incomplete

2. **Scan Results:**
   - Secrets found: count and locations
   - False positives to baseline

3. **Recommendations:**
   - If secrets found: remediation steps
   - If protection missing: setup commands

## Output Format

```markdown
## Secrets Check Report

### Protection Status
- Pre-commit hooks: [✅ Installed / ❌ Missing]
- Secrets baseline: [✅ Present / ❌ Missing]
- .gitignore patterns: [✅ Complete / ⚠️ Partial / ❌ Missing]

### Scan Results
[Results from gitleaks or detect-secrets]

### Recommendations
[Specific actions to take]
```

## If Secrets Found

If any secrets are detected:

1. **DO NOT commit them** — stop immediately
2. **Identify the secret type** (API key, password, etc.)
3. **Suggest fix:**
   - Move to `.env` file
   - Use environment variable
   - If false positive: update baseline

4. **If already in git history:**
   - Rotate the credential immediately
   - Consider git filter-branch or BFG
   - Warn about exposure
