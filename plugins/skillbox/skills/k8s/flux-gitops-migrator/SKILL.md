---
name: flux-gitops-migrator
description: >
  This skill should be used when the user asks to "migrate Flux project",
  "convert GitOps structure", "upgrade Flux layout", "transform flux structure",
  "migrate to standard GitOps", "analyze Flux project for migration",
  "restructure my gitops repo", or mentions "flux migration" or "gitops restructuring".
  Provides semi-automatic migration from any Flux structure to the standard
  flux-gitops-scaffold pattern.
globs: ["**/clusters/**/*.yaml", "**/apps/**/*.yaml", "**/infra/**/*.yaml"]
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# Flux GitOps Migrator

## Purpose

Migrate existing Flux CD GitOps repositories to the standardized structure defined in `flux-gitops-scaffold`. Supports migration from:

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
- Migrating legacy Flux projects to standard structure
- Analyzing existing GitOps repositories for compliance
- Converting flat HelmRelease files to base/overlay pattern
- Adding missing infrastructure (CRDs, aggregators, dependsOn)

## Migration Workflow

### Phase 1: Analysis

Run the analysis script to understand current structure:

```bash
python scripts/analyze-structure.py /path/to/gitops-repo
```

The script outputs a JSON report with:
- Structure type classification
- Detected environments
- Component inventory (controllers, configs, services, apps)
- Issues requiring migration
- HelmRelease analysis

Review the output and identify migration scope.

### Phase 2: Migration Plan

Based on analysis results, generate a migration plan:

| Structure Type | Migration Steps |
|----------------|-----------------|
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

Execute migration semi-automatically:

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
      version: "1.14.0"
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
      version: "1.14.0"
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

After migration, validate with:

```bash
# Validate Kustomize structure
kubectl kustomize infra/dev/cluster/controllers/cert-manager

# Validate entire environment
kubectl kustomize infra/dev/cluster/controllers
```

## Target Structure Reference

The target structure follows `flux-gitops-scaffold` pattern. Consult:
- `../flux-gitops-scaffold/references/project-structure.md` - Full directory layout
- `../flux-gitops-scaffold/references/infra-components.md` - Component patterns

## Migration Checklist

Before completing migration:

- [ ] All HelmRelease files have `valuesFrom: ConfigMap`
- [ ] All overlays use `configMapGenerator` (not patches for values)
- [ ] `disableNameSuffixHash: true` in all overlay kustomizations
- [ ] CRDs vendored for components that have them
- [ ] `crds: Skip` in HelmRelease for vendored CRDs
- [ ] Aggregator `kustomization.yaml` in each directory
- [ ] Flux Kustomizations have proper `dependsOn` chain
- [ ] `kubectl kustomize` passes for all paths

## Additional Resources

### Reference Files

- **`references/migration-patterns.md`** - Detailed transformation patterns
- **`../flux-gitops-scaffold/SKILL.md`** - Target structure specification

### Example Files

- **`examples/migration-report.md`** - Sample analysis report

### Scripts

- **`scripts/analyze-structure.py`** - Structure analyzer

## Anti-Patterns to Fix

| Source Anti-Pattern | Migration Target |
|--------------------|------------------|
| Inline values in HelmRelease | Extract to values.yaml + ConfigMapGenerator |
| Patches for Helm values | ConfigMapGenerator |
| `installCRDs: true` | `installCRDs: false` + vendored CRDs |
| `crds: CreateReplace` | `crds: Skip` |
| Missing dependsOn | Add proper dependency chain |
| Flat structure | Full base/overlay separation |

## Related Skills

- **`flux-gitops-scaffold`** - Target structure and patterns
- **`helm-chart-developer`** - Helm chart best practices
