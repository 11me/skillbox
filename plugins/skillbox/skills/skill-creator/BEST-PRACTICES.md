# Skill Authoring Best Practices

## Writing Effective Descriptions

The `description` field is critical for Claude to discover your skill. It must answer two questions:
1. **What** does this skill do?
2. **When** should Claude use it?

### Good vs Bad Descriptions

**Bad (too vague):**
```yaml
description: Helps with documents
```

**Bad (missing "when"):**
```yaml
description: Extract text and tables from PDF files
```

**Good:**
```yaml
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.
```

### Include Trigger Words

Think about what users will say when they need this skill:

```yaml
# For a database skill
description: Design PostgreSQL schemas, write migrations, optimize queries. Use when working with databases, SQL, PostgreSQL, migrations, or query optimization.
```

The words "databases", "SQL", "PostgreSQL", "migrations", "query optimization" are all triggers.

### Be Specific About Scope

```yaml
# Too broad (will conflict with other skills)
description: For data analysis

# Specific scope
description: Analyze sales data in Excel files and CRM exports. Use for sales reports, pipeline analysis, and revenue tracking in .xlsx format.
```

## When to Use allowed-tools

### Restrict Access For:

1. **Read-only skills** (code review, documentation lookup):
```yaml
allowed-tools: Read, Grep, Glob
```

2. **Analysis skills** (no file modifications):
```yaml
allowed-tools: Read, Grep, Glob, Bash
```

3. **Focused write skills** (only specific operations):
```yaml
allowed-tools: Read, Write, Edit
```

### Full Access When:

- Skill needs complete flexibility
- Operations vary based on context
- You trust Claude's judgment for the domain

```yaml
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
```

### Available Tools

| Tool | Purpose |
|------|---------|
| `Read` | Read file contents |
| `Write` | Create new files |
| `Edit` | Modify existing files |
| `Grep` | Search file contents |
| `Glob` | Find files by pattern |
| `Bash` | Execute shell commands |

## Structuring Instructions

### Use Clear Headings

```markdown
## Purpose / When to Use
Specific scenarios.

## Step-by-Step Workflow
1. First step
2. Second step

## Common Patterns
Code snippets and examples.

## Pitfalls
Common mistakes to avoid.
```

### Be Actionable

**Bad:**
```markdown
The skill helps with deployments.
```

**Good:**
```markdown
## Deployment Workflow

1. Run `kubectl get pods` to check current state
2. Apply changes with `kubectl apply -f manifest.yaml`
3. Verify deployment: `kubectl rollout status deployment/<name>`
```

### Include Validation Steps

```markdown
## Definition of Done

1. [ ] All tests pass: `npm test`
2. [ ] Linting clean: `npm run lint`
3. [ ] Build succeeds: `npm run build`
```

## Progressive Disclosure

Keep SKILL.md focused. Move details to reference files:

```
skill-name/
├── SKILL.md              # Core instructions (read always)
├── REFERENCE.md          # Detailed docs (read when needed)
├── ADVANCED.md           # Edge cases (read rarely)
└── snippets/             # Templates (copy when used)
```

### Link from SKILL.md:

```markdown
For detailed API reference, see [REFERENCE.md](REFERENCE.md).

Common patterns are available in [snippets/](snippets/).
```

Claude reads referenced files only when the task requires them.

## Versioning Skills

Track changes in the Version History section:

```markdown
## Version History

- 2.0.0 — Breaking: Changed API version requirements
- 1.2.0 — Added support for multi-cluster deployments
- 1.1.0 — Added GitOps patterns
- 1.0.0 — Initial release
```

### When to Version:

- **Major (X.0.0)**: Breaking changes, restructured instructions
- **Minor (x.Y.0)**: New features, additional patterns
- **Patch (x.y.Z)**: Bug fixes, typo corrections

## Avoiding Conflicts

When multiple skills might match:

1. **Use distinct trigger words**:
```yaml
# Skill 1
description: ...for sales reports, pipeline analysis, revenue...

# Skill 2
description: ...for log analysis, system metrics, performance monitoring...
```

2. **Be explicit about file types**:
```yaml
description: ...when working with .xlsx Excel files...
```

3. **Mention the technology stack**:
```yaml
description: ...for PostgreSQL databases, not MongoDB or Redis...
```

## Testing Skills

### Manual Testing

1. Start Claude Code with your skill:
```bash
claude --plugin-dir ./plugins/skillbox
```

2. Ask questions that should trigger the skill:
```
Create a Helm chart for my app
```

3. Ask questions that should NOT trigger:
```
What's the weather today?
```

### Debug Mode

```bash
claude --debug
```

Shows skill loading and matching information.

## Common Mistakes

### 1. Description Too Generic
Skills don't activate because Claude can't match them.

### 2. Too Many Skills Overlap
Claude picks the wrong skill or gets confused.

### 3. Instructions Too Vague
Claude doesn't know what steps to follow.

### 4. Missing Examples
Users don't know how to invoke the skill.

### 5. No allowed-tools When Needed
Skill has unnecessary access or prompts for permissions.

### 6. Broken Links
Referenced files don't exist.
