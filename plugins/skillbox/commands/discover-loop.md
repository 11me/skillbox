---
name: discover-loop
description: Iterative discovery loop for deep research (Ralph pattern)
---

# Discovery Loop

Iterative self-questioning loop for deep research. Uses the Ralph Wiggum pattern: Stop hook intercepts session exit and re-injects the research question with accumulated findings.

## Research Question

$ARGUMENTS

## Setup

First, run the setup script to initialize the discovery loop:

```bash
${CLAUDE_PLUGIN_ROOT}/scripts/discovery/setup-loop.sh "$ARGUMENTS"
```

Parse any options from arguments:
- `--max-iterations N` (default: 10)
- `--completion-signal "TEXT"` (default: "INSIGHT_COMPLETE")

## Loop Behavior

Each iteration you will:

1. **Review previous findings** (from state file and modified files)
2. **Generate new questions** based on what you've learned
3. **Explore one question deeply**
4. **Document findings** in a structured way
5. **Assess completion** - have you found the insight?

## State File Location

`.claude/discovery-loop.local.md`

Read this file at the start of each iteration to understand:
- Current iteration number
- Previous question chain
- Accumulated findings

## Iteration Workflow

### Step 1: Read State

```bash
cat .claude/discovery-loop.local.md
```

Note the iteration number and any previous findings.

### Step 2: Generate Next Question

Based on the original research question and previous findings:
- What's the most important unanswered question?
- What assumption hasn't been challenged yet?
- What cross-domain analogy could provide insight?

### Step 3: Explore the Question

Use available tools:
- `Grep` / `Glob` - search codebase for patterns
- `Read` - examine specific implementations
- `WebSearch` - find external knowledge
- `WebFetch` - retrieve documentation

### Step 4: Document Finding

Write your finding to a file that persists across iterations:

```bash
# Append to findings file
echo "## Iteration N: [Question]" >> .claude/discovery-findings.md
echo "[Your insight]" >> .claude/discovery-findings.md
```

### Step 5: Check Completion

Have you found a significant insight that answers the research question?

If YES, output the completion signal:
```
<discovered>INSIGHT_COMPLETE</discovered>
```

If NO, the Stop hook will automatically continue to the next iteration.

## Completion Criteria

Output `<discovered>INSIGHT_COMPLETE</discovered>` when:
- You've found a novel, non-obvious insight
- The insight directly addresses the research question
- You can articulate WHY this insight is valuable
- Further questioning would yield diminishing returns

## Final Summary

When complete, provide:

```
## Discovery Complete

**Research Question:** [original question]

**Iterations:** [N]

**Key Insight:** [The main discovery]

**Evidence:** [What led to this insight]

**Implications:** [What this means for the problem]

**Next Steps:** [Concrete actions based on discovery]

<discovered>INSIGHT_COMPLETE</discovered>
```

## Cancel

To cancel the loop manually:
```bash
rm .claude/discovery-loop.local.md
```

Or use: `/cancel-discover-loop`

## Notes

- Each iteration builds on previous findings
- The same research question is re-injected each time
- Your previous work persists in files
- Maximum iterations prevent infinite loops
- Trust the process - insights often emerge after 3-5 iterations
