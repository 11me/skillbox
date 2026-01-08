---
name: helm-chart-developer
description: Build/refactor production Helm charts and GitOps deploy via Flux HelmRelease + Kustomize overlays. Includes External Secrets Operator patterns (ExternalSecret per app/env + ClusterSecretStore) and optional chart-managed ExternalSecret templates. Use when authoring Helm charts, converting raw manifests to Helm, designing values/schema, or debugging helm template/lint/dry-run issues.
allowed-tools: [Read, Grep, Glob, Write, Edit, Bash]
---

# Helm Chart Developer

## Purpose / When to Use

Use this skill when:
- Creating new Helm charts from scratch
- Converting raw Kubernetes manifests to Helm
- Designing values.yaml API and schema
- Setting up GitOps deployment with Flux HelmRelease
- Integrating External Secrets Operator (ESO)
- Debugging `helm template`, `helm lint`, `helm install --dry-run`
- Setting up multi-environment overlays (dev/prod)

## Version & API Compatibility

**Always use these API versions:**

| Component | API Version | Notes |
|-----------|-------------|-------|
| Helm Chart | `apiVersion: v2` | Helm 3+ charts |
| Flux HelmRelease | `helm.toolkit.fluxcd.io/v2` | Current stable |
| Flux HelmRepository | `source.toolkit.fluxcd.io/v1` | Current stable |
| Flux Kustomization | `kustomize.toolkit.fluxcd.io/v1` | Current stable |
| ExternalSecret | `external-secrets.io/v1` | ESO v1 API |
| ClusterSecretStore | `external-secrets.io/v1` | With `spec.conditions[].namespaceSelector` |

**Target baseline:** Kubernetes 1.29+, Helm 3.14+, Flux v2.3+, ESO 0.10+

**Rule:** If adding manifests, ALWAYS use these API versions.
If CRD/apiVersion mismatch detected in repo, STOP and propose migration plan.

See [VERSIONS.md](VERSIONS.md) for full compatibility matrix.

## Definition of Done (DoD)

Before completing any Helm chart work:

1. **Linting**: `helm lint .` passes
2. **Template rendering**: `helm template <release> .` succeeds
3. **Schema validation** (optional): `helm template . | kubeconform -strict`
4. **Dry-run**: `helm install <release> . --dry-run --debug` works
5. **Two secrets modes validated**:
   - GitOps mode: `--set secrets.existingSecretName=<name>`
   - Chart-managed ESO: `--set secrets.externalSecret.enabled=true`
6. **API versions match** the compatibility table above

Run `/helm-validate` to execute all checks.

## Step-by-Step Workflow

### 1. Chart Structure

```
charts/app/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── externalsecret.yaml  # optional, gated
│   ├── hpa.yaml
│   ├── pdb.yaml
│   └── serviceaccount.yaml
```

### 2. Values API Contract

See [reference-gitops-eso.md](reference-gitops-eso.md) for full details.

Key sections:
- `image`: repository, tag, pullPolicy
- `secrets`: existingSecretName, externalSecret.*, inject.envFrom
- `service`: enabled, type, port
- `ingress`: enabled, className, hosts, tls
- `resources`: requests, limits
- `autoscaling`: enabled, minReplicas, maxReplicas

### 3. Secrets Integration

**Mode A: ExternalSecret in Overlay (Recommended)**

Chart only references secret by name:
```yaml
# values.yaml
secrets:
  existingSecretName: "app-secrets"
  inject:
    envFrom: true
```

ExternalSecret lives in GitOps overlay, not in chart.

**Mode B: Chart-Managed ExternalSecret (Optional)**

Chart renders ExternalSecret when enabled:
```yaml
secrets:
  externalSecret:
    enabled: true
    refreshInterval: 1h
    refreshPolicy: OnChange
    secretStoreRef:
      kind: ClusterSecretStore
      name: aws-secrets-manager
    dataFrom:
      extractKey: "fce/dev/app"
    target:
      name: "app-secrets"
      creationPolicy: Owner
```

### Secrets Determinism (ESO)

**refreshPolicy options:**

| Policy | Behavior | Use Case |
|--------|----------|----------|
| `OnChange` | Updates when ExternalSecret manifest changes | GitOps (default, recommended) |
| `CreatedOnce` | Never updates after creation | Immutable credentials |
| `Periodic` | Updates on interval (refreshInterval) | Legacy, auto-rotation |

