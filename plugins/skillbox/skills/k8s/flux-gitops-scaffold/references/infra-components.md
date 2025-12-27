# Infrastructure Components Reference

## Component Types

| Type | Location | Examples |
|------|----------|----------|
| Controllers | `base/cluster/controllers/` | cert-manager, ingress-nginx, ESO |
| Cluster Configs | `base/cluster/configs/` | ClusterIssuer, ClusterSecretStore |
| Services | `base/services/` | redis, postgres |
| CRDs | `crds/` | Vendored from upstream releases |

## Supported Controllers

| Controller | Chart Repository | Has CRDs |
|------------|------------------|----------|
| cert-manager | https://charts.jetstack.io | Yes |
| ingress-nginx | https://kubernetes.github.io/ingress-nginx | No |
| external-secrets | https://charts.external-secrets.io | Yes |
| external-dns | https://kubernetes-sigs.github.io/external-dns | No |
| prometheus-stack | https://prometheus-community.github.io/helm-charts | Yes |

## cert-manager (Controller + CRDs)

### Directory Structure

```
infra/
├── base/cluster/controllers/cert-manager/
│   ├── kustomization.yaml
│   └── helm.yaml
├── crds/cert-manager/
│   ├── kustomization.yaml
│   └── crds.yaml              # Vendored
└── dev/cluster/controllers/cert-manager/
    ├── kustomization.yaml
    └── values.yaml            # installCRDs: false
```

### Vendored CRDs

```bash
curl -sL https://github.com/cert-manager/cert-manager/releases/download/v1.17.0/cert-manager.crds.yaml \
  > infra/crds/cert-manager/crds.yaml
```

```yaml
# infra/crds/cert-manager/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - crds.yaml
```

### Base Controller

```yaml
# infra/base/cluster/controllers/cert-manager/helm.yaml
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
      version: "v1.17.0"
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

### Overlay

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
# CRITICAL: CRDs managed separately
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

## ClusterIssuer (Cluster Config)

Depends on cert-manager controller being ready.

### Directory Structure

```
infra/
├── base/cluster/configs/cluster-issuer/
│   ├── kustomization.yaml
│   └── cluster-issuer.yaml
└── dev/cluster/configs/cluster-issuer/
    ├── kustomization.yaml
    └── cluster-issuer.yaml
```

### Base Config

```yaml
# infra/base/cluster/configs/cluster-issuer/cluster-issuer.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-account-key
    solvers:
      - http01:
          ingress:
            class: nginx
```

### Overlay

```yaml
# infra/dev/cluster/configs/cluster-issuer/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../../../base/cluster/configs/cluster-issuer
patches:
  - path: cluster-issuer.yaml
```

```yaml
# infra/dev/cluster/configs/cluster-issuer/cluster-issuer.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory  # Staging for dev
    email: dev-admin@example.com
```

## ingress-nginx (Controller, No CRDs)

### Directory Structure

```
infra/
├── base/cluster/controllers/ingress-nginx/
│   ├── kustomization.yaml
│   └── helm.yaml
└── dev/cluster/controllers/ingress-nginx/
    ├── kustomization.yaml
    └── values.yaml
```

### Base Controller

```yaml
# infra/base/cluster/controllers/ingress-nginx/helm.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: ingress-nginx
  namespace: flux-system
spec:
  interval: 1h
  url: https://kubernetes.github.io/ingress-nginx
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: ingress-nginx
  namespace: flux-system
spec:
  interval: 30m
  targetNamespace: ingress-nginx
  chart:
    spec:
      chart: ingress-nginx
      version: "4.12.0"
      sourceRef:
        kind: HelmRepository
        name: ingress-nginx
        namespace: flux-system
  install:
    createNamespace: true
  upgrade:
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: ingress-nginx-values
      valuesKey: values.yaml
```

### Overlay

```yaml
# infra/dev/cluster/controllers/ingress-nginx/values.yaml
controller:
  replicaCount: 2
  service:
    type: LoadBalancer
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
      service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      memory: 256Mi
```

## external-secrets (Controller + CRDs)

### Directory Structure

```
infra/
├── base/cluster/controllers/external-secrets/
│   ├── kustomization.yaml
│   └── helm.yaml
├── crds/external-secrets/
│   ├── kustomization.yaml
│   └── crds.yaml              # Vendored
└── dev/cluster/controllers/external-secrets/
    ├── kustomization.yaml
    └── values.yaml            # installCRDs: false
