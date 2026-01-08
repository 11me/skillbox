# GitOps Project Structure Reference

## Complete Directory Layout

```
gitops/
├── clusters/                             # Flux orchestration (per-cluster)
│   ├── dev/
│   │   ├── kustomization.yaml            # Aggregates all Flux Kustomizations
│   │   ├── 00-crds.yaml                  # CRDs (prune: false, wait: true)
│   │   ├── 01-controllers.yaml           # dependsOn: crds
│   │   ├── 02-cluster-configs.yaml       # dependsOn: controllers
│   │   ├── 03-services.yaml              # dependsOn: cluster-configs
│   │   ├── 99-apps.yaml                  # dependsOn: services, cluster-configs
│   │   └── flux-system/                  # Flux bootstrap
│   │       ├── gotk-components.yaml
│   │       ├── gotk-sync.yaml
│   │       └── kustomization.yaml
│   ├── staging/
│   └── prod/
│
├── infra/
│   ├── base/
│   │   ├── cluster/
│   │   │   ├── controllers/              # Operators/Controllers (Helm)
│   │   │   │   ├── cert-manager/
│   │   │   │   │   ├── kustomization.yaml
│   │   │   │   │   └── helm.yaml         # HelmRepo + HelmRelease
│   │   │   │   ├── ingress-nginx/
│   │   │   │   └── external-secrets/
│   │   │   └── configs/                  # Resources depending on controllers
│   │   │       ├── cluster-issuer/
│   │   │       │   ├── kustomization.yaml
│   │   │       │   └── cluster-issuer.yaml
│   │   │       └── secrets-store/
│   │   │           ├── kustomization.yaml
│   │   │           └── cluster-secret-store.yaml
│   │   └── services/                     # Namespace-specific services
│   │       └── redis/
│   │           ├── kustomization.yaml
│   │           └── helm.yaml
│   │
│   ├── crds/                             # Vendored CRDs (applied first)
│   │   ├── kustomization.yaml            # Aggregates all CRD subdirs
│   │   ├── cert-manager/
│   │   │   ├── kustomization.yaml
│   │   │   └── crds.yaml                 # Vendored from upstream
│   │   └── external-secrets/
│   │       ├── kustomization.yaml
│   │       └── crds.yaml
│   │
│   ├── dev/
│   │   ├── cluster/
│   │   │   ├── controllers/
│   │   │   │   ├── kustomization.yaml    # AGGREGATOR
│   │   │   │   ├── cert-manager/
│   │   │   │   │   ├── kustomization.yaml
│   │   │   │   │   └── values.yaml       # installCRDs: false
│   │   │   │   └── ingress-nginx/
│   │   │   └── configs/
│   │   │       ├── kustomization.yaml    # AGGREGATOR
│   │   │       ├── cluster-issuer/
│   │   │       │   ├── kustomization.yaml
│   │   │       │   └── cluster-issuer.yaml
│   │   │       └── secrets-store/
│   │   └── services/
│   │       ├── kustomization.yaml        # AGGREGATOR
│   │       └── redis/
│   │           ├── kustomization.yaml
│   │           ├── values.yaml           # With envFrom
│   │           ├── configs/
│   │           │   └── redis.config.yaml
│   │           └── secrets/
│   │               └── redis.external.yaml
│   ├── staging/
│   └── prod/
│
├── apps/
│   ├── base/
│   │   └── {app}/
│   │       ├── kustomization.yaml
│   │       └── helm.yaml
│   ├── dev/
│   │   ├── kustomization.yaml            # AGGREGATOR
│   │   ├── namespace.yaml
│   │   ├── image-automation.yaml
│   │   └── {app}/
│   │       ├── kustomization.yaml
│   │       ├── kustomizeconfig.yaml
│   │       ├── values.yaml               # With envFrom
│   │       ├── patches.yaml              # Image tag + automation marker
│   │       ├── configs/
│   │       │   └── {app}.config.yaml     # Plain ConfigMap
│   │       └── secrets/
│   │           └── {app}.external.yaml   # ExternalSecret
│   ├── staging/
│   └── prod/
│
└── charts/
    └── app/                              # Generic application chart
        ├── Chart.yaml
        ├── values.yaml
        └── templates/
```

## Orchestration Files

