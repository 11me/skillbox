# Linting Configuration (golangci-lint v2)

Production-ready golangci-lint v2 setup as AI quality gate.

## Installation

```bash
# Pin version for reproducibility (recommended)
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.1
golangci-lint --version
```

## Commands

| Command | Purpose |
|---------|---------|
| `golangci-lint fmt` | Format only |
| `golangci-lint run ./...` | Lint (no format) |
| `golangci-lint run --fix ./...` | Lint + autofix |

**Important:** `run` does NOT format files. Use `fmt` for formatting.

## v2 Schema

v2 config **must** start with `version: "2"`. Key differences from v1:

| v1 | v2 |
|----|-----|
| `disable-all: true` | `linters.default: none` |
| formatters in `linters.enable` | `formatters.enable` (separate section) |
| `goerr113` | `err113` |
| `typecheck` in enable | built-in, remove from enable |
| `issues.exclude-rules` | `linters.exclusions.rules` |

**Note:** `wsl` keeps the same name in v2 (not `wsl_v5`).

## Configuration Template

Use `.golangci.yml` in project root. Template: `templates/.golangci.yml`

## Formatters

In v2, formatters are separate from linters:

```yaml
formatters:
  enable:
    - gofumpt       # Stricter gofmt
    - goimports     # Import ordering
    # - gci         # Import section grouping (optional)
  exclusions:
    generated: strict
```

## Linters

### Core Correctness

```yaml
linters:
  default: none
  enable:
    - govet         # Go vet checks
    - staticcheck   # Static analysis
    # typecheck is built-in in v2, cannot be enabled/disabled
    - errcheck      # Unchecked errors
    - unused        # Unused code (replaces deadcode, structcheck, varcheck)
```

### Bugs / Resources / Context

```yaml
    - bodyclose       # HTTP response body not closed
    - sqlclosecheck   # SQL rows/stmt close
    - noctx           # Context in HTTP requests
    - contextcheck    # Context propagation
    - containedctx    # Context in structs (anti-pattern)
    - errorlint       # Error wrapping issues
    - nilerr          # Returning nil instead of error
```

### Security / Error Style

```yaml
    - gosec           # Security issues
    - err113          # Error wrapping (was goerr113)
    - errname         # Error naming conventions
```

### Maintainability

```yaml
    - unparam         # Unused parameters
    - wastedassign    # Wasted assignments
    - ineffassign     # Ineffective assignments
```

### AI Quality

```yaml
    - dupword         # Duplicate words in comments
    - misspell        # Spelling mistakes
    - nolintlint      # Anti-cheat (see below)
```

### Style

```yaml
    - revive          # Comprehensive linter
    - wsl             # Whitespace linter
    - whitespace      # Trailing whitespace
```

## nolintlint Anti-Cheat

Prevents AI from hiding issues with `//nolint`:

```yaml
linters:
  settings:
    nolintlint:
      allow-unused: true
      require-specific: true      # Must specify linter
      require-explanation: true   # Must explain why
```

### Correct Usage

```go
// BAD - will fail lint
//nolint

// BAD - no explanation
//nolint:gosec

// GOOD - specific linter + explanation
//nolint:gosec // acceptable risk: user input validated upstream
```

## revive Configuration

```yaml
linters:
  settings:
    revive:
      ignore-generated-header: true
      severity: error
      enable-all-rules: true
      confidence: 0.8
      rules:
        # Limits
        - name: argument-limit
          arguments: [4]
        - name: function-result-limit
          arguments: [3]
        - name: function-length
          arguments: [80, 0]
        - name: line-length-limit
          arguments: [150]
        - name: cognitive-complexity
          arguments: [50]
        - name: cyclomatic
          arguments: [50]

        # Naming (ID not Id)
        - name: var-naming
          arguments:
            - ["ID"]  # AllowList
            - ["VM"]  # DenyList

        # Security
        - name: imports-blacklist
          arguments:
            - "crypto/md5"
            - "crypto/sha1"

        # Disabled - too strict
        - name: add-constant
          disabled: true
        - name: file-header
          disabled: true
        - name: max-public-structs
          disabled: true
```

## Exclusions (v2 Path)

```yaml
linters:
  exclusions:
    generated: strict
    warn-unused: true
    presets:
      - std-error-handling
      - common-false-positives

    rules:
      # Test files
      - path: "_test\\.go"
        linters:
          - err113
          - gosec
          - containedctx

      # Main function
      - path: "(^|/)main\\.go$"
        linters:
          - revive

      # Generated code
      - path: "zz_generated"
        linters:
          - all
```

## Removed Linters (v2)

These linters were **removed** from golangci-lint v2:

| Removed | Replacement |
|---------|-------------|
| `deadcode` | `unused` |
| `structcheck` | `unused` |
| `varcheck` | `unused` |
| `goerr113` | `err113` |
| `golint` | `revive` |
| `interfacer` | removed |
| `maligned` | `fieldalignment` |
| `scopelint` | `exportloopref` |
| `typecheck` | built-in (cannot enable/disable) |

## Duplicated (Don't Enable)

| Linter | Covered by |
|--------|------------|
| `cyclop` | `revive:cyclomatic` |
| `gocyclo` | `revive:cyclomatic` |
| `gocognit` | `revive:cognitive-complexity` |
| `lll` | `revive:line-length-limit` |
| `nlreturn` | `wsl` |

## Too Strict (Optional)

| Linter | Issue |
|--------|-------|
| `exhaustruct` | Requires all struct fields |
| `wrapcheck` | Too noisy for internal code |
| `ireturn` | Blocks returning interfaces |
| `varnamelen` | Too pedantic |
| `gomnd` | Magic number detection (noisy) |

## AI Agent Workflow

After any Go code changes:

```bash
# 1. Format
golangci-lint fmt

# 2. Test
go test ./...

# 3. Lint
golangci-lint run ./...
```

## Makefile Integration

```makefile
fmt: ## Format code
	golangci-lint fmt

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Install: curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.1" && exit 1)
	golangci-lint run ./...

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix ./...
```

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `typecheck is not a linter` | Built-in in v2 | Remove from enable |
| `unknown linters: 'wsl_v5'` | Wrong name | Use `wsl` |
| `unknown linters: 'goerr113'` | Renamed in v2 | Use `err113` |

### Verify linter names

```bash
golangci-lint help linters | grep -E "^[a-z]"
```

## err113 Migration Warning

`err113` requires refactoring all dynamic errors:

```go
// Before (will fail err113)
fmt.Errorf("invalid: %s", x)
err == SomeError

// After (correct)
fmt.Errorf("invalid: %w", ErrInvalid)
errors.Is(err, SomeError)
```

For legacy projects, consider temporary disable:

```yaml
linters:
  exclusions:
    rules:
      - linters: [err113]
        text: "do not define dynamic errors"  # TODO: refactor
```

## CI Integration

### GitHub Actions

```yaml
- uses: golangci/golangci-lint-action@v7
  with:
    version: v2.7.1
```

### GitLab CI

```yaml
lint:
  script:
    - curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b /usr/local/bin v2.7.1
    - golangci-lint run ./...
```
