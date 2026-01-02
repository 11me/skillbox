---
name: tdd
description: Toggle TDD mode or start TDD workflow for a feature
arguments:
  - name: action
    description: "on, off, strict, or feature name to implement"
    required: false
---

# TDD Mode Management

## Action: ${action:-status}

### If no action or "status":
Check current TDD mode status:
1. Check if `.claude/tdd-enforcer.local.md` exists
2. Report enabled/disabled and strict mode status
3. Show available commands

### If action is "on":
Enable TDD mode:
1. Create `.claude/` directory if not exists
2. Create `.claude/tdd-enforcer.local.md` with:
```yaml
---
enabled: true
strictMode: false
---

## TDD Notes

Project-specific testing notes.
```
3. Confirm: "TDD mode enabled. Run tests after code changes."

### If action is "off":
Disable TDD mode:
1. Update `.claude/tdd-enforcer.local.md` to set `enabled: false`
2. Confirm: "TDD mode disabled."

### If action is "strict":
Enable strict mode (blocks session end without tests):
1. Ensure `.claude/tdd-enforcer.local.md` exists
2. Update to set `strictMode: true`
3. Confirm: "TDD strict mode enabled. Session will block if tests not run."

### If action is a feature name:
Start TDD workflow for the feature:
1. Clarify requirements for the feature
2. Identify test file location based on project type
3. Guide through RED phase: write ONE failing test
4. Verify test fails for the right reason
5. Guide through GREEN phase: minimal implementation
6. Guide through REFACTOR phase: improve code
7. Suggest appropriate commits at each phase

## Commit Convention

- `test:` — RED phase (write failing test)
- `feat:` — GREEN phase (minimal implementation)
- `refactor:` — REFACTOR phase (improve code)

## Config File Format

`.claude/tdd-enforcer.local.md`:
```yaml
---
enabled: true
strictMode: false
---
```
