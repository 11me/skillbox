---
name: go-review
description: Review Go project against production standards from go-development skill
---

# /go-review

Review Go code for compliance with go-development patterns and conventions.

## Usage

```bash
/go-review              # Review entire project
/go-review internal/    # Review specific directory
```

## What Gets Checked

### Structure
- No forbidden packages (`common`, `helpers`, `utils`, `shared`, `misc`)
- Minimal `cmd/*/main.go` (entry point only)
- Correct `internal/` organization

### Patterns
- Config: `caarlos0/env` (not viper)
- Database: `pgx` + `squirrel`
- No DI containers (wire, dig, fx)
- Service Registry pattern
- Transaction handling with `WithTx`
- Advisory locks with `Serialize()`
- Filter pattern for queries
- Sentinel errors + wrap

### Naming
- `userID` not `userId`
- `any` not `interface{}`

### Anti-patterns
- God objects (files > 500 lines)
- Circular imports

## Output

Brief report with:
- Issues found (categorized)
- Specific file:line references
- Actionable recommendations

## Example Output

```markdown
# Go Code Review

## Issues Found

### Structure
- ❌ `internal/helpers/` — forbidden package name

### Patterns
- ⚠️ Using viper instead of caarlos0/env
- ⚠️ Missing Serialize() in UserRepository

### Naming
- ❌ `userId` → `userID` in models/user.go:12

## Recommendations

1. Rename internal/helpers/ to internal/convert/
2. Migrate config from viper to caarlos0/env
3. Add Serialize() for serializable transactions
```

## Notes

- Does NOT replace `golangci-lint` — run both
- Focuses on architectural patterns, not style
- Based on go-development skill standards
