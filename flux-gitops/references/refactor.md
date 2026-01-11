---
name: flux-gitops-refactor
description: This skill should be used when the user asks to "refactor Flux project", "convert GitOps structure", "upgrade Flux layout", or mentions "restructure gitops", "Flux refactoring", "standardize GitOps".
allowed-tools: [Read, Write, Edit, Glob, Grep, Bash]
---

# Flux GitOps Refactor

## Purpose

Refactor existing Flux CD GitOps repositories to the standardized structure defined in `flux-gitops-scaffold`. Supports refactoring from:

- **Flat structures** - All manifests in one directory
- **Partial structures** - Some base/overlay separation but non-standard
- **Helm-only** - HelmRelease without Kustomize overlays
- **Custom/hybrid** - Mixed patterns

**Target structure:** `flux-gitops-scaffold` pattern with:
- Vendored CRDs in `infra/crds/`
- Base/overlay separation with ConfigMapGenerator
- Proper `dependsOn` chains in Flux Kustomizations
- ExternalSecrets for sensitive data

## When to Use

Activate for:
- Refactoring legacy Flux projects to standard structure
- Analyzing existing GitOps repositories for compliance
- Converting flat HelmRelease files to base/overlay pattern
- Adding missing infrastructure (CRDs, aggregators, dependsOn)

## Refactoring Workflow

### Phase 1: Analysis

Run the analysis script to understand current structure:

```bash
python scripts/analyze-structure.py /path/to/gitops-repo
```

The script outputs a JSON report with:
- Structure type classification
- Detected environments
- Component inventory (controllers, configs, services, apps)
- Issues requiring refactoring
- HelmRelease analysis

Review the output and identify refactoring scope.

### Phase 2: Refactoring Plan

Based on analysis results, generate a refactoring plan:

| Structure Type | Refactoring Steps |
|----------------|-------------------|
| **Flat** | Create full directory structure, extract values, add Kustomize overlays |
| **Partial** | Add missing base/overlay, convert patches to ConfigMapGenerator |
| **Helm-only** | Wrap in Kustomize, add overlays per environment |
| **Custom** | Hybrid approach based on detected patterns |

For each HelmRelease, plan:
1. Move to `base/` directory
2. Add `valuesFrom: ConfigMap` reference
3. Create overlay with `configMapGenerator`
4. Extract inline values to `values.yaml`

### Phase 3: Execution

Execute refactoring semi-automatically:

1. **Create directory structure**
   ```
   infra/
   ├── base/cluster/controllers/{component}/
   ├── crds/{component}/
   └── {env}/cluster/controllers/{component}/
   ```

2. **Transform HelmRelease files**
   - Add `valuesFrom` pointing to ConfigMap
   - Set `crds: Skip` for components with CRDs
   - Move to base directory

3. **Create overlay files**
   - `kustomization.yaml` with ConfigMapGenerator
   - `values.yaml` with extracted/environment-specific values
   - Ensure `disableNameSuffixHash: true`

4. **Generate orchestration**
   - Create Flux Kustomizations with `dependsOn` chain
   - Add aggregator `kustomization.yaml` files

5. **Vendor CRDs** (for components that have them)
   - Download from upstream releases
   - Place in `infra/crds/{component}/`

Present changes for user review before committing.

## Analysis Script Usage

### Basic Analysis

```bash
python scripts/analyze-structure.py /path/to/repo
```

### Output Format

```json
{
  "structure_type": "partial",
  "environments": ["dev", "staging", "prod"],
  "components": {
    "controllers": ["cert-manager", "ingress-nginx"],
    "configs": ["cluster-issuer"],
    "services": ["redis"],
    "apps": ["api", "frontend"]
  },
  "issues": [
    {"type": "missing_base", "component": "cert-manager"},
    {"type": "no_values_from", "path": "infra/base/cert-manager/helm.yaml"},
    {"type": "missing_crds", "component": "cert-manager"}
  ],
  "helm_releases": [
    {
      "name": "cert-manager",
      "path": "infra/base/cert-manager/helm.yaml",
      "has_values_from": false,
      "has_crds_skip": false
    }
  ]
}
```

### Issue Types

| Issue | Description | Fix |
|-------|-------------|-----|
| `missing_base` | No base directory | Create base with HelmRelease |
| `no_values_from` | HelmRelease without valuesFrom | Add valuesFrom reference |
| `missing_crds` | CRDs not vendored | Vendor CRDs to infra/crds/ |
| `missing_aggregator` | No kustomization.yaml | Create aggregator file |
| `inline_values` | Values inline in HelmRelease | Extract to values.yaml |
| `patches_for_values` | Using patches for Helm values | Convert to ConfigMapGenerator |

