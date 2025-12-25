# Basic Skill Template

Use this template for simple, focused skills that need only a single file.

## When to Use

- Skill provides one clear capability
- No external references or snippets needed
- Instructions fit in one file (< 200 lines)

## Template

```yaml
---
name: {{skill-name}}
description: {{What this skill does}}. Use when {{trigger scenarios}}.
---

# {{Skill Name}}

## Purpose / When to Use

Use this skill when:
- {{Scenario 1}}
- {{Scenario 2}}
- {{Scenario 3}}

## Instructions

### Step 1: {{First Step}}

{{Detailed instructions}}

### Step 2: {{Second Step}}

{{Detailed instructions}}

### Step 3: {{Third Step}}

{{Detailed instructions}}

## Examples

Prompts that should activate this skill:

1. "{{Example prompt 1}}"
2. "{{Example prompt 2}}"
3. "{{Example prompt 3}}"

## Version History

- 1.0.0 — Initial release
```

## Example: Git Commit Helper

```yaml
---
name: git-commit-helper
description: Generate clear, conventional commit messages from git diffs. Use when writing commit messages, reviewing staged changes, or formatting commits.
---

# Git Commit Helper

## Purpose / When to Use

Use this skill when:
- Writing commit messages for staged changes
- Reviewing what will be committed
- Formatting commits with conventional commit style

## Instructions

### Step 1: Review Changes

Run `git diff --staged` to see what will be committed.

### Step 2: Analyze Changes

Identify:
- Type of change (feat, fix, refactor, docs, test, chore)
- Affected components or files
- Purpose of the change

### Step 3: Write Message

Format:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Rules:
- Subject: imperative mood, under 50 chars
- Body: explain what and why, not how
- Footer: references to issues (Fixes #123)

## Examples

Prompts that should activate this skill:

1. "Write a commit message for my changes"
2. "Help me format this commit"
3. "What should I put in the commit message?"

## Version History

- 1.0.0 — Initial release
```

## Directory Structure

```
skill-name/
└── SKILL.md
```

No additional files needed for basic skills.