**Default rule:** Use `refreshPolicy: OnChange` for predictable GitOps-driven updates.

**Manual refresh (debug/runbook):**
```bash
kubectl annotate es <name> force-sync=$(date +%s) --overwrite
```

### 4. Flux + Kustomize Recipe

**Values Composition Order** (important!):
1. Chart defaults (`charts/app/values.yaml`)
2. Environment values (`apps/dev/app/values.yaml`)
3. ConfigMap via Kustomize generator
4. HelmRelease `valuesFrom` references ConfigMap
5. HelmRelease `spec.values` patches (highest priority)

See snippets:
- [flux-helmrelease.base.yaml](snippets/flux-helmrelease.base.yaml)
- [flux-helmrelease.dev.patch.yaml](snippets/flux-helmrelease.dev.patch.yaml)
- [kustomize.configmapgenerator.yaml](snippets/kustomize.configmapgenerator.yaml)

**Critical Kustomize settings:**
```yaml
generatorOptions:
  disableNameSuffixHash: true  # MUST have, otherwise names change on every apply
```

### Flux Ordering

Use `spec.dependsOn` in HelmRelease when app depends on:
- CRDs (external-secrets, cert-manager)
- ExternalSecrets/SecretStores
- Ingress controllers, databases

Example:
```yaml
spec:
  dependsOn:
    - name: external-secrets
      namespace: external-secrets
```

### 5. ESO Patterns

**ClusterSecretStore** (cluster-wide):
```yaml
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  conditions:
    - namespaceSelector:
        matchLabels:
          eso.fce.global/enabled: "true"
  provider:
    aws:
      service: SecretsManager
      region: me-central-1
```

**ExternalSecret** (per app/env):
```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
spec:
  refreshPolicy: OnChange  # GitOps-deterministic
  secretStoreRef:
    kind: ClusterSecretStore
    name: aws-secrets-manager
  dataFrom:
    - extract:
        key: fce/dev/app
  target:
    name: app-secrets
    creationPolicy: Owner
```

### 6. Pitfalls

- **CRD ordering**: ESO CRDs must exist before ExternalSecret. Use Flux Kustomization `dependsOn`.
- **OpenAPI validation**: Use `install.disableOpenAPIValidation: true` in HelmRelease if needed.
- **NEVER put secrets in values.yaml**: No passwords, tokens, API keys, credentials.
  Only references to Secret/ExternalSecret names. This is **non-negotiable**.

## Examples

Prompts that should activate this skill:

1. "Create a Helm chart for my Node.js app"
2. "Convert these Kubernetes manifests to Helm"
3. "Add External Secrets integration to my chart"
4. "Set up Flux HelmRelease for my app"
5. "Debug why helm template is failing"
6. "Design values.yaml schema for multi-environment deployment"
7. "Add ingress with TLS to my Helm chart"
8. "Integrate AWS Secrets Manager with my chart"
9. "Set up Kustomize overlays for dev and prod"
10. "Fix helm lint errors in my chart"

## Anti-Patterns

| ❌ Avoid | ✅ Instead |
|----------|-----------|
| Hardcode secrets in values.yaml | Use ExternalSecret with secretStoreRef |
| Single values.yaml for all envs | Kustomize overlays per environment |
| HelmRelease `v2beta1` / `v2beta2` API | Use `helm.toolkit.fluxcd.io/v2` |
| Manual Secret creation | ESO with ClusterSecretStore |
| Inline helm values in HelmRelease | External values.yaml + patches via valuesFrom |
| Skip `helm lint` / `helm template` | Always validate with /helm-validate |
| `refreshPolicy: Periodic` for GitOps | Use `refreshPolicy: OnChange` for determinism |
| Kustomize `namesSuffixHash: true` | Always `disableNameSuffixHash: true` for ConfigMaps |
| Put ExternalSecret in Helm chart | Put ExternalSecret in GitOps overlay (Mode A) |
| Skip `dependsOn` for CRDs | Use `spec.dependsOn` for external-secrets, cert-manager |

## Related Files

- [VERSIONS.md](VERSIONS.md) - Version compatibility matrix
- [reference-gitops-eso.md](reference-gitops-eso.md) - Full GitOps + ESO reference
- [snippets/](snippets/) - Ready-to-use YAML snippets

## Version History

- 1.1.0 — Add version compatibility, secrets determinism, Flux ordering
- 1.0.0 — Initial release with Flux + ESO patterns
