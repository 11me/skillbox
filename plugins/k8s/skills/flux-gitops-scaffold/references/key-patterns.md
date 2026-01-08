# Key Patterns Reference

Detailed YAML patterns for Flux GitOps scaffolding.

## HelmRelease with ConfigMap Values

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: app-name
  namespace: flux-system
spec:
  interval: 30m
  targetNamespace: app-namespace
  chart:
    spec:
      chart: chart-name
      version: "1.0.0"
      sourceRef:
        kind: HelmRepository
        name: repo-name
  install:
    crds: Skip
    createNamespace: true
  upgrade:
    crds: Skip
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: app-name-values
      valuesKey: values.yaml
```

## ConfigMap Generator (Kustomize)

```yaml
# infra/{env}/{component}/kustomization.yaml - Overlay pattern
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/cert-manager  # Reference shared base

generatorOptions:
  disableNameSuffixHash: true  # CRITICAL: prevents name changes

configMapGenerator:
  - name: cert-manager-values
    files:
      - values.yaml  # Environment-specific values
```

## Flux Kustomization with Dependencies

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: apps
  namespace: flux-system
spec:
  dependsOn:
    - name: ingress-nginx
    - name: secrets-store
  interval: 15m
  timeout: 10m
  path: ./apps/dev
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
  wait: true
```

## Image Automation Marker

```yaml
# In patches.yaml
spec:
  values:
    image:
      tag: dev-abc123-12345 # {"$imagepolicy": "flux-system:app-dev:tag"}
```

## Configs & Secrets Pattern

For services and apps, use configs/ + secrets/ directories.

### Plain ConfigMap (Non-Sensitive)

```yaml
# configs/{name}.config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {prefix}-{name}-config
data:
  HOST: "0.0.0.0"
  PORT: "8000"
  DB_HOST: "db.example.com"
```

### ExternalSecret (Sensitive)

```yaml
# secrets/{name}.external.yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {prefix}-{name}
spec:
  refreshInterval: 1h
  secretStoreRef:
    kind: ClusterSecretStore
    name: secrets-store
  target:
    name: {prefix}-{name}
    creationPolicy: Owner
  dataFrom:
    - extract:
        key: {secret-path}
```

### values.yaml with envFrom Injection

```yaml
envFrom:
  - configMapRef:
      name: {prefix}-{name}-config
  - secretRef:
      name: {prefix}-{name}
```

### Overlay kustomization.yaml (Full Pattern)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {env}
resources:
  - ../../../base/services/{name}  # or ../../base/{name} for apps
  - configs/{name}.config.yaml
  - secrets/{name}.external.yaml
configMapGenerator:
  - name: {name}-values
    files:
      - values.yaml=values.yaml
generatorOptions:
  disableNameSuffixHash: true
configurations:
  - kustomizeconfig.yaml
```

**Naming Convention:** If env = namespace, skip {env} suffix in resource names.
