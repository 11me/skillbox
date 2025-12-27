# GitOps Project Structure Reference

## Complete Directory Layout

```
gitops/
├── clusters/                           # Flux orchestration (per-cluster)
│   ├── dev/
│   │   ├── 00-crds.yaml                # CRD Kustomizations
│   │   ├── 02-secrets-operator.yaml    # ESO operator
│   │   ├── 03-secrets-store.yaml       # ClusterSecretStore
│   │   ├── 04-external-dns.yaml        # DNS automation (optional)
│   │   ├── 05-ingress-nginx.yaml       # Ingress controller
│   │   ├── 06-cert-manager.yaml        # TLS certificates
│   │   ├── 07-cert-manager-issuer.yaml # ClusterIssuer
│   │   ├── 99-apps-dev.yaml            # Application deployment
│   │   ├── flux-system/                # Flux bootstrap
│   │   │   ├── gotk-components.yaml    # Flux controllers
│   │   │   ├── gotk-sync.yaml          # GitRepository + Kustomization
│   │   │   └── kustomization.yaml
│   │   └── kustomization.yaml          # Root kustomization
│   ├── staging/                        # Same structure as dev
│   └── prod/                           # Same structure as dev
│
├── infra/                              # Infrastructure components
│   ├── components/
│   │   ├── base/                       # Shared HelmRepository + HelmRelease
│   │   │   ├── cert-manager/
│   │   │   │   ├── kustomization.yaml  # resources: [helm.yaml]
│   │   │   │   └── helm.yaml           # HelmRepo + HelmRelease
│   │   │   ├── external-dns/
│   │   │   ├── external-secrets-operator/
│   │   │   ├── ingress-nginx/
│   │   │   └── secrets-store/
│   │   └── crds/                       # CRD Kustomizations
│   │       ├── cert-manager/
│   │       │   ├── kustomization.yaml
│   │       │   ├── gitrepository.yaml
│   │       │   └── flux-kustomization.yaml
│   │       ├── external-secrets/
│   │       └── kustomization.yaml      # Root for all CRDs
│   │
│   ├── dev/                            # Environment overlays (values only)
│   │   ├── cert-manager/
│   │   │   ├── kustomization.yaml      # refs base + ConfigMapGenerator
│   │   │   └── values.yaml             # Dev-specific values
│   │   ├── cert-manager-issuer/
│   │   │   ├── kustomization.yaml
│   │   │   └── cluster-issuer.yaml
│   │   ├── external-dns/
│   │   ├── ingress-nginx/
│   │   ├── secrets-operator/
│   │   └── secrets-store/
│   │
│   ├── staging/                        # Same structure as dev
│   └── prod/                           # Same structure as dev
│
├── apps/                               # Application deployments
│   ├── base/                           # Base HelmReleases
│   │   └── {app-name}/
│   │       ├── helm.yaml               # HelmRelease referencing charts/app
│   │       └── kustomization.yaml
│   │
│   ├── dev/                            # Dev environment
│   │   ├── {app-name}/
│   │   │   ├── kustomization.yaml      # Patches + ConfigMapGenerator
│   │   │   ├── kustomizeconfig.yaml    # ConfigMap name reference
│   │   │   ├── values.yaml             # Dev values
│   │   │   ├── patches.yaml            # Image tag with marker
│   │   │   └── secrets/
│   │   │       └── {app}.external.yaml # ExternalSecret
│   │   ├── image-automation.yaml       # All ImageRepository/Policy/Automation
│   │   ├── namespace.yaml              # Namespace with ESO label
│   │   └── kustomization.yaml          # Root app kustomization
│   │
│   ├── staging/
│   └── prod/
│
└── charts/                             # Helm charts
    └── app/                            # Generic application chart
        ├── Chart.yaml
        ├── values.yaml
        ├── README.md
        └── templates/
            ├── _helpers.tpl
            ├── deployment.yaml
            ├── service.yaml
            ├── ingress.yaml
            ├── hpa.yaml
            ├── pdb.yaml
            └── serviceaccount.yaml
```

