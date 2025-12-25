# Full Skill Structure Template

Use this template for skills that integrate into a skillbox plugin with commands and hooks.

## When to Use

- Skill needs slash commands (/command)
- Skill needs event hooks (session start, tool use, etc.)
- Skill is part of a larger plugin ecosystem
- Team needs standardized workflows

## Template Structure

```
plugins/skillbox/
├── skills/
│   └── {{skill-name}}/
│       ├── SKILL.md              # Main skill instructions
│       ├── REFERENCE.md          # Detailed documentation
│       └── snippets/             # Code templates
│           └── example.yaml
├── commands/
│   ├── {{skill-command-1}}.md    # Slash command
│   └── {{skill-command-2}}.md    # Another command
├── hooks/
│   └── hooks.json                # Hook definitions
└── scripts/
    └── hooks/
        └── {{hook-script}}.sh    # Hook implementation
```

## SKILL.md Template

```yaml
---
name: {{skill-name}}
description: {{What this skill does}}. Use when {{trigger scenarios}}.
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# {{Skill Name}}

## Purpose / When to Use

Use this skill when:
- {{Scenario 1}}
- {{Scenario 2}}
- {{Scenario 3}}

## Commands

| Command | Description |
|---------|-------------|
| `/{{command-1}}` | {{What it does}} |
| `/{{command-2}}` | {{What it does}} |

## Workflow

### 1. {{Step 1}}

{{Instructions}}

Run `/{{command}}` to {{action}}.

### 2. {{Step 2}}

{{Instructions}}

## Definition of Done

1. [ ] {{Validation 1}}
2. [ ] {{Validation 2}}

Run `/{{validate-command}}` to verify all checks.

## Related Files

- [REFERENCE.md](REFERENCE.md) — Detailed reference
- [snippets/](snippets/) — Code templates

## Version History

- 1.0.0 — Initial release
```

## Command Template

Commands are stored in `plugins/skillbox/commands/{{command-name}}.md`:

```markdown
---
description: {{What the command does}}
---

# /{{command-name}}

{{Detailed instructions for Claude when user runs this command}}

## Usage

```
/{{command-name}} [arguments]
```

## Steps

1. {{Step 1}}
2. {{Step 2}}
3. {{Step 3}}

## Example

User: `/{{command-name}} my-app`

{{What Claude should do}}
```

## Hooks Configuration

Hooks are defined in `plugins/skillbox/hooks/hooks.json`:

```json
{
  "hooks": [
    {
      "name": "{{hook-name}}",
      "event": "{{event-type}}",
      "script": "scripts/hooks/{{script-name}}.sh",
      "description": "{{What hook does}}"
    }
  ]
}
```

### Available Events

| Event | When Triggered |
|-------|---------------|
| `SessionStart` | When Claude Code session begins |
| `UserPromptSubmit` | Before user prompt is processed |
| `PreToolUse` | Before a tool is executed |
| `PostToolUse` | After a tool execution |
| `Stop` | When session ends |

## Hook Script Template

Scripts are stored in `plugins/skillbox/scripts/hooks/`:

```bash
#!/bin/bash

# {{hook-name}}.sh
# {{Description of what this hook does}}

# Input from Claude Code
INPUT=$(cat)

# Parse input (JSON format)
# EVENT_DATA=$(echo "$INPUT" | jq -r '.eventData')

# Your logic here
# ...

# Output (optional - can modify behavior)
echo '{"allow": true}'
```

## Example: Deployment Skill with Commands

```
plugins/skillbox/
├── skills/
│   └── k8s-deployer/
│       ├── SKILL.md
│       ├── REFERENCE.md
│       └── snippets/
│           ├── deployment.yaml
│           └── service.yaml
├── commands/
│   ├── deploy.md
│   └── rollback.md
├── hooks/
│   └── hooks.json
└── scripts/
    └── hooks/
        └── pre-deploy-check.sh
```

### commands/deploy.md

```markdown
---
description: Deploy application to Kubernetes cluster
---

# /deploy

Deploy the application to the specified environment.

## Usage

```
/deploy <app-name> <environment>
```

## Steps

1. Validate deployment manifests
2. Check cluster connectivity
3. Apply manifests with `kubectl apply`
4. Verify rollout status
5. Report deployment result

## Example

User: `/deploy my-app production`

This will deploy my-app to the production cluster.
```

### hooks/hooks.json

```json
{
  "hooks": [
    {
      "name": "k8s-pre-deploy",
      "event": "PreToolUse",
      "script": "scripts/hooks/pre-deploy-check.sh",
      "description": "Validate manifests before kubectl apply"
    }
  ]
}
```

### scripts/hooks/pre-deploy-check.sh

```bash
#!/bin/bash

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool')
COMMAND=$(echo "$INPUT" | jq -r '.input.command // empty')

# Only check kubectl apply commands
if [[ "$TOOL" == "Bash" && "$COMMAND" == *"kubectl apply"* ]]; then
    # Validate manifests exist
    if [[ ! -f "manifests/deployment.yaml" ]]; then
        echo '{"allow": false, "reason": "No deployment manifest found"}'
        exit 0
    fi
fi

echo '{"allow": true}'
```

## Best Practices

1. **Commands are user-invoked**: Use for explicit actions (/deploy, /validate)
2. **Skills are model-invoked**: Use for context-aware assistance
3. **Hooks are automatic**: Use for guardrails and session setup
4. **Keep commands focused**: One command = one action
5. **Document command arguments**: Show usage examples
