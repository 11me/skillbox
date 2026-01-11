# Code Quality Tools

Static analysis tools for detecting issues that compilers and linters miss.

## Deadcode Analysis

The `deadcode` tool finds unreachable functions using Rapid Type Analysis (RTA).

### What It Detects

| Issue | Example |
|-------|---------|
| Unreachable functions | Function never called from main/init |
| Dead interface methods | Methods on types never instantiated |
| Orphaned code | Functions left after refactoring |

### Installation

```bash
go install golang.org/x/tools/cmd/deadcode@latest
```

### Usage

```bash
# Check all packages
deadcode ./...

# Include test binaries (important for libraries)
deadcode -test ./...

# Debug: explain why a function IS reachable
deadcode -whylive=myapp/internal.ProcessOrder ./...

# Filter output to specific packages
deadcode -filter=myapp/internal ./...
```

### Example

```go
package main

type Greeter interface{ Greet() }

type Helloer struct{}
type Goodbyer struct{}

func (Helloer) Greet()  { hello() }
func (Goodbyer) Greet() { goodbye() }

func hello()   { fmt.Println("hello") }
func goodbye() { fmt.Println("goodbye") }

func main() {
    var g Greeter
    g = Helloer{}  // Only Helloer is instantiated
    g.Greet()
}
```

Output:
```
greet.go:9: unreachable func: Goodbyer.Greet
greet.go:12: unreachable func: goodbye
```

The `Goodbyer` type is never instantiated, so its `Greet` method and `goodbye` function are unreachable.

### When to Run

| Scenario | Command |
|----------|---------|
| Before code review | `make deadcode` |
| In CI pipeline | `deadcode -test ./...` |
| After refactoring | `deadcode ./...` |
| Debugging reachability | `deadcode -whylive=pkg.Func ./...` |

### Limitations

- Cannot detect calls from assembly code
- Cannot detect `go:linkname` aliased functions
- Library packages need `-test` flag to analyze test entry points

## Golangci-lint Configuration

Use `.golangci.yml` in project root. Template: `templates/.golangci.yml`

### Recommended Linters

```yaml
linters:
  disable-all: true
  enable:
    # Core (always enable)
    - errcheck      # Unchecked errors
    - govet         # Go vet checks
    - staticcheck   # Static analysis
    - unused        # Unused code
    - gosimple      # Simplifications
    - ineffassign   # Ineffective assignments
    - typecheck     # Type checking

    # Style
    - gofmt         # Formatting
    - goimports     # Import ordering
    - misspell      # Spelling
    - unconvert     # Unnecessary conversions
    - whitespace    # Whitespace issues

    # Bugs
    - bodyclose     # HTTP body close
    - nilerr        # nil error returns
    - errorlint     # Error wrapping
    - gocritic      # Various checks
    - gosec         # Security issues

    # Complexity
    - gocyclo       # Cyclomatic complexity
    - funlen        # Function length
    - nestif        # Nested if depth
```

### Key Settings

```yaml
linters-settings:
  funlen:
    lines: 100
    statements: 80
  gocyclo:
    min-complexity: 30
  nestif:
    min-complexity: 5
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
```

### Test File Exclusions

```yaml
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - funlen
        - gocyclo
        - gosec
        - dupl
```

### Linter Categories

| Category | Linters | Purpose |
|----------|---------|---------|
| Core | errcheck, govet, staticcheck | Must-have checks |
| Style | gofmt, goimports, misspell | Code consistency |
| Bugs | bodyclose, nilerr, errorlint | Common mistakes |
| Complexity | gocyclo, funlen, nestif | Maintainability |
| Security | gosec | Vulnerability detection |

## Deadcode vs golangci-lint

| Tool | Detects |
|------|---------|
| `golangci-lint` (unused) | Unused variables, constants, types |
| `deadcode` | Unreachable functions via call graph analysis |

Use both for comprehensive coverage:

```bash
make lint      # golangci-lint
make deadcode  # deeper function reachability
```

## Recommended Workflow

1. Run `make lint` for general issues
2. Run `make deadcode` for unreachable code
3. Review reported functions:
   - Delete if truly unused
   - Keep if called via reflection/external code
4. Use `-whylive` to understand why code IS reachable
