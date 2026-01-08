# Multi-stage Dockerfile for Go applications
# Always use latest stable versions from https://go.dev/dl/

# Stage 1: Build
FROM golang:alpine AS builder

ARG SERVICE_VERSION=dev
WORKDIR /app

# Install build dependencies
RUN apk --no-cache add ca-certificates tzdata git

# Download dependencies first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X {{MODULE}}/internal/app.ServiceVersion=$SERVICE_VERSION" \
    -o /app/bin/service ./cmd/app

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and migrations
COPY --from=builder /app/bin/service .
COPY --from=builder /app/migrations ./migrations

# Create non-root user
RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./service"]
