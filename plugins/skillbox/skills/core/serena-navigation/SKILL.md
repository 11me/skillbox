---
name: serena-navigation
description: Semantic code navigation with Serena MCP. Use when working with serena, semantic code navigation, find symbol, code exploration, symbol search, find references, or serena memory persistence.
globs: ["**/.serena/**"]
allowed-tools: Read, Grep, Glob
---

# Serena Navigation — Semantic Code Intelligence

## What is Serena?

Serena is a semantic code navigation MCP server that provides:
- **Symbol-level operations** — Find, read, and edit code by symbol name
- **Reference tracking** — Find all usages of a symbol across the codebase
- **Memory persistence** — Save discoveries across sessions
- **Thinking tools** — Verification checkpoints during exploration

## When to Use Serena vs Traditional Tools

| Task | Traditional | Serena |
|------|-------------|--------|
| Find class definition | `Grep "class MyClass"` | `find_symbol("MyClass")` |
| Read file structure | `Read file.py` | `get_symbols_overview("file.py")` |
| Find all usages | `Grep "method_name"` | `find_referencing_symbols("method_name")` |
| Understand architecture | Multiple greps | `list_dir` + `get_symbols_overview` + `write_memory` |

**Use Serena when:**
- Project has `.serena/` directory (serena-enabled)
- Navigating complex codebases
- Need to trace symbol references
- Want to persist discoveries across sessions

**Use Traditional tools when:**
- Searching for text patterns (not symbols)
- Reading config files, docs, non-code
- Project not serena-enabled
- Quick one-off lookups

## Core Tools

### Discovery

```python
# Get overview of symbols in a file
get_symbols_overview(relative_path="src/service.py", depth=0)

# Find symbol by name pattern
find_symbol(
    name_path_pattern="UserService",  # or "UserService/create"
    include_body=True,                 # Get source code
    depth=1                            # Include children
)

# Find all references to a symbol
find_referencing_symbols(
    name_path="UserService/create",
    relative_path="src/service.py"
)
```

### Navigation

```python
# List directory structure
list_dir(relative_path="src", recursive=True, skip_ignored_files=True)

# Search with regex (when symbol name unknown)
search_for_pattern(
    substring_pattern="TODO|FIXME",
    restrict_search_to_code_files=True
)
```

### Persistence

```python
# Save discoveries to memory
write_memory(
    memory_file_name="api-endpoints.md",
    content="# API Endpoints\n\n- GET /users..."
)

# List available memories
list_memories()

# Read a memory
read_memory(memory_file_name="api-endpoints.md")
```

### Verification

```python
# Reflect on collected information
think_about_collected_information()

# Check if still on track
think_about_task_adherence()

# Verify completion
think_about_whether_you_are_done()
```

## Workflow Patterns

### Pattern 1: Symbol Lookup
```
1. find_symbol(pattern, include_body=False)  # Find location
2. find_symbol(pattern, include_body=True)   # Get full code
3. Optional: find_referencing_symbols()      # Find usages
```

### Pattern 2: File Understanding
```
1. get_symbols_overview(file)                # See structure
2. find_symbol(specific, depth=1)            # Drill into details
3. write_memory("notes.md", findings)        # Persist
```

### Pattern 3: Architecture Exploration
```
1. list_dir(recursive=True)                  # See structure
2. get_symbols_overview() on key files       # Understand modules
3. find_referencing_symbols() for key types  # Trace dependencies
4. write_memory("architecture.md", diagram)  # Document
```

### Pattern 4: Codebase Onboarding
```
1. Check for existing memories: list_memories()
2. If "architecture.md" exists: read_memory()
3. Otherwise: explore and create memory
```

## Memory Best Practices

### Good Memory Names
- `architecture-overview.md` — High-level structure
- `api-endpoints.md` — REST/GraphQL routes
- `database-schema.md` — Data models
- `testing-patterns.md` — Test conventions
- `session-checkpoint-<date>.md` — Work-in-progress

### Memory Content Format
```markdown
# Title

## Summary
Brief overview of what was discovered.

## Details
- Key finding 1
- Key finding 2

## References
- `src/main.py:UserService` — Main entry point
- `src/models/user.py:User` — Data model

## Next Steps
- TODO: Investigate X
- TODO: Document Y
```

## Anti-Patterns

| Anti-Pattern | Better Approach |
|--------------|-----------------|
| `Read` entire large files | `get_symbols_overview` first |
| `Grep` for class names | `find_symbol` with pattern |
| Losing discoveries | `write_memory` for key findings |
| Skipping verification | Use `think_about_*` tools |

## Related Skills

- **beads-workflow** — Task tracking (pairs with serena memories)

## References

- [Serena GitHub](https://github.com/oraios/serena)
- [MCP Protocol](https://modelcontextprotocol.io)

## Version History

- 1.0.0 — Initial release (migrated from t3chn/skills)