## Key Transformations

### HelmRelease: Add valuesFrom

**Before:**
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cert-manager
spec:
  chart:
    spec:
      chart: cert-manager
      version: "X.Y.Z"  # May be outdated
  values:
    installCRDs: false
```

**After:**
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cert-manager
spec:
  chart:
    spec:
      chart: cert-manager
      version: "X.Y.Z"  # Verify via Context7
  install:
    crds: Skip
  upgrade:
    crds: Skip
  valuesFrom:
    - kind: ConfigMap
      name: cert-manager-values
      valuesKey: values.yaml
```

### Overlay: ConfigMapGenerator

**Create overlay kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: flux-system
resources:
  - ../../../../base/cluster/controllers/cert-manager
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
  - name: cert-manager-values
    files:
      - values.yaml
```

**Create values.yaml:**
```yaml
installCRDs: false
fullnameOverride: cert-manager
resources:
  requests:
    cpu: 10m
    memory: 64Mi
```

## Validation

After refactoring, validate with:

```bash
# Validate Kustomize structure
kubectl kustomize infra/dev/cluster/controllers/cert-manager

# Validate entire environment
kubectl kustomize infra/dev/cluster/controllers
```

## Version Updates During Refactoring

When refactoring existing Flux repos, versions should be updated to current:

### Detection

During analysis, flag potentially outdated versions:
- Components with versions older than 6 months
- Known deprecated versions (check release notes)
- Missing version field (rare but possible)

### Update Workflow

For each HelmRelease with version field:

```
# Step 1: Resolve library ID
Tool: resolve-library-id
libraryName: "{component-name}"

# Step 2: Get current version
Tool: query-docs
libraryId: "/{org}/{project}"
topic: "helm chart latest version"

# Step 3: Compare and update if needed
```

### Migration Report

Include version updates in refactoring report:

```markdown
## Version Updates

| Component | Old Version | New Version | Notes |
|-----------|-------------|-------------|-------|
| cert-manager | v1.14.0 | v1.17.0 | CRD changes - test required |
| ingress-nginx | 4.8.0 | 4.12.0 | Check deprecated annotations |
```

### CRD Re-Vendoring

When updating versions, re-vendor CRDs:

```bash
# Use Context7 to get {VERSION} first!
curl -sL https://github.com/cert-manager/cert-manager/releases/download/{VERSION}/cert-manager.crds.yaml \
  > infra/crds/cert-manager/crds.yaml
```

## Target Structure Reference

The target structure follows `flux-gitops-scaffold` pattern. Consult:
- `../flux-gitops-scaffold/references/project-structure.md` - Full directory layout
- `../flux-gitops-scaffold/references/infra-components.md` - Component patterns

## Refactoring Checklist

Before completing refactoring:

- [ ] All HelmRelease files have `valuesFrom: ConfigMap`
- [ ] All overlays use `configMapGenerator` (not patches for values)
- [ ] `disableNameSuffixHash: true` in all overlay kustomizations
- [ ] CRDs vendored for components that have them
- [ ] `crds: Skip` in HelmRelease for vendored CRDs
- [ ] Aggregator `kustomization.yaml` in each directory
- [ ] Flux Kustomizations have proper `dependsOn` chain
- [ ] `kubectl kustomize` passes for all paths
- [ ] **Versions verified via Context7** (not copied from old manifests)

## Additional Resources

### Reference Files

- **`references/refactoring-patterns.md`** - Detailed transformation patterns
- **`../flux-gitops-scaffold/SKILL.md`** - Target structure specification

### Example Files

- **`examples/refactoring-report.md`** - Sample analysis report

### Scripts

- **`scripts/analyze-structure.py`** - Structure analyzer

## Anti-Patterns to Fix

| Source Anti-Pattern | Refactoring Target |
|--------------------|-------------------|
| Inline values in HelmRelease | Extract to values.yaml + ConfigMapGenerator |
| Patches for Helm values | ConfigMapGenerator |
| `installCRDs: true` | `installCRDs: false` + vendored CRDs |
| `crds: CreateReplace` | `crds: Skip` |
| Missing dependsOn | Add proper dependency chain |
| Flat structure | Full base/overlay separation |

## Related Skills

- **`flux-gitops-scaffold`** - Target structure and patterns
- **`helm-chart-developer`** - Helm chart best practices