### clusters/{env}/kustomization.yaml (Aggregator)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - 00-crds.yaml
  - 01-controllers.yaml
  - 02-cluster-configs.yaml
  - 03-services.yaml
  - 99-apps.yaml
```

### clusters/{env}/00-crds.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: crds
  namespace: flux-system
spec:
  interval: 1h
  prune: false                # CRITICAL: Never delete CRDs
  path: ./infra/crds          # Points to vendored CRDs
  sourceRef:
    kind: GitRepository
    name: flux-system
  wait: true                  # Wait for CRDs to be established
```

### clusters/{env}/01-controllers.yaml

```yaml
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
  wait: true                  # Wait for pods ready
  timeout: 10m
  sourceRef:
    kind: GitRepository
    name: flux-system
```

### clusters/{env}/02-cluster-configs.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: cluster-configs
  namespace: flux-system
spec:
  dependsOn:
    - name: controllers       # ClusterIssuer needs cert-manager running
  interval: 15m
  path: ./infra/dev/cluster/configs
  prune: true
  wait: true
  timeout: 5m
  sourceRef:
    kind: GitRepository
    name: flux-system
```

### clusters/{env}/03-services.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: services
  namespace: flux-system
spec:
  dependsOn:
    - name: cluster-configs   # May need ClusterSecretStore
  interval: 15m
  path: ./infra/dev/services
  prune: true
  wait: true
  timeout: 10m
  sourceRef:
    kind: GitRepository
    name: flux-system
```

### clusters/{env}/99-apps.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: apps
  namespace: flux-system
spec:
  dependsOn:
    - name: services
    - name: cluster-configs
  interval: 15m
  path: ./apps/dev
  prune: true
  wait: true
  timeout: 10m
  sourceRef:
    kind: GitRepository
    name: flux-system
```

## CRD Management (Vendored)

### infra/crds/kustomization.yaml (Aggregator)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  - external-secrets
```

### infra/crds/{component}/kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - crds.yaml
```

### Vendoring CRDs

```bash
# Download vendored CRDs (use Context7 to get {VERSION} first!)
curl -sL https://github.com/cert-manager/cert-manager/releases/download/{VERSION}/cert-manager.crds.yaml \
  > infra/crds/cert-manager/crds.yaml

curl -sL https://raw.githubusercontent.com/external-secrets/external-secrets/{VERSION}/deploy/crds/bundle.yaml \
  > infra/crds/external-secrets/crds.yaml
```

## Controller Pattern

### infra/base/cluster/controllers/{component}/helm.yaml

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
      version: "{VERSION}"  # Use Context7 to resolve
      sourceRef:
        kind: HelmRepository
        name: cert-manager
        namespace: flux-system
  install:
    createNamespace: true
    crds: Skip              # CRDs managed in infra/crds/
  upgrade:
    crds: Skip
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: cert-manager-values
      valuesKey: values.yaml
```

### infra/{env}/cluster/controllers/kustomization.yaml (Aggregator)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  - ingress-nginx
  - external-secrets
```

### infra/{env}/cluster/controllers/{component}/kustomization.yaml (Overlay)

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

### infra/{env}/cluster/controllers/{component}/values.yaml

```yaml
# CRITICAL: Disable CRD installation - managed separately
installCRDs: false

fullnameOverride: cert-manager
serviceAccount:
  create: true
  name: cert-manager
resources:
  requests:
    cpu: 10m
    memory: 64Mi
  limits:
    memory: 256Mi
```

## Cluster Config Pattern

### infra/base/cluster/configs/{component}/cluster-issuer.yaml

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: ${EMAIL}
    privateKeySecretRef:
      name: letsencrypt-account-key
    solvers:
      - http01:
          ingress:
            class: nginx
```

### infra/{env}/cluster/configs/kustomization.yaml (Aggregator)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cluster-issuer
  - secrets-store
```

### infra/{env}/cluster/configs/{component}/kustomization.yaml (Overlay)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../../../base/cluster/configs/cluster-issuer
patches:
  - path: cluster-issuer.yaml
```

## Service Pattern (with configs/secrets)

### infra/{env}/services/kustomization.yaml (Aggregator)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - redis
```

### infra/{env}/services/{component}/kustomization.yaml (Full Overlay)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: dev
resources:
  - ../../../base/services/redis
  - configs/redis.config.yaml
  - secrets/redis.external.yaml
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
  - name: redis-values
    files:
      - values.yaml=values.yaml
configurations:
  - kustomizeconfig.yaml
```

