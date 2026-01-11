# Version Matrix Reference

**CRITICAL:** Never hardcode versions. Always use Context7 to get current versions.

**Enforcement:**
1. SessionStart hook — reminds about Context7 for k8s projects
2. PreToolUse hook — checks versions in HelmRelease files

## Getting Chart Versions

**Recommended workflow** — verify versions before scaffolding:

### Option 1: Context7 (Documentation)

```
Tool: resolve-library-id
Parameter: libraryName="cert-manager"

Tool: query-docs
Parameters:
  libraryId: "/cert-manager/website"
  topic: "helm chart install version"
```

**Note:** Context7 returns documentation examples which may contain older versions.
Use as reference but verify current release.

### Option 2: Helm Repo (Authoritative)

```bash
helm repo add jetstack https://charts.jetstack.io
helm search repo jetstack/cert-manager --versions | head -5
```

### Option 3: GitHub Releases

Check the project's GitHub releases page for latest stable version.

## Component Library IDs

| Component | Library Name | Context7 ID |
|-----------|--------------|-------------|
| cert-manager | cert-manager | /cert-manager/website |
| ingress-nginx | ingress-nginx | /kubernetes/ingress-nginx |
| external-secrets | external-secrets | /external-secrets/external-secrets |
| external-dns | external-dns | /kubernetes-sigs/external-dns |
| prometheus | kube-prometheus-stack | /prometheus-community/helm-charts |

**NO hardcoded versions are provided intentionally.**

## Helm Chart Repositories

| Component | Repository URL |
|-----------|----------------|
| cert-manager | https://charts.jetstack.io |
| ingress-nginx | https://kubernetes.github.io/ingress-nginx |
| external-secrets | https://charts.external-secrets.io |
| external-dns | https://kubernetes-sigs.github.io/external-dns |
| prometheus-stack | https://prometheus-community.github.io/helm-charts |

## Flux Components (API Versions)

| Component | API Version | Notes |
|-----------|-------------|-------|
| HelmRelease | helm.toolkit.fluxcd.io/v2 | Stable since Flux 2.0 |
| HelmRepository | source.toolkit.fluxcd.io/v1 | Stable |
| Kustomization | kustomize.toolkit.fluxcd.io/v1 | Stable |
| GitRepository | source.toolkit.fluxcd.io/v1 | Stable |
| ImageRepository | image.toolkit.fluxcd.io/v1 | Stable |
| ImagePolicy | image.toolkit.fluxcd.io/v1 | Stable |
| ImageUpdateAutomation | image.toolkit.fluxcd.io/v1 | Stable |

## Deprecated API Versions

**DO NOT USE:**

| Deprecated | Replacement |
|------------|-------------|
| helm.toolkit.fluxcd.io/v2beta1 | helm.toolkit.fluxcd.io/v2 |
| helm.toolkit.fluxcd.io/v2beta2 | helm.toolkit.fluxcd.io/v2 |
| source.toolkit.fluxcd.io/v1beta2 | source.toolkit.fluxcd.io/v1 |
| kustomize.toolkit.fluxcd.io/v1beta2 | kustomize.toolkit.fluxcd.io/v1 |
| image.toolkit.fluxcd.io/v1beta2 | image.toolkit.fluxcd.io/v1 |
| external-secrets.io/v1beta1 | external-secrets.io/v1 |

## Upgrade Considerations

### cert-manager

- Check CRD changes between versions via release notes
- Apply new CRDs before upgrading HelmRelease
- Test with `kubectl apply --dry-run=server`

### external-secrets

- ESO v1 API is stable (no breaking changes expected)
- Check provider-specific changes in release notes

### ingress-nginx

- Check for deprecated annotations in release notes
- Verify IngressClass configuration
- Test with canary deployment

## Version Pinning Strategy

| Environment | Strategy | Example |
|-------------|----------|---------|
| Development | Flexible | `version: ">=X.0.0"` |
| Staging | Minor pinned | `version: "~X.Y.0"` |
| Production | Exact pinned | `version: "X.Y.Z"` |

**Get X.Y.Z from Context7 before applying.**