```

### Vendored CRDs

```bash
curl -sL https://raw.githubusercontent.com/external-secrets/external-secrets/v0.15.0/deploy/crds/bundle.yaml \
  > infra/crds/external-secrets/crds.yaml
```

### Base Controller

```yaml
# infra/base/cluster/controllers/external-secrets/helm.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: external-secrets
  namespace: flux-system
spec:
  interval: 1h
  url: https://charts.external-secrets.io
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: external-secrets
  namespace: flux-system
spec:
  interval: 30m
  targetNamespace: external-secrets
  chart:
    spec:
      chart: external-secrets
      version: "0.15.0"
      sourceRef:
        kind: HelmRepository
        name: external-secrets
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
      name: external-secrets-values
      valuesKey: values.yaml
```

### Overlay

```yaml
# infra/dev/cluster/controllers/external-secrets/values.yaml
# CRITICAL: CRDs managed separately
installCRDs: false

serviceAccount:
  create: true
  name: external-secrets
  annotations:
    # AWS IRSA
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/external-secrets
resources:
  requests:
    cpu: 10m
    memory: 64Mi
  limits:
    memory: 128Mi
```

## ClusterSecretStore (Cluster Config)

Depends on external-secrets controller being ready.

### Directory Structure

```
infra/
├── base/cluster/configs/secrets-store/
│   ├── kustomization.yaml
│   └── cluster-secret-store.yaml
└── dev/cluster/configs/secrets-store/
    ├── kustomization.yaml
    └── cluster-secret-store.yaml
```

### AWS Secrets Manager

```yaml
# infra/dev/cluster/configs/secrets-store/cluster-secret-store.yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: secrets-store
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
```

### GCP Secret Manager

```yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: secrets-store
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    gcpsm:
      projectID: my-gcp-project
```

### Yandex Lockbox

```yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: yandex-lockbox
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    yandexlockbox:
      auth:
        authorizedKeySecretRef:
          name: yc-auth
          key: authorized-key
```

## external-dns (Controller, No CRDs)

### Directory Structure

```
infra/
├── base/cluster/controllers/external-dns/
│   ├── kustomization.yaml
│   └── helm.yaml
└── dev/cluster/controllers/external-dns/
    ├── kustomization.yaml
    └── values.yaml
```

### Overlay

```yaml
# infra/dev/cluster/controllers/external-dns/values.yaml
provider: aws
aws:
  region: us-east-1
domainFilters:
  - example.com
policy: sync
txtOwnerId: "dev-cluster"
serviceAccount:
  create: true
  name: external-dns
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/external-dns
```

## Aggregator Pattern

Each directory pointed to by Flux **MUST** have `kustomization.yaml`:

```yaml
# infra/dev/cluster/controllers/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  - ingress-nginx
  - external-secrets
  - external-dns
```

```yaml
# infra/dev/cluster/configs/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cluster-issuer
  - secrets-store
```

```yaml
# infra/crds/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  - external-secrets
```

## Orchestration (5 Kustomizations)

```
00-crds                    # Vendored CRDs, prune: false, wait: true
    ↓
01-controllers             # dependsOn: crds, wait: true, timeout: 10m
    ↓
02-cluster-configs         # dependsOn: controllers, wait: true, timeout: 5m
    ↓
03-services                # dependsOn: cluster-configs, wait: true, timeout: 10m
    ↓
99-apps                    # dependsOn: services, cluster-configs
```

**Why this order:**
- CRDs must exist before controllers install CRs
- Controllers must be running before configs (ClusterIssuer needs cert-manager)
- Cluster configs needed before services (ClusterSecretStore for secrets)
- Services provide infrastructure for apps (redis, postgres)

## Dependency Graph

```
infra/crds/
├── cert-manager/crds.yaml          # Vendored
└── external-secrets/crds.yaml      # Vendored

infra/base/cluster/controllers/
├── cert-manager/                   # depends on: CRDs
├── ingress-nginx/                  # no CRD dependency
├── external-secrets/               # depends on: CRDs
└── external-dns/                   # no CRD dependency

infra/base/cluster/configs/
├── cluster-issuer/                 # depends on: cert-manager controller
└── secrets-store/                  # depends on: external-secrets controller

infra/base/services/
└── redis/                          # depends on: secrets-store (for secrets)
```

## Version Updates

Use Context7 to fetch latest versions:

```
resolve-library-id: libraryName="cert-manager"
get-library-docs: context7CompatibleLibraryID="/jetstack/cert-manager", topic="installation"
```

After version update:
1. Update chart version in `helm.yaml`
2. Re-vendor CRDs from new release
3. Test with `kubectl kustomize`
