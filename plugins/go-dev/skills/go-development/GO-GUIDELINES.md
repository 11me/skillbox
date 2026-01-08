# Go Code Guidelines (MANDATORY)

These rules are **enforced by linting**. Violations will fail the build.

## Naming

| Rule | ✅ Correct | ❌ Wrong |
|------|-----------|----------|
| Acronyms ALL CAPS | `userID`, `httpURL`, `xmlParser` | `userId`, `httpUrl`, `xmlparser` |
| Use `any` | `func Foo(v any)` | `func Foo(v interface{})` |
| Getters no "Get" | `user.Name()` | `user.GetName()` |

## Packages

**FORBIDDEN names:** `helpers`, `utils`, `common`, `shared`, `misc`, `base`, `types`

Name packages by what they provide: `optional/`, `retry/`, `httputil/`

## Errors

**Format:** `"<op>: <context>: %w"` with `%w` at END

```go
// ✅ Correct
fmt.Errorf("UserService.Get userID=%s: %w", id, err)

// ❌ Wrong
fmt.Errorf("error getting user: %w", err)
fmt.Errorf("failed: %w: more context", err)  // %w not at end
```

## Comments

- ❌ NO decorative: `// ======`, `// ------`, `// ****`, `// ####`
- ❌ NO empty line-fillers or section separators
- ✅ Start with symbol name: `// User represents a registered user.`

## Structure

```
cmd/app/main.go         # Entry point ONLY
internal/               # Private code
  config/               # caarlos0/env config
  errs/                 # Sentinel errors
  models/               # Domain models + mappers
  services/             # Business logic + registry
  storage/              # Repository implementations
  http/v1/              # Handlers + router
pkg/                    # Public reusable code
```

## Quick Reference

→ Full documentation: `skills/go/go-development/references/`
→ Examples: `skills/go/go-development/examples/`
→ Linting: Run `golangci-lint run ./...` after changes
