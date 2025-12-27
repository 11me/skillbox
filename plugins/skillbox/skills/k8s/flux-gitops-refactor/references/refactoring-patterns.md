# Refactoring Patterns Reference

Detailed transformation patterns for refactoring Flux GitOps structures to the standard `flux-gitops-scaffold` pattern.

## Structure Types

### Flat Structure

All manifests in a single directory or minimal organization.

**Characteristics:**
- No base/overlay separation
- HelmRelease files with inline values
- No Kustomize overlays
- Single environment or environment embedded in filenames

**Example source:**
```
gitops/
├── cert-manager.yaml       # HelmRelease + inline values
├── ingress-nginx.yaml
├── redis.yaml
└── apps/
    ├── api.yaml
    └── frontend.yaml
```

### Partial Structure

Some organization exists but doesn't follow standard patterns.

**Characteristics:**
- Has base/overlay but non-standard paths
- May use patches instead of ConfigMapGenerator
- Missing CRDs separation
- Missing aggregator kustomization.yaml files

**Example source:**
```
gitops/
├── base/
│   ├── cert-manager/
│   │   └── helmrelease.yaml    # No valuesFrom
│   └── redis/
├── overlays/                    # Non-standard name
│   ├── dev/
│   │   └── cert-manager/
│   │       └── patch.yaml      # Using patches
│   └── prod/
└── clusters/
```

### Helm-only Structure

HelmRelease files without Kustomize integration.

**Characteristics:**
- Direct HelmRelease application
- No Kustomize overlays
- Environment differences in values inline or separate files
- May use Helm's native templating

**Example source:**
```
releases/
├── dev/
│   ├── cert-manager.yaml
│   └── redis.yaml
├── prod/
│   ├── cert-manager.yaml
│   └── redis.yaml
└── charts/
```

## Refactoring Pattern: Flat → Standard

### Step 1: Analyze Current Files

For each HelmRelease file:
1. Extract chart name, version, repository URL
2. Extract inline values
3. Identify component type (controller/service/app)

### Step 2: Create Directory Structure

```bash
mkdir -p infra/base/cluster/controllers/{component}
mkdir -p infra/crds/{component}  # If has CRDs
mkdir -p infra/{env}/cluster/controllers/{component}
```

### Step 3: Transform HelmRelease

**Before (flat):**
```yaml
# cert-manager.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: cert-manager
  namespace: flux-system
spec:
  interval: 1h
  url: https://charts.jetstack.io
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cert-manager
  namespace: flux-system
spec:
  interval: 30m
  chart:
    spec:
      chart: cert-manager
      version: "1.14.0"
      sourceRef:
        kind: HelmRepository
        name: cert-manager
  values:
    installCRDs: false
    fullnameOverride: cert-manager
    resources:
      requests:
        cpu: 10m
        memory: 64Mi
```

**After (base/cluster/controllers/cert-manager/helm.yaml):**
```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: cert-manager
  namespace: flux-system
spec:
  interval: 1h
  url: https://charts.jetstack.io
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cert-manager
  namespace: flux-system
spec:
  interval: 30m
  targetNamespace: cert-manager
  chart:
    spec:
      chart: cert-manager
      version: "1.14.0"
      sourceRef:
        kind: HelmRepository
        name: cert-manager
        namespace: flux-system
  install:
    createNamespace: true
    crds: Skip
  upgrade:
    crds: Skip
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: cert-manager-values
      valuesKey: values.yaml
```

### Step 4: Create Base Kustomization

**infra/base/cluster/controllers/cert-manager/kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

### Step 5: Create Overlay

**infra/dev/cluster/controllers/cert-manager/kustomization.yaml:**
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

**infra/dev/cluster/controllers/cert-manager/values.yaml:**
```yaml
installCRDs: false
fullnameOverride: cert-manager
resources:
  requests:
    cpu: 10m
    memory: 64Mi
  limits:
    memory: 256Mi
```

### Step 6: Vendor CRDs

```bash
# Get version from HelmRelease
VERSION="v1.14.0"

# Download CRDs
curl -sL "https://github.com/cert-manager/cert-manager/releases/download/${VERSION}/cert-manager.crds.yaml" \
  > infra/crds/cert-manager/crds.yaml
```

**infra/crds/cert-manager/kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - crds.yaml
```

### Step 7: Create Aggregators

**infra/dev/cluster/controllers/kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  - ingress-nginx
  # Add other controllers
```

