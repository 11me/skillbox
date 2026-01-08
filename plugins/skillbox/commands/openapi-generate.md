---
name: openapi-generate
description: Generate Go code from OpenAPI specification using oapi-codegen
---

# /openapi-generate

Generate Go server/client code from OpenAPI specification using oapi-codegen.

## Usage

```
/openapi-generate [--config <path>] [--version <v1>]
```

## Arguments

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `--config` | No | `oapi-codegen.yaml` | Path to oapi-codegen config |
| `--version` | No | `v1` | API version to generate |

## Prerequisites

- OpenAPI spec exists (`api/{version}/openapi.yaml`)
- `oapi-codegen.yaml` config exists
- Redocly CLI installed (for linting/bundling)
- oapi-codegen installed

## Generated Files

Depends on config, typically:

```
internal/http/{version}/
└── api.gen.go          # Generated types + server interface
```

## Steps

1. **Verify prerequisites**:
   - Check `api/{version}/openapi.yaml` exists
   - Check `oapi-codegen.yaml` exists
   - Check tools are installed

2. **Install tools if missing**:
   ```bash
   # Redocly CLI
   npm install -g @redocly/cli@latest

   # oapi-codegen
   go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
   ```

3. **Lint specification**:
   ```bash
   redocly lint api/{version}/openapi.yaml
   ```
   - If errors, report and stop
   - If warnings, report but continue

4. **Bundle specification** (resolve $refs):
   ```bash
   redocly bundle api/{version}/openapi.yaml -o api/{version}/bundle.yaml
   ```

5. **Generate Go code**:
   ```bash
   oapi-codegen --config oapi-codegen.yaml api/{version}/bundle.yaml
   ```

6. **Validate generated code**:
   ```bash
   golangci-lint run --fast internal/http/{version}/
   ```

7. **Report results**:
   - Generated file path
   - Number of types generated
   - Number of operations generated

## Example

```
/openapi-generate
```

Output:
```
Linting OpenAPI spec...
✓ api/v1/openapi.yaml is valid

Bundling spec...
✓ Created api/v1/bundle.yaml

Generating Go code...
✓ Generated internal/http/v1/api.gen.go
  - 12 types
  - 8 operations (Chi server)

Validating generated code...
✓ golangci-lint passed

Generation complete!

Next steps:
1. Implement ServerInterface in internal/http/v1/handler_impl.go
2. Wire router in cmd/server/main.go:
   handler := v1.NewHandler(services)
   router := v1.HandlerFromMux(handler, chi.NewRouter())
```

## Configuration Options

### Default config (oapi-codegen.yaml):

```yaml
package: api
generate:
  chi-server: true
  models: true
output: internal/http/v1/api.gen.go
```

### Generate client instead:

```yaml
generate:
  client: true
  models: true
output: pkg/client/client.gen.go
```

### Generate types only:

```yaml
generate:
  models: true
output: internal/models/types.gen.go
```

## Common Issues

### "redocly: command not found"

```bash
npm install -g @redocly/cli@latest
```

### "oapi-codegen: command not found"

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

### "$ref resolution failed"

Check that all `$ref` paths are correct and files exist.

### "operationId must be unique"

Each operation needs a unique operationId across the entire spec.

## Final Validation (REQUIRED)

After generation, run:

```bash
go build ./...
golangci-lint run ./...
```

Fix any issues before reporting completion.

## Next Steps After Generation

1. Create handler implementation file:
   ```go
   // internal/http/v1/handler_impl.go
   type Handler struct {
       services *services.Registry
   }

   func NewHandler(svc *services.Registry) *Handler {
       return &Handler{services: svc}
   }

   // Implement ServerInterface methods...
   ```

2. Wire router in main:
   ```go
   handler := v1.NewHandler(services)
   router := v1.HandlerFromMux(handler, chi.NewRouter())
   ```

3. Add middleware (logging, auth, etc.)

4. Write tests for handlers
