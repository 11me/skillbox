---
name: init
description: Initialize OpenAPI spec structure for a Go project
---

# /openapi-init

Initialize a modular OpenAPI 3.x specification structure with Redocly linting and oapi-codegen configuration.

## Usage

```
/openapi-init [api-version]
```

## Arguments

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `api-version` | No | `v1` | API version (v1, v2, etc.) |

## Prerequisites

- Go project with `go.mod`
- Node.js (for Redocly CLI)

## Generated Structure

```
project/
├── api/{version}/
│   ├── openapi.yaml              # Main spec with $ref
│   ├── paths/
│   │   └── _index.yaml           # Paths aggregator
│   ├── components/
│   │   ├── schemas/_index.yaml   # Schemas aggregator
│   │   ├── parameters/_index.yaml
│   │   ├── responses/_index.yaml
│   │   └── requests/_index.yaml
│   └── .redocly.yaml             # Linter config
├── oapi-codegen.yaml             # Code generation config
└── Makefile                      # Updated with openapi targets
```

## Steps

1. **Validate arguments**:
   - API version must match pattern: `v[0-9]+` (e.g., v1, v2)

2. **Create directory structure**:
   ```bash
   mkdir -p api/{version}/paths
   mkdir -p api/{version}/components/{schemas,parameters,responses,requests}
   ```

3. **Create main openapi.yaml**:
   - Set title from project name (from go.mod)
   - Configure server URLs
   - Add $ref to component aggregators

4. **Create aggregator files** (`_index.yaml`):
   - Empty aggregators for paths, schemas, parameters, responses, requests
   - Each with comment explaining purpose

5. **Create .redocly.yaml**:
   - Use `recommended` ruleset
   - Enable operationId, summary, kebab-case rules

6. **Create oapi-codegen.yaml** in project root:
   - Package: `api`
   - Generate: `chi-server: true`, `models: true`
   - Output: `internal/http/{version}/api.gen.go`

7. **Update Makefile** (or create if not exists):
   - Add targets: `openapi-lint`, `openapi-bundle`, `openapi-generate`
   - Add `tools-openapi` target

8. **Report created files** and next steps.

## Example

```
/openapi-init v1
```

Creates:

**api/v1/openapi.yaml:**
```yaml
openapi: "3.0.3"
info:
  title: "myproject API"
  version: "1.0.0"

servers:
  - url: http://localhost:8080/api/v1
    description: Local development

paths:
  $ref: "paths/_index.yaml"

components:
  schemas:
    $ref: "components/schemas/_index.yaml"
  parameters:
    $ref: "components/parameters/_index.yaml"
  responses:
    $ref: "components/responses/_index.yaml"
  requestBodies:
    $ref: "components/requests/_index.yaml"
```

**api/v1/components/schemas/_index.yaml:**
```yaml
# Schema definitions
# Add schemas here:
#   User:
#     $ref: ./user.yaml#/User
```

**oapi-codegen.yaml:**
```yaml
package: api
generate:
  chi-server: true
  models: true
output: internal/http/v1/api.gen.go
```

## Final Validation (REQUIRED)

After generating files, run:

```bash
make openapi-lint
```

Fix any issues before reporting completion.

## Next Steps After Generation

1. Add your first resource: `/openapi-add-path users`
2. Define schemas in `api/v1/components/schemas/`
3. Run `make openapi-generate` to generate Go code
4. Implement handlers in `internal/http/v1/handler_impl.go`
