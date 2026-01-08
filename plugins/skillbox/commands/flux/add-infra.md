---
name: add-infra
description: Add infrastructure component (cert-manager, ingress-nginx, external-secrets, etc.) to GitOps project
argument-hint: "<component-name>"
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - mcp__plugin_context7_context7__*
---

# Add Infrastructure Component

This command adds an infrastructure component to an existing Flux GitOps project.

## Supported Components

| Component | Chart Repository | Has CRDs |
|-----------|------------------|----------|
| cert-manager | jetstack | Yes |
| ingress-nginx | ingress-nginx | No |
| external-secrets | external-secrets | Yes |
| external-dns | bitnami | No |
| prometheus | prometheus-community | Yes |
| metrics-server | metrics-server | No |

## Workflow

### Step 1: Validate Component

Check that `{component}` is in the supported list. If not, inform user and suggest alternatives.

### Step 2: Detect Project Structure

Use Glob to find existing GitOps structure:

```
**/clusters/*/kustomization.yaml
**/infra/*/kustomization.yaml
**/infra/crds/*/kustomization.yaml
```

Identify:
- Available environments (dev, staging, prod)
- Existing infrastructure components in each environment
- Existing CRDs in `infra/crds/`
- Project root directory

### Step 3: Get Latest Version

Use Context7 to fetch current chart version:

```
resolve-library-id: "{component}"
get-library-docs: topic="helm" or "installation"
```

Extract version or use `references/version-matrix.md` as fallback.

### Step 4: Create CRDs (if applicable)

If component has CRDs, create `infra/components/crds/{component}/`:

**kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - gitrepository.yaml
  - flux-kustomization.yaml
```

**gitrepository.yaml:**
```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: {component}-crds
  namespace: flux-system
spec:
  interval: 1h
  url: https://github.com/{org}/{repo}
  ref:
    tag: v{version}
  ignore: |
    /*
    !/deploy/crds
```

**flux-kustomization.yaml:**
```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: {component}-crds
  namespace: flux-system
spec:
  interval: 1h
  prune: false  # CRITICAL: Never delete CRDs
  sourceRef:
    kind: GitRepository
    name: {component}-crds
  path: ./deploy/crds
  wait: true
```

### Step 5: Create Base Component

Create shared base in `infra/components/base/{component}/`:

**kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

**helm.yaml:**
```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: {component}
  namespace: flux-system
spec:
  interval: 1h
  url: {chart-repository-url}
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: {component}
  namespace: flux-system
spec:
  interval: 30m
  targetNamespace: {target-namespace}
  chart:
    spec:
      chart: {chart-name}
      version: "{version}"
      sourceRef:
        kind: HelmRepository
        name: {component}
        namespace: flux-system
  install:
    createNamespace: true
    crds: Skip  # If has CRDs
  upgrade:
    crds: Skip
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: {component}-values
      valuesKey: values.yaml
```

### Step 6: Create Environment Overlays

For each environment (dev, staging, prod), create `infra/{env}/{component}/`:

**kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/{component}

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: {component}-values
    files:
      - values.yaml
```

**values.yaml:**
```yaml
# Environment-specific values for {component}
# Customize per environment
```

**Validation:** Run `kubectl kustomize infra/{env}/{component}` to validate.

### Step 7: Update Cluster Orchestration

For each environment, update `clusters/{env}/` with appropriate numbered Kustomization:

Determine correct number based on dependencies:
- 00: CRDs (prune: false)
- 02: secrets-operator (external-secrets)
- 10: cert-manager
- 20: ingress controllers
- 30: monitoring (prometheus, metrics-server)
- 40: dns (external-dns)

Create `clusters/{env}/XX-{component}.yaml`:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: {component}
  namespace: flux-system
spec:
  interval: 1h
  retryInterval: 1m
  timeout: 5m
  sourceRef:
    kind: GitRepository
    name: flux-system
  path: ./infra/{env}/{component}
  prune: true
  wait: true
  dependsOn:
    - name: {dependency}  # Based on component type
```

Update `clusters/{env}/kustomization.yaml` to include new file.

### Step 8: Update CRDs Reference (if applicable)

If component has CRDs, update `clusters/{env}/00-crds.yaml` to include healthCheck:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: crds
  namespace: flux-system
spec:
  path: ./infra/components/crds
  prune: false
  healthChecks:
    - apiVersion: kustomize.toolkit.fluxcd.io/v1
      kind: Kustomization
      name: {component}-crds  # Add new component
```

## Component-Specific Patterns

### cert-manager

- Namespace: cert-manager
- CRDs: Yes (install separately)
- Dependencies: CRDs first
- Common values: installCRDs: false

### ingress-nginx

- Namespace: ingress-nginx
- Dependencies: cert-manager (for TLS)
- Common values: controller.service.type, controller.ingressClassResource

### external-secrets

- Namespace: external-secrets
- CRDs: Yes
- Dependencies: CRDs first
- Requires: ClusterSecretStore configuration per environment

### prometheus

- Namespace: monitoring
- CRDs: Yes (ServiceMonitor, PodMonitor, etc.)
- Dependencies: CRDs first
- Common values: server.retention, alertmanager.enabled

## Output

After completion, provide:

1. Summary of created files
2. Reminder to commit and push changes
3. How to verify deployment: `flux reconcile kustomization {component} --with-source`

## References

- Load `references/infra-components.md` for detailed patterns
- Load `references/version-matrix.md` for versions
- Use `examples/helmrelease-base.yaml` as template
