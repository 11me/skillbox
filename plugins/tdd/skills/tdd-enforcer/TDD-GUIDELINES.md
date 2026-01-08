# TDD Guidelines (Active)

## Cycle: RED → GREEN → REFACTOR

### RED (Write Failing Test)
1. Write ONE test for ONE behavior
2. Test MUST fail before implementation
3. Fail for RIGHT reason (not syntax error)
4. Commit: `test: add test for <behavior>`

### GREEN (Minimal Implementation)
1. Write ONLY enough code to pass
2. Don't anticipate future requirements
3. Ugly code is OK at this stage
4. Commit: `feat: implement <behavior>`

### REFACTOR (Improve Safely)
1. Tests must pass after EVERY change
2. If tests fail → undo immediately
3. Extract methods, improve names, remove duplication
4. Commit: `refactor: improve <aspect>`

## Test Commands

| Language | Command |
|----------|---------|
| Go | `go test ./...` |
| TypeScript | `pnpm vitest run` |
| Python | `pytest` |
| Rust | `cargo test` |

## Guardrails

- Write failing test FIRST
- NEVER skip refactor phase
- Run tests after EVERY change
- One test = one behavior

→ Full reference: `skills/core/tdd-enforcer/SKILL.md`
