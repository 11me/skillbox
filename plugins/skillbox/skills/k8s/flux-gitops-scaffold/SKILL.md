---
name: flux-gitops-scaffold
description: Scaffold Flux GitOps projects with multi-environment support. Use when creating GitOps projects, scaffolding Flux projects, setting up GitOps repositories, adding Flux applications, configuring image automation, adding infrastructure components, or setting up Flux with Kustomize.
allowed-tools: [Read, Write, Edit, Glob, Grep, Bash]
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
├── clusters/{env}/              # Flux orchestration layer
│   ├── kustomization.yaml       # Aggregates all Flux Kustomizations
│   ├── 00-crds.yaml             # CRDs (prune: false, wait: true)
│   ├── 01-controllers.yaml      # dependsOn: crds
│   ├── 02-cluster-configs.yaml  # dependsOn: controllers
│   ├── 03-services.yaml         # dependsOn: cluster-configs
│   ├── 99-apps.yaml             # dependsOn: services, cluster-configs
│   └── flux-system/
│
├── infra/
│   ├── base/
│   │   ├── cluster/
│   │   │   ├── controllers/     # cert-manager, ingress-nginx, ESO
│   │   │   └── configs/         # ClusterIssuer, ClusterSecretStore
│   │   └── services/            # redis, postgres (with configs/secrets)
│   ├── crds/                    # Vendored CRDs (applied first)
│   │   ├── kustomization.yaml   # Aggregates all CRD subdirs
│   │   ├── cert-manager/
│   │   └── external-secrets/
│   └── {env}/
│       ├── cluster/
│       │   ├── controllers/     # values only (ConfigMapGenerator)
│       │   └── configs/         # plain manifests
│       └── services/            # values + configs + secrets
│
├── apps/
│   ├── base/{app}/              # Base HelmRelease
│   └── {env}/{app}/             # values + configs + secrets
│       ├── configs/             # Plain ConfigMaps
│       └── secrets/             # ExternalSecrets
│
└── charts/app/                  # Generic application chart
```

**Structure Principle:** Base + overlay with explicit layering. Controllers → Configs → Services → Apps.

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

Three workflows depending on component type:

#### 2a. Add Controller (cert-manager, ingress-nginx, ESO)

1. **FIRST: Get latest version via Context7** (see Version Management section)
2. Vendor CRDs to `infra/crds/{component}/`:
   - `kustomization.yaml` - Resources reference
   - `crds.yaml` - Vendored from upstream (curl from release)
3. Create `infra/base/cluster/controllers/{component}/`:
   - `kustomization.yaml` - Resources reference
   - `helm.yaml` - HelmRepository + HelmRelease (installCRDs: false)
4. Create `infra/{env}/cluster/controllers/{component}/`:
   - `kustomization.yaml` - refs base + ConfigMapGenerator
   - `values.yaml` - Environment values (installCRDs: false)
5. Update `infra/crds/kustomization.yaml` aggregator
6. Update `infra/{env}/cluster/controllers/kustomization.yaml` aggregator

#### 2b. Add Cluster Config (ClusterIssuer, ClusterSecretStore)

1. Create `infra/base/cluster/configs/{component}/`:
   - `kustomization.yaml` - Resources reference
   - `{component}.yaml` - Plain manifest template
2. Create `infra/{env}/cluster/configs/{component}/`:
   - `kustomization.yaml` - refs base
   - `{component}.yaml` - Environment-specific manifest
3. Update `infra/{env}/cluster/configs/kustomization.yaml` aggregator

#### 2c. Add Service (redis, postgres)

1. **FIRST: Get latest version via Context7** (see Version Management section)
2. Create `infra/base/services/{component}/`:
   - `kustomization.yaml` - Resources reference
   - `helm.yaml` - HelmRepository + HelmRelease
3. Create `infra/{env}/services/{component}/`:
   - `kustomization.yaml` - refs base + ConfigMapGenerator + configs + secrets
   - `values.yaml` - With envFrom injection
   - `configs/{component}.config.yaml` - Plain ConfigMap
   - `secrets/{component}.external.yaml` - ExternalSecret
4. Update `infra/{env}/services/kustomization.yaml` aggregator

**Validation:** Run `kubectl kustomize infra/{env}/...` to validate before commit.

See `references/infra-components.md` for supported components.

### 3. Add Application

To add application with image automation:

1. Gather via AskUserQuestion:
   - Registry type (ECR/GCR/ACR/GHCR)
   - Image repository URL
   - Tag pattern (dev: run_id, prod: semver)

2. Create `apps/base/{app}/`:
   - `kustomization.yaml` - Resources reference
   - `helm.yaml` - HelmRelease referencing `charts/app`

3. Create `apps/{env}/{app}/`:
   - `kustomization.yaml` - refs base + ConfigMapGenerator + configs + secrets
   - `values.yaml` - With envFrom injection
   - `patches.yaml` - Image tag with automation marker
   - `kustomizeconfig.yaml` - ConfigMap name replacement
   - `configs/{app}.config.yaml` - Plain ConfigMap (non-sensitive)
   - `secrets/{app}.external.yaml` - ExternalSecret (sensitive)

4. Create image automation in `apps/{env}/`:
   - `image-automation.yaml` - ImageRepository + ImagePolicy + ImageUpdateAutomation

5. Update `apps/{env}/kustomization.yaml` aggregator

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

See **[`references/key-patterns.md`](references/key-patterns.md)** for detailed YAML patterns:
- HelmRelease with ConfigMap Values
- ConfigMap Generator (Kustomize overlay pattern)
- Flux Kustomization with Dependencies
- Image Automation Markers
- Configs & Secrets Pattern (ConfigMap, ExternalSecret, envFrom)

## Orchestration

Five Flux Kustomizations with explicit `dependsOn` chain:

| File | DependsOn | Path | Critical Settings |
|------|-----------|------|-------------------|
| `00-crds.yaml` | - | `./infra/crds` | prune: false, wait: true |
| `01-controllers.yaml` | crds | `./infra/{env}/cluster/controllers` | wait: true, timeout: 10m |
| `02-cluster-configs.yaml` | controllers | `./infra/{env}/cluster/configs` | wait: true, timeout: 5m |
| `03-services.yaml` | cluster-configs | `./infra/{env}/services` | wait: true, timeout: 10m |
| `99-apps.yaml` | services, cluster-configs | `./apps/{env}` | wait: true, timeout: 10m |

**Why this order:**
- CRDs must exist before controllers can install CRs
- Controllers must be running before configs (ClusterIssuer needs cert-manager)
- Cluster configs (ClusterSecretStore) needed before services can fetch secrets
- Services provide infrastructure for apps (redis, postgres)

### Aggregator Pattern

Each directory Flux points to **MUST** have `kustomization.yaml`:

```yaml
# infra/{env}/cluster/controllers/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - cert-manager
  - ingress-nginx
  - external-secrets
