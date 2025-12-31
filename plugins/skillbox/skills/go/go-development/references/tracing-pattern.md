# Tracing Pattern

OpenTelemetry distributed tracing for Go services.

## Core Components

| Component | Purpose |
|-----------|---------|
| **TracerProvider** | Creates and manages tracers |
| **Exporter** | Sends traces to backend (Jaeger, OTLP) |
| **Propagator** | Passes trace context across services |
| **Span** | Single operation within a trace |

## TracerProvider Setup

```go
package tracing

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Config configures the tracer.
type Config struct {
    ServiceName    string
    ServiceVersion string
    OTLPEndpoint   string // e.g. "localhost:4317"
    Insecure       bool   // true for local dev
}

// InitTracer initializes OpenTelemetry tracer.
// Returns shutdown function to flush pending traces.
func InitTracer(ctx context.Context, cfg Config) (func(context.Context) error, error) {
    // Create OTLP exporter
    opts := []otlptracegrpc.Option{
        otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
    }
    if cfg.Insecure {
        opts = append(opts, otlptracegrpc.WithInsecure())
    }

    exporter, err := otlptracegrpc.New(ctx, opts...)
    if err != nil {
        return nil, err
    }

    // Create resource with service info
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.ServiceVersion),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Create TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.AlwaysSample()), // adjust for production
    )

    // Register as global provider
    otel.SetTracerProvider(tp)

    // Set up context propagation (W3C TraceContext)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp.Shutdown, nil
}
```

## Main Application Setup

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Initialize tracer
    shutdown, err := tracing.InitTracer(ctx, tracing.Config{
        ServiceName:    "my-service",
        ServiceVersion: "1.0.0",
        OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
        Insecure:       true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)

    // ... rest of application
}
```

## HTTP Middleware

Use `router.Use()` with tracing middleware:

```go
// pkg/tracing/handler.go
package tracing

type Middleware func(next http.Handler) http.Handler

type Option func(*options)

// Handler returns OpenTelemetry tracing middleware.
func Handler(opts ...Option) Middleware {
    cfg := defaultOptions()
    for _, opt := range opts {
        opt(cfg)
    }

    filter := func(r *http.Request) bool {
        _, ignored := cfg.ignorePaths[r.URL.Path]
        return !ignored
    }

    return func(next http.Handler) http.Handler {
        return otelhttp.NewHandler(next, cfg.serverName,
            otelhttp.WithFilter(filter),
        )
    }
}

// Options
func WithServerName(name string) Option { ... }
func WithIgnorePaths(paths ...string) Option { ... }
```

**Usage:**

```go
router := chi.NewRouter()

router.Use(
    tracing.Handler(),                          // default: ignores /health, /ready
    tracing.Handler(WithServerName("api")),     // custom server name
    tracing.Handler(WithIgnorePaths("/metrics")), // custom ignore paths
    middleware.RequestID,
    middleware.Recoverer,
)

router.Get("/users/{id}", handler.GetUser)
router.Post("/users", handler.CreateUser)
```

## Manual Span Creation

For service layer and custom operations:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("user-service")

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Start span
    ctx, span := tracer.Start(ctx, "CreateUser")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("user.email", req.Email),
    )

    // Do work...
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        // Record error
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    // Add result attributes
    span.SetAttributes(attribute.String("user.id", user.ID.String()))
    return user, nil
}
```

## Database Tracing (pgx)

Use `otelpgx` for automatic database tracing:

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/exaring/otelpgx"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    cfg, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }

    // Add OTEL tracer
    cfg.ConnConfig.Tracer = otelpgx.NewTracer()

    return pgxpool.NewWithConfig(ctx, cfg)
}
```

This automatically creates spans for:
- Query execution
- Transaction begin/commit/rollback
- Connection acquire/release

## Context Propagation

Always pass context through the call chain:

```go
// HTTP Handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Context contains trace info from HTTP middleware
    ctx := r.Context()

    user, err := h.services.Users().GetByID(ctx, id)
    // ...
}

// Service
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    ctx, span := tracer.Start(ctx, "GetByID")
    defer span.End()

    // Pass context to repository
    return s.repo.FindByID(ctx, id)
}

// Repository
func (r *userRepo) FindByID(ctx context.Context, id string) (*User, error) {
    // Context carries trace through to database
    row := r.db.QueryRow(ctx, query, id)
    // ...
}
```

## Sampling Strategies

For production, don't sample everything:

```go
// Always sample (dev/testing)
trace.WithSampler(trace.AlwaysSample())

// Never sample (disable tracing)
trace.WithSampler(trace.NeverSample())

// Ratio-based (sample 10% of traces)
trace.WithSampler(trace.TraceIDRatioBased(0.1))

// Parent-based with ratio fallback
trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.1)))
```

## Error Recording

```go
func (s *Service) DoWork(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "DoWork")
    defer span.End()

    if err := s.step1(ctx); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "step1 failed")
        return err
    }

    span.SetStatus(codes.Ok, "")
    return nil
}
```

## Adding Events

```go
func (s *Service) ProcessOrder(ctx context.Context, orderID string) error {
    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()

    // Add event when something notable happens
    span.AddEvent("order.validated", trace.WithAttributes(
        attribute.String("order.id", orderID),
    ))

    // ... process order

    span.AddEvent("order.completed")
    return nil
}
```

## Baggage (Cross-Service Context)

Pass values across service boundaries:

```go
import "go.opentelemetry.io/otel/baggage"

// Set baggage
m, _ := baggage.NewMember("user.id", userID)
b, _ := baggage.New(m)
ctx = baggage.ContextWithBaggage(ctx, b)

// Get baggage (in another service)
b := baggage.FromContext(ctx)
userID := b.Member("user.id").Value()
```

## Best Practices

### DO:
- ✅ Always pass context through the call chain
- ✅ Use meaningful span names
- ✅ Add relevant attributes (user ID, order ID, etc.)
- ✅ Record errors with `span.RecordError()`
- ✅ Use sampling in production
- ✅ Filter health check endpoints

### DON'T:
- ❌ Create spans for trivial operations
- ❌ Add sensitive data as attributes (passwords, tokens)
- ❌ Use `AlwaysSample()` in production
- ❌ Forget to call `span.End()`
- ❌ Block on span operations

## Related

- [database-pattern.md](database-pattern.md) — Database patterns
- [logging-pattern.md](logging-pattern.md) — Logging with slog
- [middleware-pattern.md](middleware-pattern.md) — HTTP middleware