### infra/{env}/services/{component}/configs/redis.config.yaml

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  REDIS_HOST: "redis-master"
  REDIS_PORT: "6379"
  REDIS_DB: "0"
```

### infra/{env}/services/{component}/secrets/redis.external.yaml

```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: redis
spec:
  refreshInterval: 1h
  secretStoreRef:
    kind: ClusterSecretStore
    name: secrets-store
  target:
    name: redis
    creationPolicy: Owner
  dataFrom:
    - extract:
        key: dev/redis
```

### infra/{env}/services/{component}/values.yaml

```yaml
auth:
  enabled: true
  existingSecret: redis
  existingSecretPasswordKey: password

master:
  persistence:
    enabled: true
    size: 1Gi
  resources:
    requests:
      cpu: 50m
      memory: 64Mi

replica:
  replicaCount: 0  # Dev environment
```

## App Pattern (with configs/secrets)

### apps/{env}/kustomization.yaml (Aggregator)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - namespace.yaml
  - image-automation.yaml
  - app-name
```

### apps/{env}/{app}/kustomization.yaml (Full Overlay)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: dev
resources:
  - ../../base/app-name
  - configs/app-name.config.yaml
  - secrets/app-name.external.yaml
patches:
  - path: patches.yaml
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
  - name: app-name-values
    files:
      - values.yaml=values.yaml
configurations:
  - kustomizeconfig.yaml
```

### apps/{env}/{app}/configs/{app}.config.yaml

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-name-config
data:
  PROJECT: "AppName"
  HOST: "0.0.0.0"
  PORT: "8000"
  LOG_LEVEL: "info"
  DB_HOST: "postgres.dev.svc"
  REDIS_HOST: "redis-master.dev.svc"
```

### apps/{env}/{app}/secrets/{app}.external.yaml

```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: app-name
spec:
  refreshInterval: 1h
  secretStoreRef:
    kind: ClusterSecretStore
    name: secrets-store
  target:
    name: app-name
    creationPolicy: Owner
  dataFrom:
    - extract:
        key: dev/app-name
```

### apps/{env}/{app}/values.yaml

```yaml
replicaCount: 1

image:
  repository: registry.example.com/app-name
  pullPolicy: IfNotPresent
  tag: latest  # Overridden by patches.yaml

envFrom:
  - configMapRef:
      name: app-name-config
  - secretRef:
      name: app-name

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
  hosts:
    - host: app.dev.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: app-name-tls
      hosts:
        - app.dev.example.com
```

### apps/{env}/{app}/patches.yaml

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: app-name
spec:
  values:
    image:
      tag: dev-abc123-12345 # {"$imagepolicy": "flux-system:app-name-dev:tag"}
```

### apps/{env}/{app}/kustomizeconfig.yaml

```yaml
nameReference:
  - kind: ConfigMap
    version: v1
    fieldSpecs:
      - path: spec/valuesFrom/name
        kind: HelmRelease
  - kind: Secret
    version: v1
    fieldSpecs:
      - path: spec/valuesFrom/name
        kind: HelmRelease
```

## Naming Conventions

| Resource | Pattern | Example |
|----------|---------|---------|
| HelmRelease | `{name}` | `redis`, `app-name` |
| ConfigMap (values) | `{name}-values` | `redis-values` |
| ConfigMap (config) | `{name}-config` | `app-name-config` |
| ExternalSecret | `{name}` | `app-name` |
| Secret (target) | `{name}` | `app-name` |
| ImageRepository | `{name}-{env}` | `app-name-dev` |
| ImagePolicy | `{name}-{env}` | `app-name-dev` |

**Note:** If env = namespace, skip `{env}` suffix in resource names for isolation via namespace.

## Environment Differences

| Aspect | Dev | Staging | Prod |
|--------|-----|---------|------|
| Image tag pattern | `dev-{sha}-{run_id}` | `staging-{sha}` | `v{semver}` |
| ImagePolicy | numerical (run_id) | numerical | semver |
| Auto-deploy | On push to main | On PR merge | On git tag |
| Replicas | 1-2 | 2 | 3+ |
| Resource limits | Low | Medium | High |
| Certificate issuer | letsencrypt-staging | letsencrypt-staging | letsencrypt-prod |
| Secret store | dev secrets | staging secrets | prod secrets |