```

### CRD Management

**CRITICAL:** Vendor CRDs into repository. Do NOT use nested Flux Kustomizations for CRDs.

```bash
# Download vendored CRDs (use Context7 to get {VERSION} first!)
curl -sL https://github.com/cert-manager/cert-manager/releases/download/{VERSION}/cert-manager.crds.yaml \
  > infra/crds/cert-manager/crds.yaml
```

For HelmReleases with CRDs, disable CRD installation:
```yaml
# values.yaml
installCRDs: false
```

## Version Management (MANDATORY)

**CRITICAL:** Never hardcode versions. Always use Context7 to get current versions.

### Enforcement Layers

1. **SessionStart hook** — Reminds about Context7 when k8s/Flux project detected
2. **PreToolUse hook** — Checks versions in HelmRelease files:
   - Empty version (`version: ""`) → Blocks and requires Context7
   - Hardcoded version (`version: "1.2.3"`) → Asks for verification
3. **Documentation** — No hardcoded versions provided (only Context7 workflow)

### Workflow (REQUIRED for every HelmRelease)

Before creating ANY HelmRelease or vendoring CRDs:

```
# Step 1: Resolve library ID
Tool: resolve-library-id
Parameter: libraryName="{component}"

# Step 2: Query docs for version
Tool: query-docs
Parameters:
  libraryId: "/{org}/{project}"
  topic: "helm chart installation version"

# Step 3: Use version from documentation
```

### Example: cert-manager

```
Tool: resolve-library-id
libraryName: "cert-manager"
# Returns: /jetstack/cert-manager

Tool: query-docs
libraryId: "/jetstack/cert-manager"
topic: "helm installation version"
# Extract version from documentation
```

### Guardrails

**NEVER:**
- Copy versions from examples or documentation files
- Use versions from memory or previous sessions
- Hardcode versions without Context7 verification

**ALWAYS:**
- Fetch version via Context7 for EVERY new HelmRelease
- Verify version is current before deployment
- Use empty `version: ""` as placeholder until Context7 fetched

See `references/version-matrix.md` for Context7 library IDs per component.

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
| `crds: CreateReplace` | `crds: Skip` + vendored CRDs |
| `installCRDs: true` in values | `installCRDs: false` (CRDs managed separately) |
| Nested Flux Kustomization for CRDs | Vendor CRDs into repo (race condition!) |
| Missing `wait: true` | Always use `wait: true` + `timeout` |
| Missing aggregator kustomization.yaml | Every Flux path needs kustomization.yaml |
| `v2beta1`/`v2beta2` APIs | Use stable `v2`/`v1` APIs |
| Hash suffix on ConfigMaps | `disableNameSuffixHash: true` |
| `{name}-{env}` suffix | Skip suffix if env = namespace |

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

**Orchestration:**
- **`orchestration-kustomizations.yaml`** - All 5 Flux Kustomizations with dependsOn chain

**Controllers & Configs:**
- **`infra-base-helm.yaml`** - Base HelmRepo + HelmRelease for controller
- **`cluster-controller-overlay.yaml`** - Controller overlay (values only)
- **`cluster-config-overlay.yaml`** - Cluster config overlay (ClusterIssuer)

**Services & Apps (with configs/secrets):**
- **`service-overlay-full.yaml`** - Service with configs/secrets
- **`app-overlay-full.yaml`** - App with configs/secrets
- **`config-configmap.yaml`** - Plain ConfigMap pattern
- **`values-with-envfrom.yaml`** - values.yaml with envFrom injection

**CRDs & Image Automation:**
- **`crds-kustomization.yaml`** - Vendored CRDs aggregator
- **`image-automation-ecr.yaml`** - ECR automation
- **`image-automation-ghcr.yaml`** - GHCR automation
- **`external-secret.yaml`** - ESO pattern

### Assets

Generic chart template in `assets/charts/app/` - copy to target project.

## Related Skills

- **`helm-chart-developer`** - Helm chart authoring and ESO integration
- **`conventional-commit`** - Commit message formatting