## File Contents Reference

### clusters/{env}/kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 00-crds.yaml
  - 02-secrets-operator.yaml
  - 03-secrets-store.yaml
  - 05-ingress-nginx.yaml
  - 06-cert-manager.yaml
  - 07-cert-manager-issuer.yaml
  - 99-apps-dev.yaml
```

### clusters/{env}/00-crds.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: crds
  namespace: flux-system
spec:
  interval: 10m
  path: ./infra/components/crds
  prune: false  # CRITICAL: Never delete CRDs
  sourceRef:
    kind: GitRepository
    name: flux-system
  healthChecks:
    - apiVersion: kustomize.toolkit.fluxcd.io/v1
      kind: Kustomization
      name: cert-manager-crds
    - apiVersion: kustomize.toolkit.fluxcd.io/v1
      kind: Kustomization
      name: external-secrets-crds
```

### infra/components/crds/{component}/kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - gitrepository.yaml
  - flux-kustomization.yaml
```

### infra/components/crds/{component}/gitrepository.yaml

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: cert-manager-crds
  namespace: flux-system
spec:
  interval: 1h
  url: https://github.com/cert-manager/cert-manager
  ref:
    tag: v1.17.0
  ignore: |
    /*
    !/deploy/crds
```

### infra/components/crds/{component}/flux-kustomization.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: cert-manager-crds
  namespace: flux-system
spec:
  interval: 1h
  prune: false  # CRITICAL: Never delete CRDs
  sourceRef:
    kind: GitRepository
    name: cert-manager-crds
  path: ./deploy/crds
  wait: true
  healthChecks:
    - apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: certificates.cert-manager.io
    - apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: issuers.cert-manager.io
```

### infra/components/base/{component}/kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

### infra/components/base/{component}/helm.yaml

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
      version: "v1.17.0"
      sourceRef:
        kind: HelmRepository
        name: cert-manager
        namespace: flux-system
  install:
    createNamespace: true
    crds: Skip  # CRDs managed in infra/components/crds/
  upgrade:
    remediation:
      retries: 3
    crds: Skip
  valuesFrom:
    - kind: ConfigMap
      name: cert-manager-values
      valuesKey: values.yaml
```

### clusters/{env}/99-apps-dev.yaml

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: apps-dev
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

### infra/{env}/{component}/kustomization.yaml (Overlay)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/cert-manager  # Reference shared base

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: cert-manager-values
    files:
      - values.yaml
```

### infra/{env}/{component}/values.yaml

```yaml
# Environment-specific values for cert-manager
fullnameOverride: cert-manager

crds:
  enabled: false  # Managed via GitRepository

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

### apps/{env}/{app}/kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: app-namespace

resources:
  - ../../base/app-name
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

### apps/{env}/namespace.yaml

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: app-namespace
  labels:
    eso.domain.com/enabled: "true"  # Required for ClusterSecretStore access
```

## Naming Conventions

| Resource | Pattern | Example |
|----------|---------|---------|
| HelmRelease | `{app-name}` | `fce-web` |
| ConfigMap (values) | `{app-name}-values` | `fce-web-values` |
| ExternalSecret | `{app-name}-{env}` | `fce-web-dev` |
| ImageRepository | `{app-name}-{env}` | `fce-web-dev` |
| ImagePolicy | `{app-name}-{env}` | `fce-web-dev` |
| ImageUpdateAutomation | `{app-name}-auto-{env}` | `fce-web-auto-dev` |
| Namespace | `{env}` or `{app-name}` | `dev` |

## Environment Differences

| Aspect | Dev | Staging | Prod |
|--------|-----|---------|------|
| Image tag pattern | `dev-{sha}-{run_id}` | `staging-{sha}` | `v{semver}` |
| ImagePolicy | numerical (run_id) | numerical | semver |
| Auto-deploy | On push to main | On PR merge | On git tag |
| Replicas | 1-2 | 2 | 3+ |
| Resource limits | Low | Medium | High |
| Certificate issuer | letsencrypt-staging | letsencrypt-staging | letsencrypt-prod |
