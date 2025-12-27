# Infrastructure Components Reference

## Supported Components

| Component | Chart Repository | Purpose |
|-----------|------------------|---------|
| cert-manager | https://charts.jetstack.io | TLS certificates |
| ingress-nginx | https://kubernetes.github.io/ingress-nginx | Ingress controller |
| external-secrets | https://charts.external-secrets.io | Secret management |
| external-dns | https://kubernetes-sigs.github.io/external-dns | DNS automation |
| prometheus-stack | https://prometheus-community.github.io/helm-charts | Monitoring |

## cert-manager

### Base + Overlay Structure

```
infra/
├── components/
│   ├── base/cert-manager/       # Shared HelmRepo + HelmRelease
│   │   ├── kustomization.yaml
│   │   └── helm.yaml
│   └── crds/cert-manager/       # CRD management
│       ├── kustomization.yaml
│       ├── gitrepository.yaml
│       └── flux-kustomization.yaml
└── dev/cert-manager/            # Overlay (values only)
    ├── kustomization.yaml       # refs base + ConfigMapGenerator
    └── values.yaml
```

### CRDs (infra/components/crds/cert-manager/)

```yaml
# infra/components/crds/cert-manager/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - gitrepository.yaml
  - flux-kustomization.yaml
```

```yaml
# infra/components/crds/cert-manager/gitrepository.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: cert-manager-crds
  namespace: flux-system
spec:
  interval: 1h
  url: https://github.com/cert-manager/cert-manager
  ref:
    tag: v1.17.0  # Use Context7 for latest
  ignore: |
    /*
    !/deploy/crds
```

```yaml
# infra/components/crds/cert-manager/flux-kustomization.yaml
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
```

### Base (infra/components/base/cert-manager/)

```yaml
# infra/components/base/cert-manager/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

```yaml
# infra/components/base/cert-manager/helm.yaml
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
      version: "v1.17.0"  # Use Context7 for latest
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

### Overlay (infra/{env}/cert-manager/)

```yaml
# infra/dev/cert-manager/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/cert-manager

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: cert-manager-values
    files:
      - values.yaml
```

```yaml
# infra/dev/cert-manager/values.yaml
fullnameOverride: cert-manager

crds:
  enabled: false  # Managed via GitRepository

serviceAccount:
  create: true
  name: cert-manager

global:
  leaderElection:
    namespace: cert-manager

resources:
  requests:
    cpu: 10m
    memory: 64Mi
  limits:
    memory: 256Mi
```

### ClusterIssuer

```yaml
# infra/dev/cert-manager-issuer/clusterissuer.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
      - http01:
          ingress:
            class: nginx
```


## ingress-nginx

### Base + Overlay Structure

```
infra/
├── components/
│   └── base/ingress-nginx/      # Shared HelmRepo + HelmRelease
│       ├── kustomization.yaml
│       └── helm.yaml
└── dev/ingress-nginx/           # Overlay (values only)
    ├── kustomization.yaml
    └── values.yaml
```

Note: ingress-nginx has no CRDs, so no entry in `infra/components/crds/`.

### Base (infra/components/base/ingress-nginx/)

```yaml
# infra/components/base/ingress-nginx/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

```yaml
# infra/components/base/ingress-nginx/helm.yaml
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
      version: "4.12.0"  # Use Context7 for latest
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

### Overlay (infra/{env}/ingress-nginx/)

```yaml
# infra/dev/ingress-nginx/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/ingress-nginx

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: ingress-nginx-values
    files:
      - values.yaml
```

```yaml
# infra/dev/ingress-nginx/values.yaml
controller:
  replicaCount: 2

  service:
    type: LoadBalancer
    annotations:
      # AWS NLB
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
      service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing

  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      memory: 256Mi

  metrics:
    enabled: true
    serviceMonitor:
      enabled: false  # Enable if using prometheus-stack
```

## external-secrets-operator

### Base + Overlay Structure

```
infra/
├── components/
│   ├── base/external-secrets-operator/  # Shared HelmRepo + HelmRelease
│   │   ├── kustomization.yaml
│   │   └── helm.yaml
│   └── crds/external-secrets/           # CRD management
│       ├── kustomization.yaml
│       ├── gitrepository.yaml
│       └── flux-kustomization.yaml
└── dev/secrets-operator/                # Overlay (values only)
    ├── kustomization.yaml
    └── values.yaml
```

### CRDs (infra/components/crds/external-secrets/)

```yaml
# infra/components/crds/external-secrets/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - gitrepository.yaml
  - flux-kustomization.yaml
```

