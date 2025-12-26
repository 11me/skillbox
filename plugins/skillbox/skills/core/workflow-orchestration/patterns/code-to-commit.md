# Code to Commit Pattern

## Overview

How to create traceable commits that link to tasks.

## The Flow

```
code changes
    ↓
stage files
    ↓
write conventional commit
    ↓
include task reference
```

## Commit Message Format

```
<type>(<scope>): <description>

[body with details]

Refs: BEADS-<id>
```

### Example

```bash
git commit -m "feat(api): add pagination to user list

- Add PaginationParams type
- Update getUsers to accept offset/limit
- Add total count to response

Refs: BEADS-77"
```

## Task Reference Placement

**Footer (preferred):**
```
Refs: BEADS-77
```

**Body (alternative):**
```
Implements BEADS-77: Add pagination to user list API
```

**Branch name (optional):**
```
feature/BEADS-77-pagination
```

## Multiple Tasks

For commits spanning multiple tasks:

```
Refs: BEADS-77, BEADS-78
```

Or:
```
Refs: BEADS-77
Fixes: BEADS-80
```

## Best Practices

1. **One task per commit** when possible
2. **Reference parent** for subtask commits
3. **Update memory** if commit changes architecture
4. **Keep scope tight** — don't mix unrelated changes

## Automation Opportunities

The conventional-commit skill can auto-suggest references:

```
# If working on BEADS-77, commit will suggest:
feat(api): your message

Refs: BEADS-77
```
