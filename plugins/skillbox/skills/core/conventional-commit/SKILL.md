---
name: conventional-commit
description: Generate beautiful git commit messages following Conventional Commits spec. Use when writing commits, creating commit messages, or when user mentions git commit, conventional commits, or commit message.
allowed-tools: Read, Grep, Glob, Bash
---

# Conventional Commit

## Purpose / When to Use

Use this skill when:
- Writing git commit messages
- User asks to create a commit
- User mentions "conventional commits" or commit formatting
- Reviewing staged changes before committing

## Workflow

### Step 1: Analyze Changes

Run these commands to understand what changed:

```bash
git status
git diff --staged
```

If nothing is staged, check unstaged changes:
```bash
git diff
```

### Step 2: Determine Commit Type

| Type | When to Use |
|------|-------------|
| `feat` | New feature for the user |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, whitespace (no code change) |
| `refactor` | Code restructuring (no feature/fix) |
| `perf` | Performance improvement |
| `test` | Adding/fixing tests |
| `build` | Build system, dependencies |
| `ci` | CI/CD configuration |
| `chore` | Maintenance, tooling |
| `revert` | Reverting previous commit |

### Step 3: Detect Scope

Analyze changed file paths to suggest scope:

| File Pattern | Suggested Scope |
|--------------|-----------------|
| `src/api/*`, `api/*` | `api` |
| `src/components/*`, `components/*` | `ui` |
| `src/services/*` | `services` |
| `src/utils/*`, `lib/*` | `utils` |
| `tests/*`, `*_test.*`, `*.test.*` | `test` |
| `*.config.*`, `config/*` | `config` |
| `docs/*`, `*.md` | `docs` |
| `cmd/*` | `cli` |
| `internal/*` | `internal` |
| `pkg/*` | `pkg` |

If files span multiple directories, use the most specific common component or omit scope.

### Step 4: Write Commit Message

Format:
```
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

Rules:
- **Subject**: imperative mood ("add" not "added"), max 50 chars, no period
- **Body**: explain what and why (not how), wrap at 72 chars
- **Footer**: references, breaking changes

### Step 5: Handle Breaking Changes

If changes break backward compatibility:

1. Add `!` after type/scope: `feat(api)!: remove deprecated endpoint`
2. Add footer: `BREAKING CHANGE: description of what breaks`

Patterns that indicate breaking changes:
- Removing public API methods/endpoints
- Changing function signatures
- Removing configuration options
- Changing default behavior
- Database schema changes requiring migration

## Gitmoji (Optional)

Only add emoji when user explicitly requests with `--emoji` flag or mentions emoji:

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

Format with emoji: `<emoji> <type>(<scope>): <subject>`

## Examples

### Simple feature:
```
feat(auth): add password reset functionality
```

### Bug fix with body:
```
fix(api): handle null response from external service

The payment gateway occasionally returns null instead of an error object.
Added null check and proper error handling.

Fixes #234
```

### Breaking change:
```
feat(config)!: change config file format to YAML

Migrate configuration from JSON to YAML for better readability
and support for comments.

BREAKING CHANGE: config.json is no longer supported.
Run `migrate-config` to convert existing configuration.
```

### With emoji:
```
:sparkles: feat(ui): add dark mode toggle
```

## Definition of Done

Before committing:
1. [ ] Changes are staged (`git add`)
2. [ ] Type accurately reflects the change
3. [ ] Scope matches affected component
4. [ ] Subject is imperative, under 50 chars
5. [ ] Breaking changes are documented

## Related Files

- [REFERENCE.md](REFERENCE.md) - Full Conventional Commits specification

## Version History

- 1.0.0 - Initial release
