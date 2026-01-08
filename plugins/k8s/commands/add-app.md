---
name: add-app
description: Add application to GitOps project with image automation
argument-hint: "<app-name> [namespace]"
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
---

# Add Application to GitOps

This command adds an application to an existing Flux GitOps project with image automation support.

## Workflow

### Step 1: Gather Application Information

Use AskUserQuestion to collect:

1. **Registry type** - ECR, GCR, ACR, or GHCR
2. **Image URL** - Full image path (e.g., `123456789.dkr.ecr.us-east-1.amazonaws.com/myapp`)
3. **Tag pattern** - How tags are formatted:
   - Numerical (dev): `^main-[a-f0-9]+-(?P<ts>[0-9]+)$`
   - Semver (prod): `^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)$`
4. **Namespace** - Target Kubernetes namespace
5. **Port** - Application port (default: 8080)
6. **Enable ingress** - Yes/No, and hostname if yes

### Step 2: Detect Project Structure

Use Glob to find existing GitOps structure:

```
**/apps/base/*/kustomization.yaml
**/clusters/*/99-apps.yaml
```

Identify:
- Available environments
- Existing applications
- Project root and charts location

### Step 3: Create Base Application

Create `apps/base/{name}/`:

**kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {namespace}
resources:
  - helmrelease.yaml
```

**helmrelease.yaml:**
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: {name}
spec:
  interval: 5m
  chart:
    spec:
      chart: ./charts/app
      sourceRef:
        kind: GitRepository
        name: flux-system
        namespace: flux-system
      reconcileStrategy: Revision
  valuesFrom:
    - kind: ConfigMap
      name: {name}-values
```

### Step 4: Create Environment Overlays

For each environment, create `apps/{env}/{name}/`:

**kustomization.yaml:**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {namespace}
resources:
  - ../../base/{name}
  - image-automation.yaml
patches:
  - path: patches.yaml
configMapGenerator:
  - name: {name}-values
    files:
      - values.yaml
generatorOptions:
  disableNameSuffixHash: true
configurations:
  - kustomizeconfig.yaml
```

**kustomizeconfig.yaml:**
```yaml
images:
  - path: spec/values/image
    kind: HelmRelease
```

**patches.yaml:**
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: {name}
spec:
  values:
    image:
      repository: {image-url}  # {"$imagepolicy": "flux-system:{name}:name"}
      tag: latest  # {"$imagepolicy": "flux-system:{name}:tag"}
```

**values.yaml:**
```yaml
replicaCount: 1

ports:
  - containerPort: {port}
    name: http
    protocol: TCP

service:
  enabled: true
  port: 80
  targetPort: {port}

ingress:
  enabled: {ingress-enabled}
  className: nginx
  hosts:
    - host: {hostname}
      paths:
        - path: /
          pathType: Prefix

livenessProbe:
  enabled: true
  httpGet:
    path: /health
    port: {port}
  initialDelaySeconds: 10

readinessProbe:
  enabled: true
  httpGet:
    path: /ready
    port: {port}
  initialDelaySeconds: 5

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Step 5: Configure Image Automation

Create `apps/{env}/{name}/image-automation.yaml`:

**For ECR:**
```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: {name}
  namespace: flux-system
spec:
  interval: 1m
  image: {image-url}
  provider: aws
---
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImagePolicy
metadata:
  name: {name}
  namespace: flux-system
spec:
  imageRepositoryRef:
    name: {name}
  filterTags:
    pattern: '{tag-pattern}'
    extract: '$ts'  # or '$major.$minor.$patch' for semver
  policy:
    numerical:
      order: asc
```

**For GHCR:**
```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageRepository
metadata:
  name: {name}
  namespace: flux-system
spec:
  interval: 1m
  image: {image-url}
  secretRef:
    name: ghcr-auth
```

Use `references/image-automation.md` for other registry types.

### Step 6: Create External Secret (Optional)

If application needs secrets, create `apps/{env}/{name}/external-secret.yaml`:

```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {name}
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: cluster-secret-store
    kind: ClusterSecretStore
  target:
    name: {name}-secrets
  data:
    - secretKey: DATABASE_URL
      remoteRef:
        key: {env}/{name}/database-url
```

Update HelmRelease to reference secret:
```yaml
envFrom:
  - secretRef:
      name: {name}-secrets
```

### Step 7: Update Cluster Orchestration

Update `clusters/{env}/99-apps.yaml` to include new application path:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: apps
  namespace: flux-system
spec:
  path: ./apps/{env}  # This path includes all apps
```

Or if using per-app Kustomizations, create `clusters/{env}/99-{name}.yaml`.

### Step 8: Configure Image Update Automation

Ensure `ImageUpdateAutomation` exists for the environment:

```yaml
apiVersion: image.toolkit.fluxcd.io/v1
kind: ImageUpdateAutomation
metadata:
  name: {env}-automation
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
        name: fluxcdbot
        email: fluxcdbot@users.noreply.github.com
      messageTemplate: 'chore({env}): update images'
    push:
      branch: main
  update:
    path: ./apps/{env}
    strategy: Setters
```

## Output

After completion, provide:

1. Summary of created files
2. Image automation configuration summary
3. Commands to verify:
   ```bash
   flux reconcile kustomization apps --with-source
   flux get image repository {name}
   flux get image policy {name}
   ```
4. Reminder about secrets configuration if applicable

## References

- Load `references/image-automation.md` for registry-specific patterns
- Load `references/project-structure.md` for directory layout
- Use `examples/image-automation-ecr.yaml` or `examples/image-automation-ghcr.yaml`
- Use `examples/external-secret.yaml` for secrets pattern
