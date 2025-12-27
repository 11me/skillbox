---
name: flux-gitops-scaffold
description: >
  This skill should be used when the user asks to "create a GitOps project",
  "scaffold Flux project", "set up GitOps repository", "add Flux application",
  "configure image automation", "add infrastructure component to GitOps",
  "set up Flux with Kustomize", or mentions "Flux CD project structure".
  Provides scaffolding for Flux GitOps projects with multi-environment support,
  image automation, and External Secrets integration.
globs: ["**/clusters/**/*.yaml", "**/apps/**/*.yaml", "**/infra/**/*.yaml"]
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# Flux GitOps Scaffold

## Purpose

Scaffold and manage Flux CD GitOps repositories with:
- Multi-environment structure (dev/staging/prod)
- Infrastructure components (cert-manager, ingress-nginx, external-secrets)
- Application deployments with image automation
- External Secrets integration for secret management

**Complementary to:** `helm-chart-developer` skill (for Helm chart authoring).
This skill focuses on GitOps project structure and Flux-specific patterns.

## When to Use

Activate for:
- Creating new GitOps repositories from scratch
- Adding infrastructure components to existing GitOps projects
- Setting up application deployments with image automation
- Configuring Flux Kustomization dependencies
- Multi-environment deployment patterns

## Project Structure Pattern

```
gitops/
├── clusters/                    # Flux orchestration layer
│   ├── dev/
│   │   ├── 00-crds.yaml         # CRDs first (prune: false)
│   │   ├── 02-secrets-operator.yaml
│   │   ├── 05-ingress-nginx.yaml
│   │   ├── 06-cert-manager.yaml
│   │   ├── 99-apps.yaml         # Apps last (depends on infra)
│   │   ├── flux-system/
│   │   └── kustomization.yaml
│   └── prod/
├── infra/
│   ├── components/
│   │   ├── base/                # Shared HelmRepository + HelmRelease
│   │   │   ├── cert-manager/
│   │   │   │   ├── kustomization.yaml
│   │   │   │   └── helm.yaml    # HelmRepo + HelmRelease
│   │   │   └── ingress-nginx/
│   │   └── crds/                # CRD Kustomizations
│   │       ├── cert-manager/
│   │       └── external-secrets/
│   ├── dev/                     # Environment overlays (values only)
│   │   └── cert-manager/
│   │       ├── kustomization.yaml  # refs ../../components/base/cert-manager
│   │       └── values.yaml
│   └── prod/
├── apps/
│   ├── base/{app}/              # Base HelmRelease
│   └── {env}/{app}/             # Values + patches + image automation
└── charts/app/                  # Generic application chart
```

**Structure Principle:** Base + overlay pattern. `components/base/` contains shared HelmRelease, `{env}/` overlays provide only values.

See `references/project-structure.md` for detailed layout.

## Core Workflows

### 1. Initialize GitOps Project

To create a new GitOps project:

1. Gather requirements via AskUserQuestion:
   - Project name
   - Environments (dev, staging, prod)
   - Secrets provider (AWS/GCP/Azure/Vault)
   - Cloud region/project

2. Create directory structure per `references/project-structure.md`

3. Generate cluster orchestration files (numbered Kustomizations)

4. Copy generic Helm chart from `assets/charts/app/`

5. Fetch latest component versions via Context7

### 2. Add Infrastructure Component

To add infra component (cert-manager, ingress-nginx, etc.):

1. Get latest version via Context7:
   ```
   resolve-library-id: "cert-manager"
   get-library-docs: topic="installation"
   ```

2. If component has CRDs, create `infra/components/crds/{component}/`:
   - `kustomization.yaml` - Resources reference
   - `gitrepository.yaml` - Source for CRD manifests
   - `flux-kustomization.yaml` - `prune: false`, healthChecks

3. Create `infra/components/base/{component}/`:
   - `kustomization.yaml` - Resources reference
   - `helm.yaml` - HelmRepository + HelmRelease (single file)

4. Create `infra/{env}/{component}/` overlay for each environment:
   - `kustomization.yaml` - refs `../../components/base/{component}` + ConfigMapGenerator
   - `values.yaml` - Environment-specific values only

5. Add to `clusters/{env}/` orchestration with proper numbering

**Validation:** Run `kubectl kustomize infra/{env}/{component}` to validate before commit.

See `references/infra-components.md` for supported components.

### 3. Add Application

To add application with image automation:

1. Gather via AskUserQuestion:
   - Registry type (ECR/GCR/ACR/GHCR)
   - Image repository URL
   - Tag pattern (dev: run_id, prod: semver)

2. Create `apps/base/{app}/`:
   - `helm.yaml` - HelmRelease referencing `charts/app`
   - `kustomization.yaml`

3. Create `apps/{env}/{app}/`:
   - `kustomization.yaml` - Patches + ConfigMapGenerator
   - `values.yaml` - Environment values
   - `patches.yaml` - Image tag with automation marker
   - `kustomizeconfig.yaml` - ConfigMap reference
   - `secrets/` - ExternalSecret

4. Create image automation in `apps/{env}/`:
   - `image-automation.yaml` - ImageRepository + ImagePolicy + ImageUpdateAutomation

See `references/image-automation.md` for registry-specific patterns.

## API Versions

**Always use these API versions:**

