# OpenAPI Extensions for Go Code Generation

## Overview

oapi-codegen supports several vendor extensions (`x-*` properties) that control how Go code is generated from OpenAPI schemas.

## Available Extensions

### x-go-type-name

Override the generated Go type name.

**Use case**: Create meaningful type names instead of auto-generated ones.

```yaml
components:
  schemas:
    UserIdentifier:
      type: string
      format: uuid
      x-go-type-name: UserID
```

**Generated Go:**

```go
type UserID = string
```

**Without extension:**

```go
type UserIdentifier = string
```

### x-go-type

Map schema to an existing Go type.

**Use case**: Reuse existing types from your codebase or third-party libraries.

```yaml
components:
  schemas:
    Timestamp:
      type: string
      format: date-time
      x-go-type: time.Time

    Money:
      type: string
      x-go-type: github.com/shopspring/decimal.Decimal

    CustomID:
      type: string
      x-go-type: github.com/yourorg/types.ID
```

**Generated Go:**

```go
import (
    "time"
    "github.com/shopspring/decimal"
    "github.com/yourorg/types"
)

type Timestamp = time.Time
type Money = decimal.Decimal
type CustomID = types.ID
```

### x-go-type-import

Control import path and alias when using x-go-type.

**Use case**: Handle import conflicts or use specific aliases.

```yaml
components:
  schemas:
    UUID:
      type: string
      format: uuid
      x-go-type: uuid.UUID
      x-go-type-import:
        path: github.com/google/uuid

    # With alias
    ExternalUser:
      type: object
      x-go-type: user.User
      x-go-type-import:
        path: github.com/external/service/models
        name: externalmodels
```

**Generated Go:**

```go
import (
    "github.com/google/uuid"
    externalmodels "github.com/external/service/models"
)

type UUID = uuid.UUID
type ExternalUser = externalmodels.User
```

### x-oapi-codegen-extra-tags

Add custom struct tags to generated fields.

**Use case**: Integrate with validation libraries, ORMs, or custom serialization.

```yaml
components:
  schemas:
    User:
      type: object
      required:
        - email
        - name
      properties:
        email:
          type: string
          format: email
          x-oapi-codegen-extra-tags:
            validate: "required,email"
            db: "email"
        name:
          type: string
          minLength: 2
          maxLength: 100
          x-oapi-codegen-extra-tags:
            validate: "required,min=2,max=100"
            db: "name"
        status:
          type: string
          enum: [active, inactive]
          x-oapi-codegen-extra-tags:
            db: "status"
            gorm: "type:varchar(20)"
```

**Generated Go:**

```go
type User struct {
    Email  string `json:"email" validate:"required,email" db:"email"`
    Name   string `json:"name" validate:"required,min=2,max=100" db:"name"`
    Status string `json:"status,omitempty" db:"status" gorm:"type:varchar(20)"`
}
```

**Common tag libraries:**

| Tag | Library | Example |
|-----|---------|---------|
| `validate` | go-playground/validator | `validate:"required,email"` |
| `db` | jmoiron/sqlx | `db:"column_name"` |
| `gorm` | gorm.io/gorm | `gorm:"column:name"` |
| `bson` | go.mongodb.org/mongo-driver | `bson:"field_name"` |
| `mapstructure` | mitchellh/mapstructure | `mapstructure:"key"` |

### x-go-json-ignore

Exclude field from JSON serialization.

**Use case**: Hide sensitive or internal fields.

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        password:
          type: string
          x-go-json-ignore: true
        internalScore:
          type: integer
          x-go-json-ignore: true
```

**Generated Go:**

```go
type User struct {
    ID            string `json:"id"`
    Password      string `json:"-"`
    InternalScore int    `json:"-"`
}
```

### x-enum-varnames

Customize enum constant names.

**Use case**: Use Go-idiomatic constant names.

```yaml
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive
        - pending_review
      x-enum-varnames:
        - StatusActive
        - StatusInactive
        - StatusPendingReview
```

**Generated Go:**

```go
type Status string

const (
    StatusActive        Status = "active"
    StatusInactive      Status = "inactive"
    StatusPendingReview Status = "pending_review"
)
```

### x-nullable

Mark field as nullable (generates pointer type).

```yaml
components:
  schemas:
    User:
      properties:
        middleName:
          type: string
          x-nullable: true
```

**Generated Go:**

```go
type User struct {
    MiddleName *string `json:"middleName,omitempty"`
}
```

## Combining Extensions

Extensions can be combined for maximum control:

```yaml
components:
  schemas:
    Order:
      type: object
      properties:
        id:
          type: string
          format: uuid
          x-go-type-name: OrderID
          x-oapi-codegen-extra-tags:
            db: "id"
            gorm: "primaryKey"

        amount:
          type: string
          x-go-type: decimal.Decimal
          x-go-type-import:
            path: github.com/shopspring/decimal
          x-oapi-codegen-extra-tags:
            db: "amount"
            gorm: "type:decimal(10,2)"

        status:
          type: string
          enum: [pending, completed, cancelled]
          x-enum-varnames:
            - OrderStatusPending
            - OrderStatusCompleted
            - OrderStatusCancelled
          x-oapi-codegen-extra-tags:
            validate: "required,oneof=pending completed cancelled"
            db: "status"

        internalNotes:
          type: string
          x-go-json-ignore: true
          x-oapi-codegen-extra-tags:
            db: "internal_notes"
```

## Best Practices

### 1. Use x-go-type for Common Types

```yaml
# Good: Use standard time.Time
Timestamp:
  type: string
  format: date-time
  x-go-type: time.Time

# Good: Use decimal for money
Price:
  type: string
  x-go-type: github.com/shopspring/decimal.Decimal
```

### 2. Add Validation Tags

```yaml
# Good: Validate at struct level
Email:
  type: string
  format: email
  x-oapi-codegen-extra-tags:
    validate: "required,email"
```

### 3. Use x-go-type-name for IDs

```yaml
# Good: Typed IDs prevent mixing up IDs
UserID:
  type: string
  format: uuid
  x-go-type-name: UserID

OrderID:
  type: string
  format: uuid
  x-go-type-name: OrderID
```

### 4. Hide Sensitive Fields

```yaml
# Good: Never serialize passwords
password:
  type: string
  x-go-json-ignore: true
```

### 5. Document Extensions

Add comments explaining why extensions are used:

```yaml
User:
  properties:
    # x-go-type-name creates typed ID to prevent mixing with other IDs
    id:
      type: string
      x-go-type-name: UserID
```
