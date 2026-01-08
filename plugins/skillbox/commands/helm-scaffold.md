---
name: helm-scaffold
description: Scaffold complete GitOps structure for a new application
---

# /helm-scaffold

Scaffold complete GitOps structure for a new application.

## Required Parameters

Before running, ensure you have:
- **appName**: Application name (kebab-case)
- **namespace**: Target namespace (dev, prod)
- **image**: Container image repository
- **secretPath**: AWS Secrets Manager path (e.g., `project/dev/app`)
- **ingressHost**: Ingress hostname (optional)

## What Gets Created

```
apps/
├── base/<appName>/
│   ├── helm.yaml              # Base HelmRelease
│   └── kustomization.yaml
├── dev/<appName>/
│   ├── values.yaml            # Dev values
│   ├── patches.yaml           # Image tag patch
│   ├── kustomization.yaml     # ConfigMap generator
│   ├── kustomizeconfig.yaml   # HelmRelease field refs
│   └── secrets/
│       └── <appName>.external.yaml  # ExternalSecret
└── prod/<appName>/
    └── (same structure as dev)
```

## Usage

```
/helm-scaffold appName=my-app namespace=dev image=123456.dkr.ecr.region.amazonaws.com/my-app secretPath=project/dev/my-app ingressHost=app.dev.example.com
```

## Steps

1. Create base HelmRelease in `apps/base/<appName>/`
2. Create dev overlay with values, patches, kustomization
3. Create dev ExternalSecret referencing ClusterSecretStore
4. Create prod overlay (similar to dev)
5. Run `/helm-validate` to verify

## Notes

- Uses `refreshPolicy: OnChange` for deterministic secret updates
- Assumes ClusterSecretStore `aws-secrets-manager` exists
- Namespace must have label `eso.fce.global/enabled: "true"`