**infra/crds/kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  # Add other CRDs
```

## Refactoring Pattern: Partial → Standard

### Differences from Flat Refactoring

1. **Preserve existing base** - May only need to add `valuesFrom`
2. **Convert patches to ConfigMapGenerator** - Replace patch files with values.yaml
3. **Reorganize paths** - Move to standard directory structure

### Converting Patches to ConfigMapGenerator

**Before (overlay with patches):**
```yaml
# overlays/dev/cert-manager/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../../base/cert-manager
patches:
  - path: patch-values.yaml
```

```yaml
# overlays/dev/cert-manager/patch-values.yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cert-manager
spec:
  values:
    resources:
      requests:
        cpu: 10m
```

**After (overlay with ConfigMapGenerator):**
```yaml
# infra/dev/cluster/controllers/cert-manager/kustomization.yaml
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

```yaml
# infra/dev/cluster/controllers/cert-manager/values.yaml
resources:
  requests:
    cpu: 10m
    memory: 64Mi
```

**And update base to use valuesFrom:**
```yaml
# Add to base HelmRelease
spec:
  valuesFrom:
    - kind: ConfigMap
      name: cert-manager-values
      valuesKey: values.yaml
```

## Refactoring Pattern: Helm-only → Standard

### Key Transformation

Wrap each HelmRelease in Kustomize structure:

1. Create base directory with HelmRelease (add valuesFrom)
2. Create overlay per environment with ConfigMapGenerator
3. Create Flux Kustomization to apply the overlay

### Example

**Before:**
```
releases/dev/redis.yaml
releases/prod/redis.yaml
```

**After:**
```
infra/
├── base/services/redis/
│   ├── kustomization.yaml
│   └── helm.yaml              # HelmRelease with valuesFrom
├── dev/services/redis/
│   ├── kustomization.yaml     # ConfigMapGenerator
│   └── values.yaml
└── prod/services/redis/
    ├── kustomization.yaml
    └── values.yaml
```

## Orchestration Refactoring

### Creating Flux Kustomizations

For proper dependency ordering, create numbered Flux Kustomizations:

```yaml
# clusters/dev/00-crds.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: crds
  namespace: flux-system
spec:
  interval: 15m
  path: ./infra/crds
  prune: false  # CRITICAL: Never prune CRDs
  wait: true
  sourceRef:
    kind: GitRepository
    name: flux-system
```

```yaml
# clusters/dev/01-controllers.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: controllers
  namespace: flux-system
spec:
  dependsOn:
    - name: crds
  interval: 15m
  path: ./infra/dev/cluster/controllers
  prune: true
  wait: true
  timeout: 10m
  sourceRef:
    kind: GitRepository
    name: flux-system
```

```yaml
# clusters/dev/02-cluster-configs.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: cluster-configs
  namespace: flux-system
spec:
  dependsOn:
    - name: controllers
  interval: 15m
  path: ./infra/dev/cluster/configs
  prune: true
  wait: true
  timeout: 5m
  sourceRef:
    kind: GitRepository
    name: flux-system
```

## Common Issues and Fixes

### Issue: HelmRelease without targetNamespace

**Fix:** Add `targetNamespace` to spec:
```yaml
spec:
  targetNamespace: cert-manager
```

### Issue: Missing namespace creation

**Fix:** Add to install section:
```yaml
spec:
  install:
    createNamespace: true
```

### Issue: CRDs installed by Helm

**Fix:** Disable in HelmRelease and vendor separately:
```yaml
spec:
  install:
    crds: Skip
  upgrade:
    crds: Skip
```

And in values.yaml:
```yaml
installCRDs: false
```

### Issue: Hash suffix on ConfigMap names

**Fix:** Add to overlay kustomization:
```yaml
generatorOptions:
  disableNameSuffixHash: true
```

### Issue: Missing remediation

**Fix:** Add upgrade remediation:
```yaml
spec:
  upgrade:
    remediation:
      retries: 3
```

## Validation Commands

After refactoring, validate each component:

```bash
# Validate single component overlay
kubectl kustomize infra/dev/cluster/controllers/cert-manager

# Validate entire controller set
kubectl kustomize infra/dev/cluster/controllers

# Dry-run apply
kubectl kustomize infra/dev/cluster/controllers | kubectl apply --dry-run=server -f -
```

## Rollback Strategy

If refactoring causes issues:

1. Keep original files in a backup branch
2. Test refactoring in non-production first
3. Use Flux's suspend feature during refactoring:
   ```bash
   flux suspend kustomization controllers
   ```
4. After validation, resume:
   ```bash
   flux resume kustomization controllers
   ```
