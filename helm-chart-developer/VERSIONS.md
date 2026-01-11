# Version Compatibility Matrix

## Target Baseline

| Component | Version | Notes |
|-----------|---------|-------|
| Kubernetes | 1.29+ | API deprecations considered |
| Helm | 3.14+ | Chart apiVersion: v2 |
| Flux | v2.3+ | HelmRelease v2 API |
| External Secrets Operator | 0.10+ | v1 API stable |

## API Versions

| Resource | API Version | Documentation |
|----------|-------------|---------------|
| Chart.yaml | `apiVersion: v2` | [helm.sh/docs/topics/charts](https://helm.sh/docs/topics/charts/) |
| HelmRelease | `helm.toolkit.fluxcd.io/v2` | [fluxcd.io/flux/components/helm](https://fluxcd.io/flux/components/helm/) |
| HelmRepository | `source.toolkit.fluxcd.io/v1` | [fluxcd.io/flux/components/source](https://fluxcd.io/flux/components/source/) |
| GitRepository | `source.toolkit.fluxcd.io/v1` | [fluxcd.io/flux/components/source](https://fluxcd.io/flux/components/source/) |
| Kustomization | `kustomize.toolkit.fluxcd.io/v1` | [fluxcd.io/flux/components/kustomize](https://fluxcd.io/flux/components/kustomize/) |
| ExternalSecret | `external-secrets.io/v1` | [external-secrets.io/latest/api](https://external-secrets.io/latest/api/) |
| ClusterSecretStore | `external-secrets.io/v1` | [external-secrets.io/latest/api](https://external-secrets.io/latest/api/) |
| SecretStore | `external-secrets.io/v1` | [external-secrets.io/latest/api](https://external-secrets.io/latest/api/) |

## Migration Notes

If repository uses older APIs, propose migration:

| Old API | New API | Action |
|---------|---------|--------|
| `helm.toolkit.fluxcd.io/v2beta1` | `helm.toolkit.fluxcd.io/v2` | Update apiVersion |
| `helm.toolkit.fluxcd.io/v2beta2` | `helm.toolkit.fluxcd.io/v2` | Update apiVersion |
| `external-secrets.io/v1beta1` | `external-secrets.io/v1` | Update apiVersion |
| `source.toolkit.fluxcd.io/v1beta2` | `source.toolkit.fluxcd.io/v1` | Update apiVersion |
| `kustomize.toolkit.fluxcd.io/v1beta2` | `kustomize.toolkit.fluxcd.io/v1` | Update apiVersion |

## Deprecated Features

- Helm 2 charts (`apiVersion: v1`) — not supported
- Flux v1 (fluxcd.io/v1) — migrated to Flux v2
- ESO v1alpha1/v1beta1 — upgrade to v1

## Last Updated

2024-12 (Kubernetes 1.29+, Flux 2.3+, ESO 0.10+)
