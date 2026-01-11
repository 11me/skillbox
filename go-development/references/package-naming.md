# Package Naming

Avoid grab-bag packages like `helpers`, `common`, `utils`, `shared`, `misc`.

## Why It Matters

Go package names appear at every call site:

| Bad | Good |
|-----|------|
| `common.Optional("")` | `optional.Of("")` |
| `util.Format()` | `date.Format()` |
| `helpers.Encode(w, v)` | `json.Encode(w, v)` |

Package name should add meaning at call site.

## Forbidden Names

**Never create packages named:**
- `helpers`, `utils`, `util`
- `common`, `shared`, `misc`
- `base`, `types`, `interfaces`

## Why They're Bad

1. **No meaning at call sites** — `util.Foo` tells nothing
2. **Dumping ground** — unrelated code accumulates
3. **Dependency bloat** — pulls unrelated things
4. **Import conflicts** — forces aliases (`util2`, `commonpkg`)
5. **Bad boundaries** — often symptom of import cycles

## What To Do Instead

### Rule 1: Name by what it provides

| Bad | Good |
|-----|------|
| `common/optional.go` | `internal/optional/` |
| `helpers/json.go` | `internal/json/` or inline |
| `utils/retry.go` | `internal/retry/` |
| `shared/id.go` | `internal/id/` |

### Rule 2: Keep helpers close to usage

If a function is used in one package — keep it there (unexported):

```go
// ❌ BAD: create common/formatter.go for one function
package common
func FormatUserName(u *User) string { ... }

// ✅ GOOD: keep in the package that uses it
package users
func formatName(u *User) string { ... }  // unexported
```

### Rule 3: Tiny helpers — copy is OK

Go culture explicitly values small dependency trees. Sometimes 3 lines of code is better than a new package:

```go
// ❌ BAD: create common/ptr.go
import "myapp/internal/common"
user.Name = common.Ptr("John")

// ✅ GOOD: inline when trivial
user.Name = ptr("John")

func ptr[T any](v T) *T { return &v }  // inline in same file
```

### Rule 4: If truly shared, use internal/<purpose>

```
internal/
├── optional/   # pointer conversion
├── retry/      # retry logic
├── sqlscan/    # database scanning
├── httperr/    # HTTP error responses
├── health/     # health checks
└── metrics/    # metrics helpers
```

## Refactoring Existing Code

If you already have `common/utils`:

1. Group functions by concept
2. Split into focused packages
3. Delete the grab-bag package

Example:
```
# Before
common/
├── optional.go
├── json.go
└── retry.go

# After
internal/
├── optional/optional.go
├── json/encode.go
└── retry/retry.go
```

## Links

- [Go Blog: Package names](https://go.dev/blog/package-names)
- [Google Go Style: Package naming](https://google.github.io/styleguide/go/decisions#package-naming)
