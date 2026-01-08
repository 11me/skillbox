---
name: commit
description: Create a git commit with a properly formatted Conventional Commits message
argument-hint: "[--emoji] [--amend] [hint]"
---

# /commit

Create a git commit with a properly formatted Conventional Commits message.

## Usage

```
/commit [options] [hint]
```

### Options

| Flag | Description |
|------|-------------|
| `--amend` | Amend the previous commit |
| `--emoji` | Add gitmoji emoji before type |

### Examples

```
/commit
/commit --emoji
/commit --amend
/commit fix login issue
/commit --emoji add dark mode
```

## Workflow

### Step 1: Check Repository State

```bash
git status
```

Verify:
- We're in a git repository
- There are staged changes (or unstaged if nothing staged)

### Step 2: Analyze Changes

```bash
git diff --staged
```

If nothing staged, show unstaged diff:
```bash
git diff
```

Identify:
- What files changed
- What type of change (feature, fix, refactor, etc.)
- What component/scope is affected
- Whether this is a breaking change

### Step 3: Generate Commit Message

Based on analysis, create message following Conventional Commits:

```
<type>(<scope>): <subject>

[body if needed]

[footer if needed]
```

If `--emoji` flag is present, prepend gitmoji:

| Type | Emoji |
|------|-------|
| feat | :sparkles: |
| fix | :bug: |
| docs | :memo: |
| style | :lipstick: |
| refactor | :recycle: |
| perf | :zap: |
| test | :white_check_mark: |
| build | :package: |
| ci | :construction_worker: |
| chore | :wrench: |
| revert | :rewind: |

### Step 4: Execute Commit

Standard commit:
```bash
git commit -m "$(cat <<'EOF'
<commit message here>
EOF
)"
```

With amend:
```bash
git commit --amend -m "$(cat <<'EOF'
<commit message here>
EOF
)"
```

### Step 5: Confirm Success

```bash
git log -1 --oneline
```

Show the created commit to confirm.

## Rules

1. **Subject line**: Imperative mood, max 50 chars, no period
2. **Scope**: Derive from changed files (api, ui, config, etc.)
3. **Body**: Explain what and why, not how
4. **Breaking changes**: Use `feat(scope)!:` pattern, add `BREAKING CHANGE:` footer
5. **References**: Include issue refs in footer (Fixes #123)

## Message Format Examples

### Simple:
```
feat(auth): add password reset functionality
```

### With body:
```
fix(api): handle null response from payment gateway

The payment service occasionally returns null instead of an error.
Added defensive null checking and proper error propagation.

Fixes #234
```

### Breaking change:
```
feat(config)!: migrate to YAML configuration

BREAKING CHANGE: JSON config files are no longer supported.
Run `migrate-config` to convert existing files.
```

### With emoji:
```
:sparkles: feat(ui): add dark mode toggle
```

## When User Provides Hint

If user provides a hint like `/commit fix login bug`:
- Use the hint to guide type selection ("fix")
- Include relevant keywords in subject
- Still analyze diff for accurate scope

## Error Handling

If no staged changes:
1. Show `git status`
2. Ask user to stage changes first: `git add <files>`
3. Or offer to stage all: `git add -A`

If not a git repository:
1. Inform user
2. Suggest `git init` if appropriate
