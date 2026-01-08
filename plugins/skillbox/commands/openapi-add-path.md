---
name: openapi-add-path
description: Add a new resource path with CRUD operations to OpenAPI spec
---

# /openapi-add-path

Add a new resource path with CRUD operations to an existing OpenAPI specification.

## Usage

```
/openapi-add-path <resource-name> [--operations <ops>] [--version <v1>]
```

## Arguments

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `resource-name` | Yes | - | Resource name in singular form (e.g., `user`, `order`) |
| `--operations` | No | `list,create,get,update,delete` | Comma-separated CRUD operations |
| `--version` | No | `v1` | API version directory |

## Prerequisites

- OpenAPI structure exists (run `/openapi-init` first)
- `api/{version}/openapi.yaml` exists

## Generated Files

```
api/{version}/
├── paths/
│   ├── _index.yaml              # Updated with new path refs
│   └── {resources}.yaml         # New path file (plural name)
└── components/
    └── schemas/
        ├── _index.yaml          # Updated with new schema refs
        └── {resource}.yaml      # New schema file (singular name)
```

## Steps

1. **Validate resource name**:
   - Must be lowercase, singular (e.g., `user`, `order`)
   - Convert to plural for paths (users, orders)

2. **Check prerequisites**:
   - Verify `api/{version}/openapi.yaml` exists
   - Verify `api/{version}/paths/_index.yaml` exists

3. **Create paths file** (`api/{version}/paths/{resources}.yaml`):
   - Collection endpoint: `/{resources}` (GET list, POST create)
   - Item endpoint: `/{resources}/{resourceID}` (GET, PUT, DELETE)
   - Include operationId, summary, description for each operation
   - Reference schemas from components

4. **Create schema file** (`api/{version}/components/schemas/{resource}.yaml`):
   - `{Resource}` - Main resource schema
   - `Create{Resource}Request` - Create request body
   - `Update{Resource}Request` - Update request body
   - `{Resource}List` - Paginated list response

5. **Update aggregators**:
   - Add path reference to `paths/_index.yaml`
   - Add schema references to `components/schemas/_index.yaml`

6. **Report created files** and next steps.

## Operations

| Operation | HTTP Method | Path | operationId |
|-----------|-------------|------|-------------|
| `list` | GET | /{resources} | list{Resources} |
| `create` | POST | /{resources} | create{Resource} |
| `get` | GET | /{resources}/{resourceID} | get{Resource} |
| `update` | PUT | /{resources}/{resourceID} | update{Resource} |
| `delete` | DELETE | /{resources}/{resourceID} | delete{Resource} |

## Example

```
/openapi-add-path order --operations list,create,get
```

Creates:

**api/v1/paths/orders.yaml:**
```yaml
/orders:
  get:
    operationId: listOrders
    summary: List orders
    tags:
      - Orders
    parameters:
      - name: limit
        in: query
        schema:
          type: integer
          default: 20
      - name: offset
        in: query
        schema:
          type: integer
          default: 0
    responses:
      "200":
        description: Success
        content:
          application/json:
            schema:
              $ref: "../components/schemas/_index.yaml#/OrderList"

  post:
    operationId: createOrder
    summary: Create an order
    tags:
      - Orders
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: "../components/schemas/_index.yaml#/CreateOrderRequest"
    responses:
      "201":
        description: Created
        content:
          application/json:
            schema:
              $ref: "../components/schemas/_index.yaml#/Order"

/orders/{orderID}:
  parameters:
    - name: orderID
      in: path
      required: true
      schema:
        type: string
        format: uuid
  get:
    operationId: getOrder
    summary: Get order by ID
    tags:
      - Orders
    responses:
      "200":
        description: Success
        content:
          application/json:
            schema:
              $ref: "../components/schemas/_index.yaml#/Order"
```

**api/v1/components/schemas/order.yaml:**
```yaml
Order:
  type: object
  required:
    - id
    - status
    - createdAt
  properties:
    id:
      type: string
      format: uuid
      x-go-type-name: OrderID
    status:
      type: string
      enum: [pending, processing, completed, cancelled]
    createdAt:
      type: string
      format: date-time

CreateOrderRequest:
  type: object
  required:
    - items
  properties:
    items:
      type: array
      items:
        type: object
        # Define order item schema

OrderList:
  type: object
  required:
    - items
    - totalCount
  properties:
    items:
      type: array
      items:
        $ref: "#/Order"
    totalCount:
      type: integer
      format: int64
```

**Updated api/v1/paths/_index.yaml:**
```yaml
/orders:
  $ref: ./orders.yaml#/~1orders
/orders/{orderID}:
  $ref: ./orders.yaml#/~1orders~1{orderID}
```

## Final Validation (REQUIRED)

After generating files, run:

```bash
make openapi-lint
```

Fix any issues before reporting completion.

## Next Steps After Generation

1. Customize schema properties in `api/v1/components/schemas/{resource}.yaml`
2. Add validation tags with `x-oapi-codegen-extra-tags`
3. Run `make openapi-generate` to regenerate Go code
4. Implement new handler methods