```yaml
# infra/components/crds/external-secrets/gitrepository.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: external-secrets-crds
  namespace: flux-system
spec:
  interval: 1h
  url: https://github.com/external-secrets/external-secrets
  ref:
    tag: v0.15.0  # Use Context7 for latest
  ignore: |
    /*
    !/deploy/crds
```

```yaml
# infra/components/crds/external-secrets/flux-kustomization.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: external-secrets-crds
  namespace: flux-system
spec:
  interval: 1h
  prune: false  # CRITICAL: Never delete CRDs
  sourceRef:
    kind: GitRepository
    name: external-secrets-crds
  path: ./deploy/crds
  wait: true
  healthChecks:
    - apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: externalsecrets.external-secrets.io
```

### Base (infra/components/base/external-secrets-operator/)

```yaml
# infra/components/base/external-secrets-operator/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

```yaml
# infra/components/base/external-secrets-operator/helm.yaml
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
      version: "0.15.0"  # Use Context7 for latest
      sourceRef:
        kind: HelmRepository
        name: external-secrets
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
      name: external-secrets-values
      valuesKey: values.yaml
```

### Overlay (infra/{env}/secrets-operator/)

```yaml
# infra/dev/secrets-operator/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/external-secrets-operator

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: external-secrets-values
    files:
      - values.yaml
```

```yaml
# infra/dev/secrets-operator/values.yaml
installCRDs: false  # Managed via GitRepository

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

## ClusterSecretStore

### AWS Secrets Manager

```yaml
# infra/components/base/secrets-store/clustersecretstore.yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      # auth via IRSA - no explicit credentials needed
```

### GCP Secret Manager

```yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: gcp-secret-manager
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    gcpsm:
      projectID: my-gcp-project
      # auth via Workload Identity
```

### Azure Key Vault

```yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: azure-key-vault
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    azurekv:
      tenantId: "tenant-id"
      vaultUrl: "https://my-vault.vault.azure.net"
      authType: ManagedIdentity
```

### HashiCorp Vault

```yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: hashicorp-vault
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.example.com/enabled: "true"
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "external-secrets"
```

## external-dns

### Base + Overlay Structure

```
infra/
├── components/
│   └── base/external-dns/      # Shared HelmRepo + HelmRelease
│       ├── kustomization.yaml
│       └── helm.yaml
└── dev/external-dns/           # Overlay (values only)
    ├── kustomization.yaml
    └── values.yaml
```

Note: external-dns has no CRDs, so no entry in `infra/components/crds/`.

### Base (infra/components/base/external-dns/)

```yaml
# infra/components/base/external-dns/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - helm.yaml
```

```yaml
# infra/components/base/external-dns/helm.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: external-dns
  namespace: flux-system
spec:
  interval: 1h
  url: https://kubernetes-sigs.github.io/external-dns
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: external-dns
  namespace: flux-system
spec:
  interval: 30m
  targetNamespace: external-dns
  chart:
    spec:
      chart: external-dns
      version: "1.16.0"  # Use Context7 for latest
      sourceRef:
        kind: HelmRepository
        name: external-dns
        namespace: flux-system
  install:
    createNamespace: true
  upgrade:
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: external-dns-values
      valuesKey: values.yaml
```

### Overlay (infra/{env}/external-dns/)

```yaml
# infra/dev/external-dns/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: flux-system

resources:
  - ../../components/base/external-dns

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: external-dns-values
    files:
      - values.yaml
```

```yaml
# infra/dev/external-dns/values.yaml
provider: aws

aws:
  region: us-east-1

domainFilters:
  - example.com

policy: sync  # upsert-only for safer updates

txtOwnerId: "dev-cluster"

serviceAccount:
  create: true
  name: external-dns
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/external-dns
```

## Cluster Orchestration Example

```yaml
# clusters/dev/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 00-crds.yaml
  - 02-secrets-operator.yaml
  - 03-secrets-store.yaml
  - 04-external-dns.yaml
  - 05-ingress-nginx.yaml
  - 06-cert-manager.yaml
  - 07-cert-manager-issuer.yaml
  - 99-apps-dev.yaml
```

## Dependency Graph

```
00-crds
    ├── cert-manager-crds
    ├── external-secrets-crds
    └── prometheus-operator-crds (optional)

02-secrets-operator
    └── depends on: crds

03-secrets-store
    └── depends on: secrets-operator

04-external-dns
    └── (no dependencies)

05-ingress-nginx
    └── (no dependencies)

06-cert-manager
    └── depends on: crds (cert-manager-crds)

07-cert-manager-issuer
    └── depends on: cert-manager

99-apps
    └── depends on: ingress-nginx, secrets-store
```
