# Image Automation Reference

## Overview

Flux Image Automation consists of three CRDs:
1. **ImageRepository** - Scans container registry for new tags
2. **ImagePolicy** - Selects the latest tag based on policy
3. **ImageUpdateAutomation** - Commits selected tag to git

## ImageRepository by Registry

### AWS ECR

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  image: 123456789012.dkr.ecr.region.amazonaws.com/app-name
  interval: 1m
  provider: aws
```

**Requirements:**
- Flux controllers need IAM role with ECR read access
- Use IRSA (IAM Roles for Service Accounts) in EKS

### Google GCR / Artifact Registry

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  image: gcr.io/project-id/app-name
  # or: region-docker.pkg.dev/project-id/repo/app-name
  interval: 1m
  provider: gcp
```

**Requirements:**
- Workload Identity configured
- Service account with Artifact Registry Reader role

### Azure ACR

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  image: myregistry.azurecr.io/app-name
  interval: 1m
  provider: azure
```

**Requirements:**
- Managed Identity with AcrPull role
- AAD Pod Identity or Workload Identity

### GitHub Container Registry (GHCR)

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  image: ghcr.io/org/app-name
  interval: 1m
  secretRef:
    name: ghcr-auth
```

**Requirements:**
- ImagePullSecret with PAT (read:packages scope)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ghcr-auth
  namespace: flux-system
type: kubernetes.io/dockerconfigjson
stringData:
  .dockerconfigjson: |
    {"auths":{"ghcr.io":{"auth":"base64(username:PAT)"}}}
```

### Docker Hub

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  image: docker.io/org/app-name
  interval: 1m
  secretRef:
    name: dockerhub-auth
```

## ImagePolicy Patterns

### Development: Numerical (by run_id)

Select latest by CI run ID (highest number wins):

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImagePolicy
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  imageRepositoryRef:
    name: app-name-dev
  filterTags:
    pattern: '^dev-[0-9a-fA-F]{7,40}-(?P<run>[0-9]+)$'
    extract: '$run'
  policy:
    numerical:
      order: asc  # Higher number = newer
```

**Tag format:** `dev-{git_sha}-{run_id}`
**Example:** `dev-abc1234-12345678`

### Production: Semver

Select latest semantic version:

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImagePolicy
metadata:
  name: app-name-prod
  namespace: flux-system
spec:
  imageRepositoryRef:
    name: app-name-prod
  filterTags:
    pattern: '^v(?P<version>[0-9]+\.[0-9]+\.[0-9]+)$'
    extract: '$version'
  policy:
    semver:
      range: '>=1.0.0'
```

**Tag format:** `v{major}.{minor}.{patch}`
**Example:** `v1.2.3`

### Staging: Alphabetical (by SHA)

Select latest by git SHA (alphabetically):

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImagePolicy
metadata:
  name: app-name-staging
  namespace: flux-system
spec:
  imageRepositoryRef:
    name: app-name-staging
  filterTags:
    pattern: '^staging-(?P<sha>[0-9a-fA-F]{7,40})$'
    extract: '$sha'
  policy:
    alphabetical:
      order: desc  # Latest SHA alphabetically
```

## ImageUpdateAutomation

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageUpdateAutomation
metadata:
  name: app-name-auto-dev
  namespace: flux-system
spec:
  interval: 5m
  sourceRef:
    kind: GitRepository
    name: flux-system
  git:
    checkout:
      ref:
        branch: main
    commit:
      author:
        name: fluxbot
        email: fluxbot@example.com
      messageTemplate: |
        chore: automated image update

        Automation: {{ .AutomationObject }}

        {{ range .Changed.Changes -}}
        - {{ .OldValue }} -> {{ .NewValue }}
        {{ end -}}
    push:
      branch: main
  update:
    strategy: Setters
    path: ./apps/dev  # Scope to environment
```

## Image Tag Markers

### In patches.yaml

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: app-name
spec:
  values:
    image:
      tag: dev-abc1234-12345678 # {"$imagepolicy": "flux-system:app-name-dev:tag"}
```

### Marker Format

```
# {"$imagepolicy": "NAMESPACE:POLICY_NAME:FIELD"}
```

**Fields:**
- `tag` - Just the tag portion
- `name` - Full image name with tag

### In raw manifests

```yaml
spec:
  containers:
    - name: app
      image: registry/app:v1.0.0 # {"$imagepolicy": "flux-system:app-name-prod"}
```

## Complete Example: Dev Environment

```yaml
# apps/dev/image-automation.yaml
---
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  image: 123456789012.dkr.ecr.me-central-1.amazonaws.com/app-name
  interval: 1m
  provider: aws

---
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImagePolicy
metadata:
  name: app-name-dev
  namespace: flux-system
spec:
  imageRepositoryRef:
    name: app-name-dev
  filterTags:
    pattern: '^dev-[0-9a-fA-F]{7,40}-(?P<run>[0-9]+)$'
    extract: '$run'
  policy:
    numerical:
      order: asc

---
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageUpdateAutomation
metadata:
  name: app-name-auto-dev
  namespace: flux-system
spec:
  interval: 5m
  sourceRef:
    kind: GitRepository
    name: flux-system
  git:
    checkout:
      ref:
        branch: main
    commit:
      author:
        name: fluxbot
        email: fluxbot@example.com
      messageTemplate: |
        chore: automated image update

        Automation: {{ .AutomationObject }}
    push:
      branch: main
  update:
    strategy: Setters
    path: ./apps/dev
```

## Troubleshooting

### Check ImageRepository status

```bash
kubectl get imagerepository -n flux-system
flux get image repository
```

### Check ImagePolicy status

```bash
kubectl get imagepolicy -n flux-system
flux get image policy
```

### Check latest selected tag

```bash
kubectl get imagepolicy app-name-dev -n flux-system -o jsonpath='{.status.latestImage}'
```

### Check automation logs

```bash
kubectl logs -n flux-system deploy/image-automation-controller
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| No tags found | Wrong regex pattern | Test pattern with `flux get image policy --all` |
| Auth failed (ECR) | Missing IAM permissions | Check IRSA setup |
| Auth failed (GHCR) | Invalid PAT | Regenerate PAT with correct scope |
| No commits | Marker not found | Check `$imagepolicy` comment format |
| Wrong tag selected | Policy misconfigured | Verify `extract` and `policy` settings |
