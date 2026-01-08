---
name: secrets-guardian
description: Use when the user asks about "secrets protection", "pre-commit hooks", "gitleaks", "scan secrets", "secret detection", "credential leaks", or needs to set up repository protection from accidental secret commits.
---

# Secrets Guardian

Multi-layered protection against accidental secret commits. Critical for AI-assisted development where agents may not recognize sensitive data.

## Quick Setup

For new projects, run this setup:

```bash
# 1. Check if pre-commit is installed
which pre-commit || pip install pre-commit

# 2. Copy pre-commit config from assets or create manually
# See assets/pre-commit-secrets.yaml

# 3. Create secrets baseline
echo '{"version": "1.5.0", "results": {}}' > .secrets.baseline

# 4. Install hooks
pre-commit install
pre-commit install --hook-type pre-push

# 5. Verify .gitignore has secret patterns
# See assets/gitignore-secrets
```

## Commands

### `/secrets-check` - Scan Project

Quick scan for secrets in the current project:

```bash
# With gitleaks (fast)
gitleaks detect --no-git -v

# With detect-secrets (detailed)
detect-secrets scan --all-files
```

### Setup Protection

When asked to setup secrets protection:

1. **Check existing setup:**
```bash
ls -la .pre-commit-config.yaml .secrets.baseline .gitignore 2>/dev/null
```

2. **If .pre-commit-config.yaml missing:**
   - Add secret scanning hooks (see assets)
   - Or merge with existing config

3. **Check .gitignore for secret patterns:**
```bash
grep -E "\.env|\.key|API_KEY|secret" .gitignore
```

4. **Create .secrets.baseline:**
```bash
echo '{"version": "1.5.0", "results": {}}' > .secrets.baseline
```

5. **Install hooks:**
```bash
pre-commit install
pre-commit install --hook-type pre-push
```

6. **Optionally add CI workflow:**
   - Copy `assets/security-workflow.yaml` to `.github/workflows/`

---

## Proactive Checks

**IMPORTANT:** When working in any project, check for secret protection:

```bash
if [ ! -f .pre-commit-config.yaml ]; then
  echo "WARNING: No pre-commit config found"
fi
```

If missing, ask user: "No secrets protection found. Set it up?"

---

## Fix Leaked Secret

When secret is detected:

1. **Identify the secret type** (API key, password, private key, etc.)

2. **Suggest remediation:**
   - Move to `.env` file (ensure it's in .gitignore)
   - Use environment variable: `os.environ.get("API_KEY")`
   - For false positives: update `.secrets.baseline`

3. **If already committed:**
   - **Rotate the credential immediately**
   - Consider git history cleanup (if not pushed)
   - Warn about exposed secrets in git history

### Update Baseline

For false positives, update the baseline:

```bash
detect-secrets scan --baseline .secrets.baseline
```

---

## Pre-commit Hooks Configuration

Add these to `.pre-commit-config.yaml`:

```yaml
repos:
  # Secret scanning - gitleaks
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.21.2
    hooks:
      - id: gitleaks

  # Secret scanning - detect-secrets with baseline
  - repo: https://github.com/Yelp/detect-secrets
    rev: v1.5.0
    hooks:
      - id: detect-secrets
        args: ['--baseline', '.secrets.baseline']

  # Detect private keys
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: detect-private-key
      - id: check-added-large-files
        args: ['--maxkb=1000']
```

---

## .gitignore Patterns

Add these patterns to `.gitignore`:

```gitignore
# Environment files
.env
.env.*
!.env.example
*.env

# Secret/credential files
secrets.yaml
*-secrets.yaml
*.secret
*.secrets

# Keys and certificates
*.key
*.pem
*.p12
*.pfx
*.crt
*.cer

# API keys in filenames
*_API_KEY*
*api_key*
*apikey*

# Cloud credentials
credentials.json
service-account*.json
*.credentials
gcloud-*.json

# SSH keys
id_rsa
id_rsa.pub
id_ed25519
id_ed25519.pub
*.ppk

# AWS
.aws/credentials
aws-credentials*

# Terraform
*.tfstate
*.tfstate.*
*.tfvars
!*.tfvars.example

# Ansible
vault.yml
*-vault.yml
*.vault
```

---

## GitHub Actions Workflow

Create `.github/workflows/security.yaml`:

```yaml
name: Security Scan

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  secret-scan:
    name: Secret Scanning
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: TruffleHog
        uses: trufflesecurity/trufflehog@main
        with:
          extra_args: --only-verified
```

---

## Tools Reference

| Tool | Purpose | Speed | Install |
|------|---------|-------|---------|
| gitleaks | Git history scanning | Fast | `brew install gitleaks` |
| detect-secrets | Baseline-aware scanning | Medium | `pip install detect-secrets` |
| trufflehog | Verified secrets only | Slow | `brew install trufflehog` |

---

## Secret Types to Watch

| Type | Pattern | Risk |
|------|---------|------|
| AWS Keys | `AKIA...` | High - full account access |
| GitHub Tokens | `ghp_...`, `gho_...` | High - repo access |
| Private Keys | `-----BEGIN RSA PRIVATE K*Y-----` | Critical |
| API Keys | Various formats | Medium-High |
| Database URLs | `postgres://user:pass@` | High |
| JWT Secrets | Long random strings | High |

---

## Definition of Done

Before completing secrets protection setup:
- [ ] `.pre-commit-config.yaml` exists with gitleaks hook
- [ ] `.secrets.baseline` created (for detect-secrets)
- [ ] `.gitignore` includes all secret patterns
- [ ] `pre-commit install` completed
- [ ] Initial scan with `gitleaks detect` shows no secrets

## Guardrails

**NEVER:**
- Commit files matching secret patterns (`.env`, `*.key`, `credentials.*`)
- Bypass pre-commit hooks with `--no-verify`
- Store secrets in version control, even encrypted
- Ignore secrets found in scan output

**MUST:**
- Always check for `.pre-commit-config.yaml` in new projects
- Rotate any accidentally committed credentials immediately
- Scan before first commit in any project
- Update baseline only for verified false positives

## Related Skills

- **unified-workflow** — Includes secrets protection in setup
- **conventional-commit** — Safe commit workflow

## Version History

- 1.0.0 — Initial release
