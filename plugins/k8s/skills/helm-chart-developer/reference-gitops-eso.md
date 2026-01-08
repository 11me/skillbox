# GitOps + External Secrets Reference

## Repository Layout

```
gitops/
├── charts/app/              # Universal Helm chart
├── apps/
│   ├── base/                # Base HelmRelease
│   ├── dev/                 # Dev overlay: values + patches + ExternalSecret
│   └── prod/                # Prod overlay: values + patches + ExternalSecret
└── infra/                   # external-secrets-operator, cert-manager, etc.
```

## Values Composition Pipeline

```
┌─────────────────────────────────────────────────────────────────┐
│                    VALUES MERGE ORDER                            │
├─────────────────────────────────────────────────────────────────┤
│  1. Chart defaults      → charts/app/values.yaml                │
│  2. Environment values  → apps/dev/app/values.yaml              │
│  3. ConfigMap           → Kustomize configMapGenerator          │
│  4. HelmRelease valuesFrom → references ConfigMap               │
│  5. HelmRelease values  → inline patches (highest priority)     │
└─────────────────────────────────────────────────────────────────┘
```

**Important**: Flux merges `valuesFrom` entries in order, then applies `spec.values` on top.

## Values API Contract

```yaml
# Full values.yaml API

nameOverride: ""
fullnameOverride: ""

image:
  repository: ""
  tag: ""
  pullPolicy: IfNotPresent

serviceAccount:
  create: true
  name: ""
  annotations: {}

replicaCount: 1

strategy:
  type: RollingUpdate

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    memory: 256Mi

# Secrets configuration
secrets:
  # Mode A: Reference existing secret (recommended for GitOps)
  existingSecretName: ""

  # Mode B: Chart-managed ExternalSecret (optional)
  externalSecret:
    enabled: false
    refreshInterval: 1h
    refreshPolicy: OnChange  # OnChange | CreatedOnce | Periodic
    secretStoreRef:
      kind: ClusterSecretStore
      name: ""  # required when enabled
    dataFrom:
      extractKey: ""  # e.g., fce/dev/app
    target:
      name: ""  # defaults to fullname
      creationPolicy: Owner

  # Injection settings
  inject:
    envFrom: true  # inject as envFrom.secretRef

# Additional environment
env: []
envFrom: []

# Ports
ports:
  - name: http
    containerPort: 8080
    protocol: TCP

# Service
service:
  enabled: true
  type: ClusterIP
  port: 80
  targetPort: http

# Ingress
ingress:
  enabled: false
  className: nginx
  annotations: {}
  hosts: []
  tls: []

# Probes
livenessProbe:
  enabled: false
  httpGet:
    path: /health
    port: http

readinessProbe:
  enabled: false
  httpGet:
    path: /ready
    port: http

# Autoscaling
autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80

# Pod Disruption Budget
podDisruptionBudget:
  enabled: false
  minAvailable: 1

# Scheduling
nodeSelector: {}
tolerations: []
affinity: {}
topologySpreadConstraints: []
```

## Helper: app.secretName

```tpl
{{- define "app.secretName" -}}
{{- $existing := .Values.secrets.existingSecretName | default "" -}}
{{- $target := .Values.secrets.externalSecret.target.name | default "" -}}
{{- if $existing -}}
{{- $existing -}}
{{- else if and (.Values.secrets.externalSecret.enabled) $target -}}
{{- $target -}}
{{- else if .Values.secrets.externalSecret.enabled -}}
{{- include "app.fullname" . -}}
{{- else -}}
{{- "" -}}
{{- end -}}
{{- end }}
```

## Deployment: Secret Injection

```yaml
{{- $secretName := include "app.secretName" . -}}
spec:
  template:
    spec:
      containers:
        - name: {{ .Chart.Name }}
          envFrom:
            {{- if .Values.envFrom }}
            {{- toYaml .Values.envFrom | nindent 12 }}
            {{- end }}
            {{- if and $secretName .Values.secrets.inject.envFrom }}
            - secretRef:
                name: {{ $secretName }}
            {{- end }}
```

## ESO refreshPolicy

| Policy | Behavior | Use Case |
|--------|----------|----------|
| `OnChange` | Updates when ExternalSecret manifest changes | GitOps (recommended) |
| `CreatedOnce` | Never updates after creation | Immutable credentials |
| `Periodic` | Updates on interval (refreshInterval) | Legacy, auto-rotation |

## Flux/ESO Ordering

To avoid race conditions (CRDs not ready when CRs apply):

```yaml
# infra/dev/kustomization.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: infra-secrets-operator
spec:
  # ... ESO operator install

---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: infra-secrets-store
spec:
  dependsOn:
    - name: infra-secrets-operator  # Wait for CRDs
  # ... ClusterSecretStore
```

## Namespace Access Control

Namespace must have label for ClusterSecretStore access:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dev
  labels:
    eso.fce.global/enabled: "true"
```
