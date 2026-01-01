---
name: ansible-validate
description: Run lint and security checks on Ansible project
allowed-tools: Read, Glob, Grep, Bash
---

# /ansible-validate

Validate Ansible project with linting, security checks, and best practice verification.

## Workflow

### Step 1: Detect Project Structure

Verify Ansible project exists:

```bash
# Check for ansible.cfg or playbooks
ls ansible.cfg 2>/dev/null || ls playbooks/*.yml 2>/dev/null
```

If not found, report error and exit.

### Step 2: Run ansible-lint

```bash
ansible-lint --strict 2>&1 || true
```

Capture output and categorize issues:
- **Errors** - Must fix before deployment
- **Warnings** - Should fix for best practices

### Step 3: Run yamllint

```bash
yamllint . 2>&1 || true
```

Report YAML syntax issues.

### Step 4: Security Checks

#### 4.1 Check for Plain Text Secrets

Search for potential secrets in plain text:

```bash
grep -rn --include="*.yml" --include="*.yaml" \
  -E "(password|secret|token|key|api_key|apikey):\s*['\"]?[^{]" \
  --exclude-dir=".git" \
  --exclude-dir="molecule" \
  . 2>/dev/null || true
```

Exclude vault-encrypted files (check for `$ANSIBLE_VAULT` header).

#### 4.2 Check for Missing no_log

Find tasks with sensitive operations without `no_log: true`:

```yaml
# Tasks that should have no_log:
# - Creating users with passwords
# - Setting secrets in environment
# - API calls with tokens
```

Search patterns:
- `password:` without `no_log: true` nearby
- `api_key:` or `token:` in uri/shell tasks

#### 4.3 Check SSH Configuration (if baseline role exists)

Verify secure defaults in role:
- `PermitRootLogin: "no"`
- `PasswordAuthentication: "no"`

### Step 5: Best Practice Checks

#### 5.1 FQCN Usage

Check for non-FQCN module names:

```bash
grep -rn --include="*.yml" --include="*.yaml" \
  -E "^\s+- (apt|yum|copy|template|file|service|user|group|command|shell):" \
  --exclude-dir=".git" \
  . 2>/dev/null || true
```

Recommend using `ansible.builtin.*` prefix.

#### 5.2 Handler Usage

Check for inline service restarts (should use handlers):

```bash
grep -rn --include="*.yml" --include="*.yaml" \
  "state: restarted" \
  --exclude="handlers/" \
  --exclude-dir=".git" \
  . 2>/dev/null || true
```

#### 5.3 Variable Naming

Check role variables follow naming convention:

```bash
# Variables should be prefixed with role_name__
grep -rn --include="main.yml" \
  -E "^[a-z_]+:" \
  roles/*/defaults/ 2>/dev/null || true
```

### Step 6: Generate Report

Produce structured report:

```
=== Ansible Project Validation Report ===

Project: /path/to/project
Date: YYYY-MM-DD

## Linting

### ansible-lint
[X] errors | [X] warnings

### yamllint
[X] errors | [X] warnings

## Security

### Plain Text Secrets
[STATUS] No secrets found / [X] potential secrets

### no_log Coverage
[STATUS] All sensitive tasks covered / [X] missing no_log

### SSH Hardening
[STATUS] Secure defaults / [X] insecure settings

## Best Practices

### FQCN Usage
[STATUS] All modules use FQCN / [X] legacy module names

### Handlers
[STATUS] Proper handler usage / [X] inline restarts

### Variable Naming
[STATUS] Consistent naming / [X] unprefixed variables

## Summary

Critical: [X]
Warnings: [X]
Passed: [X]

## Recommended Actions

1. [Priority] Fix critical issues
2. [Medium] Address warnings
3. [Low] Improve best practices
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `--fix` | No | Auto-fix with `ansible-lint --fix` |
| `--strict` | No | Treat warnings as errors |

## Example Usage

```
/ansible-validate
```

Runs full validation suite on current directory.

```
/ansible-validate --fix
```

Runs validation and auto-fixes what's possible.

## Exit Codes

- **0** - All checks passed
- **1** - Warnings found
- **2** - Errors found (critical issues)
