# Skillbox Agents Registry

Autonomous agents for specialized tasks. Each agent has its own model, tools, and trigger patterns.

## Categories

### Core (Workflow)

| Agent | Model | Description | Triggers |
|-------|-------|-------------|----------|
| [task-tracker](core/task-tracker.md) | haiku | Manage beads tasks lifecycle | "implement", "fix", "track work" |
| [session-checkpoint](core/session-checkpoint.md) | haiku | Save progress to serena memory | "save progress", "checkpoint" |
| [code-navigator](core/code-navigator.md) | sonnet | Semantic code navigation | "find where X is", "explore codebase" |

### TDD

| Agent | Model | Description | Triggers |
|-------|-------|-------------|----------|
| [test-analyzer](tdd/test-analyzer.md) | sonnet | Analyze coverage, find gaps | "check coverage", "find missing tests" |
| [tdd-coach](tdd/tdd-coach.md) | sonnet | Guide Red-Green-Refactor | "TDD", "write tests first" |

### TypeScript

| Agent | Model | Description | Triggers |
|-------|-------|-------------|----------|
| [test-generator](ts/test-generator.md) | sonnet | Generate Vitest tests | "write TS tests" |
| [project-init](ts/project-init.md) | sonnet | Scaffold TS project | "create TypeScript project" |

### Go

| Agent | Model | Description | Triggers |
|-------|-------|-------------|----------|
| [test-generator](go/test-generator.md) | sonnet | Generate Go table-driven tests | "write Go tests" |
| [project-init](go/project-init.md) | sonnet | Scaffold Go project | "create Go project" |

### Python

| Agent | Model | Description | Triggers |
|-------|-------|-------------|----------|
| [test-writer](python/test-writer.md) | opus | Generate pytest tests | "write Python tests" |

### Rust

| Agent | Model | Description | Triggers |
|-------|-------|-------------|----------|
| [test-generator](rust/test-generator.md) | sonnet | Generate Rust tests | "write Rust tests" |

## Quick Lookup

### By Task

| Task | Agent |
|------|-------|
| Track implementation work | task-tracker |
| Save session progress | session-checkpoint |
| Navigate codebase semantically | code-navigator |
| Analyze test coverage | test-analyzer |
| Guide TDD workflow | tdd-coach |
| Generate TypeScript tests | ts/test-generator |
| Scaffold TypeScript project | ts/project-init |
| Generate Go tests | go/test-generator |
| Scaffold Go project | go/project-init |
| Generate Python tests | python/test-writer |
| Generate Rust tests | rust/test-generator |

### By Model

| Model | Agents |
|-------|--------|
| haiku | task-tracker, session-checkpoint |
| sonnet | code-navigator, test-analyzer, tdd-coach, ts/*, go/*, rust/* |
| opus | python/test-writer |

### By Language

| Language | Agents |
|----------|--------|
| TypeScript | ts/test-generator, ts/project-init |
| Go | go/test-generator, go/project-init |
| Python | python/test-writer |
| Rust | rust/test-generator |
| Multi-language | tdd-coach, test-analyzer |

## Agent Interactions

```
tdd-coach ←→ test-analyzer
    │           (coverage → gaps → coach)
    ↓
Language-specific generators:
    ├─► ts/test-generator
    ├─► go/test-generator
    ├─► python/test-writer
    └─► rust/test-generator

task-tracker ←→ session-checkpoint
    │              (task → checkpoint → restore)
    ↓
code-navigator
    (semantic exploration)
```

## Dependencies

### Serena MCP (required for)
- session-checkpoint
- code-navigator

### Beads CLI (required for)
- task-tracker

### Language Runtimes
- TypeScript: Node.js, pnpm
- Go: Go 1.21+
- Python: Python 3.12+, pytest
- Rust: Cargo

## Adding New Agents

See [skill-creator](../skills/core/skill-creator/) for templates. Agents use the same YAML frontmatter:

```yaml
---
name: agent-name
description: When to trigger and what it does
model: haiku|sonnet|opus
tools: [tool-list]
color: "#hex-color"
---
```
