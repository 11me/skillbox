---
description: Initialize a new Go project with modern structure
allowed-tools: Bash(go:*), Bash(task:*), Bash(mkdir:*), Bash(ls:*), Write, Edit, Read
argument-hint: <project-name> [cli|http|library]
---

# Initialize Go Project

## Context

- Current directory: !`pwd`
- Go version: !`go version 2>/dev/null || echo "Go not installed"`

## Task

Create a new Go project:
- **Project name:** $1
- **Project type:** $2 (default: cli)

### Project Types
- `cli` — Command-line application with flag handling
- `http` — HTTP service with graceful shutdown
- `library` — Reusable package with examples

### Structure by Type

**cli:**
```
cmd/<name>/main.go
internal/app/app.go
```

**http:**
```
cmd/server/main.go
internal/handler/handler.go
internal/service/service.go
```

**library:**
```
pkg/<name>/<name>.go
examples/basic/main.go
```

### Config Files

Use templates from `${CLAUDE_PLUGIN_ROOT}/templates/go/`:
- `Taskfile.yml` — Build automation
- `.golangci.yml` — Linter config
- `.gitignore`

### Verify

```bash
go mod init <module-path>
go mod tidy
task --list
```

Use `go/project-init` agent if complex scaffolding needed.
