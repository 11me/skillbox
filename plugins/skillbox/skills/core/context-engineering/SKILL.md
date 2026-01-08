---
name: context-engineering
description: Manage AI agent context windows effectively. Use when dealing with context overflow, token limits, context budget, progressive loading, or optimizing long sessions.
allowed-tools: [Read, Grep, Glob]
---

# Context Engineering — Maximizing AI Agent Effectiveness

## Purpose / When to Use

Use this skill when:
- Context window is getting full (responses becoming shorter/vaguer)
- Planning long, complex tasks that may overflow context
- Need to optimize token usage
- Forgetting earlier decisions or repeating work
- Tool output truncation warnings appear

## The Problem

AI agents have finite context windows. Poor context management leads to:
- Hallucinations from missing information
- Task failures from context overflow
- Inefficient token usage
- Lost important details during long sessions

## Context Hierarchy

```
┌─────────────────────────────────────────────┐
│           CONTEXT HIERARCHY                  │
├─────────────────────────────────────────────┤
│  1. System Instructions (fixed)             │
│  2. Project CLAUDE.md (loaded at start)     │
│  3. Convention Skills (auto-activated)      │
│  4. Memory / Checkpoints (on-demand)        │
│  5. Task Context (injected)                 │
│  6. Tool Outputs (dynamic)                  │
│  7. Conversation History (accumulated)      │
└─────────────────────────────────────────────┘
```

## Context Budget Management

### Ideal Distribution

| Category | Budget | Purpose |
|----------|--------|---------|
| System + Instructions | ~15% | Fixed overhead |
| Project Context | ~20% | CLAUDE.md, conventions |
| Working Memory | ~35% | Current task state |
| Tool Outputs | ~20% | Search results, file contents |
| Reserve | ~10% | Safety margin |

### Warning Signs of Context Overflow

- Responses getting shorter/vaguer
- Forgetting earlier decisions
- Repeating already-done work
- Tool output truncation warnings

## Context Compression Techniques

### 1. Summarize Verbose Outputs

When tool outputs are large:
```
Instead of keeping full file:
"The UserService class (300 lines) has methods:
- authenticate(email, password) → User
- register(data) → User
- resetPassword(email) → void
Key: Uses bcrypt for passwords, JWT for tokens"
```

### 2. Use TodoWrite as State Machine

```
TodoWrite([
  {content: "Research auth patterns", status: "completed"},
  {content: "Implement UserService", status: "in_progress"},
  ...
])
```
→ Maintains state without repeating details

### 3. Reference, Don't Repeat

```
Instead of: "As I mentioned, the auth uses JWT..."
Use: "Per earlier analysis, JWT is used for auth..."
```

## Progressive Loading Pattern

**DON'T:** Read entire codebase upfront
**DO:** Load context progressively

```
Step 1: Get symbols overview (structure only)
Step 2: Find specific symbols needed
Step 3: Read only relevant code bodies
Step 4: Load references as needed
```

## Optimal Tool Usage

### For Code Exploration

```
1. Glob/Grep — Find candidates
2. Read (partial) — Understand structure
3. Read (full) — Only when editing
```

### For Search

```
1. Grep with pattern — Find candidates
2. Read specific files — Confirm matches
3. Summarize findings — Compress for memory
```

## Anti-Patterns

| Anti-Pattern | Impact | Better Approach |
|--------------|--------|-----------------|
| Reading entire files | Wastes tokens | Read specific sections |
| Keeping all tool output | Overflow | Summarize and discard |
| Loading everything upfront | Early overflow | Progressive loading |
| Repeating explanations | Token waste | Reference previous work |
| No state tracking | Context loss | Use TodoWrite |

## Context Recovery Options

### Option 1: Checkpoint and Continue
```
1. Summarize completed work
2. Save state to memory/notes
3. Continue with fresh context
```

### Option 2: Compression
```
1. Summarize completed work
2. Clear unnecessary tool outputs
3. Focus on remaining tasks only
```

### Option 3: Handoff
```
1. Create comprehensive checkpoint
2. Commit pending changes
3. Report recovery instructions for next session
```

## Guardrails

**NEVER:**
- Load entire codebase upfront
- Keep full tool outputs when summaries suffice
- Repeat explanations already given

**ALWAYS:**
- Use TodoWrite to track task state
- Summarize findings before moving on
- Reference previous work instead of repeating

## Trigger Examples

Prompts that should activate this skill:
- "My context is getting full, help me manage it"
- "How do I prevent context overflow?"
- "Summarize this session for handoff"
- "What's the best way to load code progressively?"
- "I keep forgetting earlier decisions"
- "Response is getting shorter, what's happening?"

## Related Skills

- **unified-workflow** — Session persistence
- **serena-navigation** — Memory persistence with serena
- **discovery** — Managing discovery context

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
