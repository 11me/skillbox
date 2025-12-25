# YAML Frontmatter Reference

Every SKILL.md file must begin with YAML frontmatter enclosed in `---` delimiters.

## Required Fields

### name

**Purpose**: Unique identifier for the skill.

**Rules**:
- Lowercase letters, numbers, and hyphens only
- Pattern: `[a-z0-9-]+`
- Maximum: 64 characters
- Must be unique within the skill source (personal/project/plugin)

**Valid examples**:
```yaml
name: pdf-processor
name: git-commit-helper
name: k8s-debugger
name: api-docs-generator
```

**Invalid examples**:
```yaml
name: PDF_Processor    # Uppercase and underscore
name: git commit       # Spaces not allowed
name: my.skill.name    # Dots not allowed
name: skill@v2         # Special characters not allowed
```

### description

**Purpose**: Tells Claude what the skill does and when to use it.

**Rules**:
- Maximum: 1024 characters
- Format: `<What it does>. Use when <trigger scenarios>.`
- Must include trigger keywords users would mention

**Good example**:
```yaml
description: Build production Helm charts with GitOps deployment via Flux. Use when creating Helm charts, converting manifests to Helm, setting up GitOps with Flux, or integrating External Secrets Operator.
```

**Bad example**:
```yaml
description: For Kubernetes stuff
```

## Optional Fields

### globs

**Purpose**: Auto-activate skill when working with matching files.

**Format**: Array of glob patterns.

**Rules**:
- Patterns follow standard glob syntax (`**`, `*`, `?`)
- When files match, skill context is loaded automatically
- Multiple patterns can be specified

**Examples**:

Helm chart development:
```yaml
globs: ["**/Chart.yaml", "**/values.yaml", "**/templates/**"]
```

Go project:
```yaml
globs: ["**/*.go", "**/go.mod", "**/go.sum"]
```

Python project:
```yaml
globs: ["**/*.py", "**/pyproject.toml", "**/requirements.txt"]
```

Kubernetes manifests:
```yaml
globs: ["**/kustomization.yaml", "**/*.yaml"]
```

**Behavior**:
- When matched: Skill activates automatically without explicit invocation
- When omitted: Skill activates only through description keyword matching

---

### allowed-tools

**Purpose**: Restricts which tools Claude can use when the skill is active.

**Format**: Comma-separated list of tool names.

**Available tools**:
| Tool | Description |
|------|-------------|
| `Read` | Read file contents |
| `Write` | Create new files |
| `Edit` | Modify existing files |
| `Grep` | Search within file contents |
| `Glob` | Find files by pattern |
| `Bash` | Execute shell commands |

**Examples**:

Read-only skill:
```yaml
allowed-tools: Read, Grep, Glob
```

Full write access:
```yaml
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
```

Analysis without shell:
```yaml
allowed-tools: Read, Grep, Glob, Write, Edit
```

**Behavior**:
- When set: Claude uses only specified tools without asking permission
- When omitted: Claude asks for permission to use tools as normal

## Complete Frontmatter Examples

### Minimal (required fields only)

```yaml
---
name: code-reviewer
description: Review code for best practices, security issues, and performance. Use when reviewing PRs, checking code quality, or analyzing existing code.
---
```

### With Tool Restrictions

```yaml
---
name: safe-file-reader
description: Read and analyze files without making changes. Use when you need read-only access to explore code or documentation.
allowed-tools: Read, Grep, Glob
---
```

### With Globs (Auto-activation)

```yaml
---
name: go-conventions
description: Go code review context. Use when reviewing Go code, checking conventions, or writing Go.
globs: ["**/*.go", "**/go.mod", "**/go.sum"]
allowed-tools: Read, Grep, Glob
---
```

### Full Featured

```yaml
---
name: helm-chart-developer
description: Build production Helm charts with GitOps via Flux HelmRelease + Kustomize overlays. Includes External Secrets Operator patterns. Use when authoring Helm charts, converting raw manifests to Helm, designing values/schema, or debugging helm template/lint/dry-run issues.
globs: ["**/Chart.yaml", "**/values.yaml", "**/templates/**", "**/kustomization.yaml"]
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---
```

## Validation Errors

### Missing Opening Delimiter

```yaml
name: my-skill
description: Does something
---
```
**Error**: Frontmatter must start with `---` on line 1.

### Missing Closing Delimiter

```yaml
---
name: my-skill
description: Does something

# Content starts here
```
**Error**: Frontmatter must end with `---` before content.

### Invalid YAML Syntax

```yaml
---
name: my-skill
description: Has: colons: without: quotes
---
```
**Error**: Values with special characters must be quoted.

**Fix**:
```yaml
---
name: my-skill
description: "Has: colons: with: quotes"
---
```

### Tabs Instead of Spaces

```yaml
---
name: my-skill
	description: Indented with tab
---
```
**Error**: YAML requires spaces for indentation.

### Name Too Long

```yaml
---
name: this-is-a-very-long-skill-name-that-exceeds-the-sixty-four-character-limit
description: Something
---
```
**Error**: Name exceeds 64 character limit.

## Debugging Frontmatter

View frontmatter of existing skill:
```bash
head -n 10 SKILL.md
```

Validate YAML syntax:
```bash
# Using yq
cat SKILL.md | sed -n '/^---$/,/^---$/p' | yq .

# Using Python
python3 -c "import yaml; print(yaml.safe_load(open('SKILL.md').read().split('---')[1]))"
```

Run Claude Code in debug mode to see skill loading:
```bash
claude --debug
```