| Resource | API Version |
|----------|-------------|
| HelmRelease | `helm.toolkit.fluxcd.io/v2` |
| HelmRepository | `source.toolkit.fluxcd.io/v1` |
| Kustomization (Flux) | `kustomize.toolkit.fluxcd.io/v1` |
| GitRepository | `source.toolkit.fluxcd.io/v1` |
| ImageRepository | `image.toolkit.fluxcd.io/v1` |
| ImagePolicy | `image.toolkit.fluxcd.io/v1` |
| ImageUpdateAutomation | `image.toolkit.fluxcd.io/v1` |
| ExternalSecret | `external-secrets.io/v1` |
| ClusterSecretStore | `external-secrets.io/v1` |

**CRITICAL:** Never use deprecated versions (v2beta1, v2beta2). If found, propose migration.

## Key Patterns

### HelmRelease with ConfigMap Values

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

### ConfigMap Generator (Kustomize)

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

### Flux Kustomization with Dependencies

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

### Image Automation Marker

```yaml
# In patches.yaml
spec:
  values:
    image:
      tag: dev-abc123-12345 # {"$imagepolicy": "flux-system:app-dev:tag"}
```

## Orchestration Rules

### Numbered Ordering

Use numbered prefixes for deployment order:
- `00-crds.yaml` - CRDs (always first, prune: false)
- `02-secrets-operator.yaml` - ESO operator
- `03-secrets-store.yaml` - ClusterSecretStore
- `05-ingress-nginx.yaml` - Ingress controller
- `06-cert-manager.yaml` - Certificate manager
- `07-cert-manager-issuer.yaml` - ClusterIssuer
- `99-apps.yaml` - Applications (always last)

### Dependencies

Always set `dependsOn` for:
- Apps depend on ingress-nginx, secrets-store
- cert-manager-issuer depends on cert-manager
- secrets-store depends on secrets-operator

### CRD Safety

For CRD Kustomizations:
```yaml
spec:
  prune: false  # Never delete CRDs
```

For HelmReleases with CRDs:
```yaml
spec:
  install:
    crds: Skip  # CRDs managed separately
  upgrade:
    crds: Skip
```

## Version Management

Use Context7 to fetch latest versions:

```
# Step 1: Resolve library ID
resolve-library-id: libraryName="cert-manager"

# Step 2: Get docs with version info
get-library-docs: context7CompatibleLibraryID="/jetstack/cert-manager", topic="helm installation"
```

See `references/version-matrix.md` for current versions.

## Validation

Before committing GitOps manifests, validate structure with `kubectl kustomize`:

```bash
# Validate infra component
kubectl kustomize infra/dev/cert-manager

# Validate apps
kubectl kustomize apps/dev

# Validate entire environment (from cluster kustomization)
kubectl kustomize clusters/dev
```

This catches:
- Missing files referenced in kustomization.yaml
- Invalid patches
- YAML syntax errors
- Invalid resource references

## Definition of Done

Before completing GitOps scaffolding:

1. **Structure**: All directories created per pattern
2. **Validation**: `kubectl kustomize` passes for all directories
3. **Dependencies**: `dependsOn` set correctly in Kustomizations
4. **Versions**: Latest versions fetched via Context7
5. **API Versions**: All using current stable APIs
6. **Values**: ConfigMapGenerator with `disableNameSuffixHash: true`
7. **Secrets**: ExternalSecret configured (not hardcoded secrets)
8. **Image Automation**: Markers set for automated updates

## Anti-Patterns

| Avoid | Instead |
|-------|---------|
| Hardcoded secrets in values | ExternalSecret + secretRef |
| `prune: true` for CRDs | `prune: false` to prevent deletion |
| Missing `dependsOn` | Always set dependencies |
| `crds: CreateReplace` | `crds: Skip` (manage separately) |
| `v2beta1`/`v2beta2` APIs | Use stable `v2`/`v1` APIs |
| Hash suffix on ConfigMaps | `disableNameSuffixHash: true` |
| Single values.yaml for all envs | Base + overlay pattern |

## Examples

Trigger phrases:
- "Create a new GitOps repository for my microservices"
- "Add cert-manager to my Flux project"
- "Set up image automation for my app"
- "Configure multi-environment deployment"
- "Add ingress-nginx infrastructure"
- "Set up External Secrets with AWS"

## Additional Resources

### Reference Files

For detailed patterns, consult:
- **`references/project-structure.md`** - Full directory layout
- **`references/image-automation.md`** - Registry-specific automation
- **`references/infra-components.md`** - Infrastructure patterns
- **`references/version-matrix.md`** - Current versions + Context7 usage

### Example Files

Working examples in `examples/`:
- **`cluster-kustomization.yaml`** - Flux Kustomization with dependencies
- **`helmrelease-base.yaml`** - Base HelmRelease pattern
- **`helmrelease-env-patch.yaml`** - Environment overlay
- **`infra-base-helm.yaml`** - Base HelmRepo + HelmRelease (single file)
- **`infra-overlay-kustomization.yaml`** - Environment overlay kustomization
- **`crds-*.yaml`** - CRD management patterns
- **`image-automation-ecr.yaml`** - ECR automation
- **`image-automation-ghcr.yaml`** - GHCR automation
- **`external-secret.yaml`** - ESO pattern

### Assets

Generic chart template in `assets/charts/app/` - copy to target project.

## Related Skills

- **`helm-chart-developer`** - Helm chart authoring and ESO integration
- **`conventional-commit`** - Commit message formatting
