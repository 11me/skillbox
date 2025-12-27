# Flux GitOps Migration Analysis Report

**Repository:** `/home/user/repos/my-gitops`
**Structure Type:** partial
**Environments:** dev, staging, prod

## Summary

- Total HelmReleases: 8
- Total Issues: 12
- High Severity: 5
- Migration Required: Yes

## Components

### Controllers
- cert-manager
- ingress-nginx
- external-secrets

### Configs
- cluster-issuer

### Services
- redis
- postgres

### Apps
- api
- frontend

## Issues

🔴 **no_values_from**
   - Component: cert-manager
   - Path: `infra/base/cert-manager/helm.yaml`
   - Fix: Add valuesFrom: ConfigMap reference

🔴 **no_values_from**
   - Component: ingress-nginx
   - Path: `infra/base/ingress-nginx/helm.yaml`
   - Fix: Add valuesFrom: ConfigMap reference

🔴 **missing_crds_skip**
   - Component: cert-manager
   - Path: `infra/base/cert-manager/helm.yaml`
   - Fix: Add crds: Skip and vendor CRDs separately

🔴 **missing_crds_skip**
   - Component: external-secrets
   - Path: `infra/base/external-secrets/helm.yaml`
   - Fix: Add crds: Skip and vendor CRDs separately

🔴 **missing_crds_dir**
   - Components: cert-manager, external-secrets
   - Fix: Create infra/crds/ with vendored CRDs

🟡 **inline_values**
   - Component: redis
   - Path: `infra/base/redis/helm.yaml`
   - Fix: Extract inline values to values.yaml

🟡 **inline_values**
   - Component: postgres
   - Path: `infra/base/postgres/helm.yaml`
   - Fix: Extract inline values to values.yaml

🟡 **missing_aggregator**
   - Path: `infra/dev/cluster/controllers`
   - Fix: Create kustomization.yaml aggregator

🟡 **missing_aggregator**
   - Path: `infra/staging/cluster/controllers`
   - Fix: Create kustomization.yaml aggregator

🟡 **missing_aggregator**
   - Path: `infra/prod/cluster/controllers`
   - Fix: Create kustomization.yaml aggregator

🟡 **patches_for_values**
   - Component: cert-manager
   - Path: `infra/dev/cert-manager/patch.yaml`
   - Fix: Convert to ConfigMapGenerator

🟡 **patches_for_values**
   - Component: ingress-nginx
   - Path: `infra/dev/ingress-nginx/patch.yaml`
   - Fix: Convert to ConfigMapGenerator

## HelmReleases

| Name | Path | valuesFrom | CRDs Skip |
|------|------|------------|-----------|
| cert-manager | `infra/base/cert-manager/helm.yaml` | ❌ | ❌ |
| ingress-nginx | `infra/base/ingress-nginx/helm.yaml` | ❌ | ✅ |
| external-secrets | `infra/base/external-secrets/helm.yaml` | ❌ | ❌ |
| redis | `infra/base/redis/helm.yaml` | ❌ | ✅ |
| postgres | `infra/base/postgres/helm.yaml` | ❌ | ✅ |
| api | `apps/base/api/helm.yaml` | ✅ | ✅ |
| frontend | `apps/base/frontend/helm.yaml` | ✅ | ✅ |

## Migration Plan

### Phase 1: Vendor CRDs

```bash
# cert-manager
curl -sL "https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.crds.yaml" \
  > infra/crds/cert-manager/crds.yaml

# external-secrets
curl -sL "https://raw.githubusercontent.com/external-secrets/external-secrets/v0.9.0/deploy/crds/bundle.yaml" \
  > infra/crds/external-secrets/crds.yaml
```

### Phase 2: Update HelmReleases

Add `valuesFrom` and `crds: Skip` to all HelmRelease files in base.

### Phase 3: Convert Overlays

Replace patch files with ConfigMapGenerator + values.yaml pattern.

### Phase 4: Create Aggregators

Create `kustomization.yaml` in:
- `infra/dev/cluster/controllers/`
- `infra/staging/cluster/controllers/`
- `infra/prod/cluster/controllers/`

### Phase 5: Validate

```bash
kubectl kustomize infra/dev/cluster/controllers
kubectl kustomize infra/staging/cluster/controllers
kubectl kustomize infra/prod/cluster/controllers
```

## JSON Report

```json
{
  "repository": "/home/user/repos/my-gitops",
  "structure_type": "partial",
  "environments": ["dev", "staging", "prod"],
  "components": {
    "controllers": ["cert-manager", "ingress-nginx", "external-secrets"],
    "configs": ["cluster-issuer"],
    "services": ["redis", "postgres"],
    "apps": ["api", "frontend"]
  },
  "helm_releases": [
    {
      "name": "cert-manager",
      "path": "infra/base/cert-manager/helm.yaml",
      "chart": "cert-manager",
      "version": "1.14.0",
      "has_values_from": false,
      "has_crds_skip": false,
      "has_inline_values": false
    },
    {
      "name": "ingress-nginx",
      "path": "infra/base/ingress-nginx/helm.yaml",
      "chart": "ingress-nginx",
      "version": "4.9.0",
      "has_values_from": false,
      "has_crds_skip": true,
      "has_inline_values": false
    }
  ],
  "issues": [
    {
      "type": "no_values_from",
      "component": "cert-manager",
      "path": "infra/base/cert-manager/helm.yaml",
      "severity": "high",
      "fix": "Add valuesFrom: ConfigMap reference"
    },
    {
      "type": "missing_crds_dir",
      "components": ["cert-manager", "external-secrets"],
      "severity": "high",
      "fix": "Create infra/crds/ with vendored CRDs"
    }
  ],
  "summary": {
    "total_helm_releases": 8,
    "total_issues": 12,
    "high_severity": 5,
    "migration_required": true
  }
}
```
