---
name: discover
description: Self-questioning discovery for novel problem-solving
argument-hint: "<problem> [--depth quick|deep]"
---

# Self-Questioning Discovery

You are about to solve a problem using **systematic self-questioning**. Before jumping to solutions, you MUST first generate questions that challenge assumptions and explore cross-domain analogies.

## Problem to Analyze

$ARGUMENTS

## Depth Detection

Analyze the problem complexity:
- **Quick** (default): Simple optimization, single-concern problems -> 3-5 questions
- **Deep**: Architecture decisions, multi-system problems -> launch discovery-explorer agent

If `--depth deep` is specified OR problem contains keywords like "architecture", "design", "scale", "system", "pattern" -> use deep mode.

## Quick Mode Workflow

### Step 1: Generate Self-Questions

Ask yourself these categories of questions:

**Assumption Check:**
- What am I assuming about this problem?
- Which constraints are real vs perceived?
- What would happen if I challenged the obvious solution?

**Cross-Domain (Tech):**
- What existing pattern in [relevant framework] solves similar problems?
- How would this scale to 10x the current load?
- What's the simplest solution that could work?

**Cross-Domain (Business/UX):**
- What would frustrate a user about the obvious approach?
- How do successful products handle this?
- What's the hidden cost of complexity here?

### Step 2: Answer Your Questions

For each question, provide a brief insight. Look for:
- Unexpected connections
- Hidden assumptions that could be challenged
- Analogies from other domains

### Step 3: Synthesize Insights

Combine your findings into:
1. **Key Insight**: The most surprising/valuable discovery
2. **Recommendation**: Concrete next step based on questioning
3. **Alternative Frame**: A different way to think about the problem

## Deep Mode

If deep mode triggered, use the Task tool to launch the discovery-explorer agent:

```
Launch discovery-explorer agent with the full problem context.
The agent will perform multi-phase exploration:
1. Decomposition into sub-questions
2. Cross-domain analogy search
3. Assumption challenging
4. Insight synthesis
```

## Output Format

```markdown
## Self-Questions Generated

1. [Question about assumptions]
2. [Question about constraints]
3. [Cross-domain question - Tech]
4. [Cross-domain question - Business/UX]
5. [Adversarial question - why might this fail?]

## Insights

### [Question 1]
[Brief answer/insight]

### [Question 2]
[Brief answer/insight]

...

## Discovery Summary

**Key Insight:** [Most valuable discovery]

**Recommendation:** [Concrete next step]

**Alternative Frame:** [Different perspective on the problem]
```

## Important

- Do NOT skip the questioning phase
- Spend more time on questions than answers
- Look for non-obvious connections
- Challenge the framing of the original problem
- If you find yourself giving a "standard" answer, ask "What would an expert in [unrelated field] notice?"
