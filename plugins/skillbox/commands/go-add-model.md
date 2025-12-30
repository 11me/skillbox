---
description: Generate a new model with mapper for Go projects
---

# /go-add-model

Generate a new domain model with optional database mapper.

## Usage

```
/go-add-model <model-name> [--fields "field1:type,field2:type"] [--no-mapper]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `model-name` | Yes | Model name in PascalCase (e.g., `User`, `Order`) |
| `--fields` | No | Comma-separated fields with types |
| `--no-mapper` | No | Skip generating mapper file |

## Generated Files

```
internal/models/
├── <name>.go         # Domain model
└── <name>_mapper.go  # Database mapper (optional)
```

## Steps

1. **Validate model name**:
   - Must be PascalCase: `[A-Z][a-zA-Z0-9]*`
   - Must not already exist in `internal/models/`

2. **Parse fields** if provided:
   - Format: `fieldName:type` (e.g., `name:string,age:int,active:bool`)
   - Supported types: `string`, `int`, `int64`, `float64`, `bool`, `time`, `uuid`

3. **Generate model file** (`internal/models/<name>.go`):

```go
package models

import (
    "time"

    "github.com/google/uuid"
)

type {{Name}} struct {
    ID        uuid.UUID  `json:"id"`
    {{#fields}}
    {{FieldName}} {{FieldType}} `json:"{{json_name}}"`
    {{/fields}}
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

// New{{Name}} creates a new {{Name}} with generated ID and timestamps.
func New{{Name}}({{constructor_params}}) *{{Name}} {
    now := time.Now().UTC()
    return &{{Name}}{
        ID:        uuid.New(),
        {{#fields}}
        {{FieldName}}: {{paramName}},
        {{/fields}}
        CreatedAt: now,
        UpdatedAt: now,
    }
}
```

4. **Generate mapper file** (`internal/models/<name>_mapper.go`) unless `--no-mapper`:

```go
package models

import (
    "time"

    "github.com/google/uuid"
)

// {{Name}}Row represents database row for {{Name}}.
type {{Name}}Row struct {
    ID        uuid.UUID
    {{#fields}}
    {{FieldName}} {{DBFieldType}}
    {{/fields}}
    CreatedAt time.Time
    UpdatedAt time.Time
}

// To{{Name}} converts database row to domain model.
func (r *{{Name}}Row) To{{Name}}() *{{Name}} {
    return &{{Name}}{
        ID:        r.ID,
        {{#fields}}
        {{FieldName}}: r.{{FieldName}},
        {{/fields}}
        CreatedAt: r.CreatedAt,
        UpdatedAt: r.UpdatedAt,
    }
}

// {{Name}}ToRow converts domain model to database row.
func {{Name}}ToRow(m *{{Name}}) *{{Name}}Row {
    return &{{Name}}Row{
        ID:        m.ID,
        {{#fields}}
        {{FieldName}}: m.{{FieldName}},
        {{/fields}}
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }
}
```

5. **Report created files** and suggest next steps.

## Examples

### Basic model with fields

```
/go-add-model User --fields "name:string,email:string,active:bool"
```

Creates `internal/models/user.go`:
```go
package models

type User struct {
    ID        uuid.UUID  `json:"id"`
    Name      string     `json:"name"`
    Email     string     `json:"email"`
    Active    bool       `json:"active"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

func NewUser(name, email string, active bool) *User {
    now := time.Now().UTC()
    return &User{
        ID:        uuid.New(),
        Name:      name,
        Email:     email,
        Active:    active,
        CreatedAt: now,
        UpdatedAt: now,
    }
}
```

### Model without mapper

```
/go-add-model CreateOrderRequest --no-mapper
```

Creates only the model file without database mapper.

### Complex model

```
/go-add-model Order --fields "userID:uuid,total:float64,status:string"
```

## Field Type Mapping

| Type | Go Type | JSON Example |
|------|---------|--------------|
| `string` | `string` | `"value"` |
| `int` | `int` | `123` |
| `int64` | `int64` | `123456789` |
| `float64` | `float64` | `123.45` |
| `bool` | `bool` | `true` |
| `time` | `time.Time` | `"2024-01-01T00:00:00Z"` |
| `uuid` | `uuid.UUID` | `"550e8400-..."` |

## Next Steps After Generation

1. Add business methods to the model if needed
2. Create repository with `/go-add-repository`
3. Create database migration
4. Add validation methods
