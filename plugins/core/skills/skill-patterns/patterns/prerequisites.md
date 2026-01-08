# Prerequisites Pattern

Document what must exist before a skill activates.

## Pattern Structure

```markdown
## Prerequisites

**Required tools:**
- tool1 (version X+)
- tool2

**Required files:**
- File or directory pattern

**Environment:**
- Context assumptions
```

## Categories

### Tools

Software that must be installed:

```markdown
**Required tools:**
- helm 3.14+ (chart development)
- kubectl with cluster access (dry-run validation)
- flux (GitOps deployment)
```

### Files

Files or directories that must exist:

```markdown
**Required files:**
- Chart.yaml in current directory
- values.yaml with base configuration
- templates/ directory
```

### Environment

Context or state requirements:

```markdown
**Environment:**
- KUBECONFIG set or default context available
- Write access to target directory
- Git repository initialized
```

## Examples

### Helm Chart Developer

```markdown
## Prerequisites

**Required tools:**
- helm 3.14+
- kubectl with cluster access
- kubeconform (optional, for strict validation)

**Required files:**
- Chart.yaml in working directory
- values.yaml with base configuration

**Environment:**
- KUBECONFIG set or default context available
```

### TypeScript Project

```markdown
## Prerequisites

**Required tools:**
- node 20+
- pnpm 9+
- biome (linting)

**Required files:**
- package.json
- tsconfig.json

**Environment:**
- npm registry accessible
```

### Python Project

```markdown
## Prerequisites

**Required tools:**
- python 3.12+
- uv (package manager)
- ruff (linting)

**Required files:**
- pyproject.toml

**Environment:**
- Virtual environment active or uv available
```

## Validation Strategy

Skills can validate prerequisites before execution:

```markdown
### Do (Execute)
1. Check prerequisites:
   - `helm version` → verify helm installed
   - `test -f Chart.yaml` → verify chart exists
2. If missing, provide installation instructions
3. Proceed with main workflow
```

## Relationship to Inputs

| Concept | Scope | Example |
|---------|-------|---------|
| Prerequisites | System/environment | helm installed |
| Inputs | User-provided data | app_name, namespace |
