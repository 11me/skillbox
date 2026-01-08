---
name: cancel-discover-loop
description: Cancel active discovery loop
---

# Cancel Discovery Loop

Run the cancellation script:

```bash
${CLAUDE_PLUGIN_ROOT}/scripts/discovery/cancel-loop.sh
```

This removes the state file and stops the discovery loop. Your findings are preserved in `.claude/discovery-findings.md`.
