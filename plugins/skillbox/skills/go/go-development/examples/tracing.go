// Package tracing provides OpenTelemetry tracing setup and utilities.
//
// This example shows:
// - TracerProvider setup with OTLP exporter
// - HTTP middleware with chi router
// - Manual span creation
// - Database tracing with otelpgx
package tracing

import (
	"context"
	"net/http"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// ---------- Middleware Type ----------

// Middleware is the standard HTTP middleware signature.
type Middleware func(next http.Handler) http.Handler

// ---------- Configuration ----------

// Config configures the OpenTelemetry tracer.
type Config struct {
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string  // e.g. "localhost:4317"
	Insecure       bool    // true for local dev
	SampleRate     float64 // 0.0 to 1.0, default 1.0 (100%)
}

// ---------- TracerProvider Setup ----------

// InitTracer initializes OpenTelemetry tracer.
// Returns a shutdown function to flush pending traces.
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

	// Determine sampler
	var sampler trace.Sampler
	if cfg.SampleRate <= 0 {
		sampler = trace.NeverSample()
	} else if cfg.SampleRate >= 1.0 {
		sampler = trace.AlwaysSample()
	} else {
		sampler = trace.ParentBased(trace.TraceIDRatioBased(cfg.SampleRate))
	}

	// Create TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(sampler),
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

// ---------- HTTP Middleware ----------

// Option configures the tracing middleware.
type Option func(*options)

type options struct {
	serverName   string
	ignorePaths  map[string]struct{}
	filterFunc   func(*http.Request) bool
}

func defaultOptions() *options {
	return &options{
		serverName: "http-server",
		ignorePaths: map[string]struct{}{
			"/check/healthz/": {},
			"/check/readyz/":  {},
			"/metrics":        {},
		},
	}
}

// WithServerName sets the server name for spans.
func WithServerName(name string) Option {
	return func(o *options) {
		o.serverName = name
	}
}

// WithIgnorePaths sets paths to exclude from tracing.
func WithIgnorePaths(paths ...string) Option {
	return func(o *options) {
		o.ignorePaths = make(map[string]struct{}, len(paths))
		for _, p := range paths {
			o.ignorePaths[p] = struct{}{}
		}
	}
}

// WithFilter sets a custom filter function.
// Return true to trace the request, false to skip.
func WithFilter(fn func(*http.Request) bool) Option {
	return func(o *options) {
		o.filterFunc = fn
	}
}

// Handler returns OpenTelemetry tracing middleware.
// Usage: router.Use(tracing.Handler())
func Handler(opts ...Option) Middleware {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	// Build filter function
	filter := func(r *http.Request) bool {
		// Check custom filter first
		if cfg.filterFunc != nil {
			return cfg.filterFunc(r)
		}
		// Check ignore paths
		_, ignored := cfg.ignorePaths[r.URL.Path]
		return !ignored
	}

	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, cfg.serverName,
			otelhttp.WithFilter(filter),
		)
	}
}

// ---------- Database Setup ----------

// NewTracedPool creates a pgxpool with OpenTelemetry tracing.
func NewTracedPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// Add OTEL tracer for automatic query tracing
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	return pgxpool.NewWithConfig(ctx, cfg)
}

// ---------- Span Helpers ----------

// StartSpan starts a new span and returns the context and span.
// Always defer span.End() after calling this.
func StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, interface{ End(...trace.SpanEndOption) }) {
	return otel.Tracer(tracerName).Start(ctx, spanName)
}

// RecordError records an error on the current span from context.
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetAttributes sets attributes on the current span from context.
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// ---------- Usage Examples ----------

// Example usage in main.go:
//
//	func main() {
//	    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
//	    defer stop()
//
//	    // Initialize tracer
//	    shutdown, err := tracing.InitTracer(ctx, tracing.Config{
//	        ServiceName:    "my-service",
//	        ServiceVersion: "1.0.0",
//	        OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
//	        Insecure:       true,
//	        SampleRate:     0.1, // 10% sampling in production
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer shutdown(ctx)
//
//	    // Create traced database pool
//	    pool, err := tracing.NewTracedPool(ctx, os.Getenv("DATABASE_URL"))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer pool.Close()
//
//	    // Setup router with tracing middleware
//	    router := chi.NewRouter()
//	    router.Use(
//	        tracing.Handler(),                    // OpenTelemetry tracing
//	        tracing.Handler(WithServerName("api")), // or with custom name
//	        middleware.RequestID,
//	        middleware.Recoverer,
//	    )
//	    router.Get("/users/{id}", handler.GetUser)
//
//	    // Start server
//	    srv := &http.Server{Addr: ":8080", Handler: router}
//	    srv.ListenAndServe()
//	}

// Example usage in service:
//
//	var tracer = otel.Tracer("user-service")
//
//	func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
//	    ctx, span := tracer.Start(ctx, "CreateUser")
//	    defer span.End()
//
//	    span.SetAttributes(
//	        attribute.String("user.email", req.Email),
//	    )
//
//	    user, err := s.repo.Create(ctx, req)
//	    if err != nil {
//	        span.RecordError(err)
//	        span.SetStatus(codes.Error, err.Error())
//	        return nil, err
//	    }
//
//	    span.SetAttributes(attribute.String("user.id", user.ID.String()))
//	    return user, nil
//	}
