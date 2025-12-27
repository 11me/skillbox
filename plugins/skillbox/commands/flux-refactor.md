---
name: flux-refactor
description: Refactor existing Flux GitOps repository to standard structure
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, AskUserQuestion
---

# Refactor Flux GitOps Repository

This command analyzes an existing Flux GitOps repository and refactors it to the standard `flux-gitops-scaffold` structure.

## Workflow

### Step 1: Identify Repository

Use AskUserQuestion to confirm:
1. **Repository path** - Path to the GitOps repository to refactor
2. **Backup confirmation** - Confirm backup/branch exists before changes

### Step 2: Run Analysis

Execute the analysis script:

```bash
python plugins/skillbox/skills/k8s/flux-gitops-refactor/scripts/analyze-structure.py /path/to/repo --format=markdown
```

Present analysis results to user showing:
- Structure type (flat, partial, helm-only, custom)
- Detected environments
- Component inventory
- Issues requiring refactoring

### Step 3: Create Refactoring Plan

Based on structure type, present refactoring steps:

| Structure Type | Refactoring Steps |
|----------------|-------------------|
| **Flat** | Create directory structure, extract values, add Kustomize overlays |
| **Partial** | Add missing base/overlay, convert patches to ConfigMapGenerator |
| **Helm-only** | Wrap in Kustomize, add overlays per environment |
| **Custom** | Hybrid approach based on detected patterns |

Get user confirmation before proceeding.

### Step 4: Execute Refactoring

For each HelmRelease:

1. **Create base structure** (if missing)
   - Move HelmRelease to `infra/base/cluster/controllers/{component}/helm.yaml`
   - Add `valuesFrom: ConfigMap` reference
   - Set `crds: Skip` for components with CRDs
   - Create `kustomization.yaml`

2. **Create overlay structure** (per environment)
   - Create `infra/{env}/cluster/controllers/{component}/kustomization.yaml`
   - Add `configMapGenerator` with `disableNameSuffixHash: true`
   - Extract values to `values.yaml`

3. **Vendor CRDs** (for cert-manager, external-secrets, etc.)
   - Create `infra/crds/{component}/`
   - Download CRDs from upstream releases

4. **Create aggregators**
   - Add `kustomization.yaml` to each directory level
   - Update cluster Flux Kustomizations with `dependsOn` chain

### Step 5: Validate

Run validation for each environment:

```bash
kubectl kustomize infra/dev/cluster/controllers
kubectl kustomize infra/staging/cluster/controllers
kubectl kustomize infra/prod/cluster/controllers
```

### Step 6: Review Changes

Present summary of all changes for user review before committing.

## Key Transformations

### HelmRelease: Add valuesFrom

Add to base HelmRelease:

```yaml
spec:
  install:
    crds: Skip
  upgrade:
    crds: Skip
  valuesFrom:
    - kind: ConfigMap
      name: {component}-values
      valuesKey: values.yaml
```

### Overlay: ConfigMapGenerator

Create overlay kustomization.yaml:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: flux-system
resources:
  - ../../../../base/cluster/controllers/{component}
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
  - name: {component}-values
    files:
      - values.yaml
```

## References

- Load `skills/k8s/flux-gitops-refactor/references/refactoring-patterns.md` for detailed patterns
- Load `skills/k8s/flux-gitops-scaffold/references/project-structure.md` for target structure
