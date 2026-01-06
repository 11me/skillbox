---
name: discovery
description: Use when the user asks about "discovery", "self-questioning", "SP-CoT", "novel insights", "systematic exploration", "hypothesis generation", "Socratic method", "deep research", or needs guidance on AI-powered discovery through systematic self-questioning.
allowed-tools: [Read, Grep, Glob, WebSearch, WebFetch]
version: 2.0.0
---

# Self-Questioning Discovery System

AI-powered discovery through systematic self-questioning. Generates novel insights by asking the right questions before answering.

## Concept

Inspired by:
- **Self-Prompted Chain-of-Thought (SP-CoT)**: LLMs generating their own reasoning chains
- **AI Scientist (Sakana AI)**: Autonomous hypothesis generation and testing
- **Socratic Method**: Discovery through questioning assumptions

## Commands

### `/discover <problem> [--depth quick|deep]`

Quick self-questioning for problem-solving.

```bash
# Quick mode (default): 3-5 focused questions
/discover "How to optimize this API endpoint?"

# Deep mode: triggers discovery-explorer agent
/discover "Design a scalable notification system" --depth deep
```

**Output:**
1. Self-generated questions about the problem
2. Cross-domain analogies (Tech + Business/UX patterns)
3. Novel insights and recommendations

### `/discover-loop <research-question> [--max-iterations N]`

Iterative discovery using Ralph-style loops. Continues until insight found or max iterations reached.

```bash
/discover-loop "Find optimal architecture for real-time collaboration" --max-iterations 10
```

**Completion signal:** `<discovered>INSIGHT_COMPLETE</discovered>`

### `/cancel-discover-loop`

Cancel active discovery loop. Preserves findings in `.claude/discovery-findings.md`.

## Agent

### discovery-explorer

Deep exploration agent for complex problems. Automatically triggered by `/discover --depth deep`.

**Phases:**
1. **Decomposition**: Break into sub-questions
2. **Cross-Domain**: Find analogies from other fields
3. **Challenge**: Question hidden assumptions
4. **Synthesis**: Combine insights into novel solution

## Question Patterns

### Tech-Focused
- What existing pattern in [framework] solves this?
- How would this scale to 10x load?
- What's the simplest solution that could work?
- Which constraint is artificial vs real?

### Business/Design
- What would frustrate a user about this?
- How do successful products handle this?
- What's the hidden cost of this complexity?
- What would a PM ask about this?

### Socratic
- What am I assuming here?
- What would happen if the opposite were true?
- What domain has solved a similar problem?
- What would an expert in [other field] notice?

## Cross-Domain Triggers

| Keyword | Explore |
|---------|---------|
| "optimize" | Caching patterns, CDN strategies |
| "scale" | Distributed systems, sharding |
| "user flow" | UX patterns, onboarding funnels |
| "error handling" | Resilience engineering, circuit breakers |
| "auth" | OAuth flows, session patterns |
| "data" | ETL patterns, data modeling |

## Loop Mechanism (Ralph Pattern)

Uses Stop hook to create self-referential discovery loop:

1. User initiates `/discover-loop "question"`
2. Setup script creates state file `.claude/discovery-loop.local.md`
3. Claude explores, generates questions, documents findings
4. Stop hook checks for completion signal
5. If not complete -> re-injects prompt with accumulated context
6. Iterates until insight found or max iterations

**State file format:**
```yaml
---
active: true
iteration: 1
max_iterations: 10
question_chain: []
findings: []
completion_signal: "<discovered>INSIGHT_COMPLETE</discovered>"
---
[Original research question]
```

## Anti-Patterns

| Anti-Pattern | Why Bad | Do Instead |
|--------------|---------|------------|
| Jump to solution | Miss non-obvious insights | Always question first |
| Ask obvious questions | Waste thinking time | Challenge assumptions |
| Single-domain thinking | Limited perspective | Cross-domain analogies |
| Accept constraints | Miss opportunities | Question every limit |

## Examples

### Example 1: API Optimization

**Problem:** "How to optimize this API endpoint?"

**Self-Questions:**
1. What's actually slow? (assumption: whole endpoint)
2. Is caching possible? (constraint: always fresh data?)
3. What would Redis solve? (cross-domain: caching)
4. What would a DBA notice? (cross-domain: query patterns)
5. Why might optimization fail? (adversarial)

**Key Insight:** 80% of requests are for the same 20 items - cache those.

### Example 2: Discovery Loop

**Research Question:** "Find optimal data model for audit logs"

**Iteration 1:** Decompose - what are the access patterns?
**Iteration 2:** Cross-domain - how do observability tools handle this?
**Iteration 3:** Challenge - do we really need queryable logs?
**Iteration 4:** Synthesis - append-only with time-partitioning

**Key Insight:** Most audit queries are time-bounded - partition by time, not entity.

## Requirements

- `jq` - JSON processing (for transcript parsing in loop)
- `perl` - regex extraction (for completion signal)

Install on macOS: `brew install jq` (perl is pre-installed)

## Files

```
discovery/
├── SKILL.md                    # This file
├── commands/
│   ├── discover.md             # Quick discovery
│   ├── discover-loop.md        # Loop variant (Ralph pattern)
│   └── cancel-discover-loop.md # Cancel active loop
├── agents/
│   └── discovery-explorer.md   # Deep exploration agent
└── scripts/
    ├── setup-loop.sh           # Initialize loop state
    ├── cancel-loop.sh          # Cancel loop
    └── discovery-stop.sh       # Stop hook for loop continuation
```

## Related Skills

- **context-engineering** - Managing discovery context
- **unified-workflow** - Persisting discoveries in workflow
- **serena-navigation** - Exploring codebase for insights

## Guardrails

**NEVER:**
- Skip the questioning phase and jump to solutions
- Ask obvious questions that waste reasoning capacity
- Limit thinking to a single domain
- Accept all constraints without questioning

**MUST:**
- Always generate questions before answering
- Include cross-domain analogies (Tech + Business/UX)
- Synthesize findings into actionable insights
- Challenge hidden assumptions explicitly

## Version History

- 2.0.0 - Added Ralph pattern loop, discovery-explorer agent, Stop hook integration
- 1.0.0 - Initial release (quick/deep modes)
