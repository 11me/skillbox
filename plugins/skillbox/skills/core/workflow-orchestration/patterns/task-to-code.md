# Task to Code Pattern

## Overview

How to transition from task creation to code exploration.

## The Flow

```
bd create → beads task
    ↓
identify scope
    ↓
serena: find_symbol
    ↓
serena: find_referencing_symbols
    ↓
write_memory with findings
```

## Step-by-Step

### 1. Create Task with Clear Scope

```bash
bd create --title "Add pagination to user list API" -t task
# → Created BEADS-77
```

### 2. Identify Key Symbols

From the task description, identify likely code locations:
- "user list" → `UserController`, `UserService`, `getUsers`
- "pagination" → might need new types, params

### 3. Explore with Serena

```
# Find the current implementation
find_symbol "getUsers"

# Find all references to understand usage
find_referencing_symbols "UserService" "user-service.ts"

# Find existing pagination (if any)
search_for_pattern "offset|limit|page"
```

### 4. Document in Memory

```
write_memory "BEADS-77-exploration.md" """
# BEADS-77: Add pagination

## Current State
- UserController.getUsers() at src/controllers/user.ts:45
- Returns all users without limit
- Called from: AdminPanel, UserSearch

## Planned Changes
1. Add PaginationParams type
2. Modify getUsers to accept params
3. Update callers
"""
```

## Why This Matters

- **Persistence:** Memory survives context reset
- **Traceability:** Task ID in memory filename links exploration to task
- **Efficiency:** No re-exploration needed in future sessions
