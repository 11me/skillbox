# Build & Deploy Pattern

Version injection, multi-stage Dockerfile, Makefile.

## Version Injection

```go
// internal/app/version.go
package app

var ServiceVersion = "dev"
```

Build with version:
```bash
go build -ldflags "-X myapp/internal/app.ServiceVersion=v1.2.3" ./cmd/app
```

## Makefile

```makefile
APP_NAME := myapp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(APP_NAME)/internal/app.ServiceVersion=$(VERSION)"

.PHONY: build run test lint docker clean migrate

build: ## Build the application
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/app

run: ## Run locally
	go run $(LDFLAGS) ./cmd/app

test: ## Run tests
	go test -race -cover ./...

lint: ## Run linter
	golangci-lint run ./...

docker: ## Build Docker image
	docker build --build-arg SERVICE_VERSION=$(VERSION) -t $(APP_NAME):$(VERSION) .

clean: ## Clean build artifacts
	rm -rf bin/

migrate: ## Run migrations
	goose -dir migrations postgres "$(DB_DSN)" up

migrate-down: ## Rollback last migration
	goose -dir migrations postgres "$(DB_DSN)" down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
```

## Multi-Stage Dockerfile

```dockerfile
# Stage 1: Build
FROM golang:alpine AS builder

ARG SERVICE_VERSION=dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags "-X myapp/internal/app.ServiceVersion=$SERVICE_VERSION" \
    -o /app/bin/service ./cmd/app

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/bin/service .
COPY --from=builder /app/migrations ./migrations

RUN chmod +x ./service

EXPOSE 8080
CMD ["./service"]
```

## docker-compose.yml

```yaml
services:
  app:
    build:
      context: .
      args:
        SERVICE_VERSION: ${VERSION:-dev}
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=debug
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=myapp
      - DB_USER=myapp
      - DB_PASSWORD=myapp
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: myapp
      POSTGRES_PASSWORD: myapp
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myapp"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

## .env.example

```bash
# Application
APP_NAME=myapp
LOG_LEVEL=info

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=myapp
DB_PASSWORD=myapp
DB_SSL_MODE=disable
DB_MAX_CONNS=10
```

## CI/CD Notes

- Use `make docker` for building images
- Pass `VERSION` from CI (git tag, commit SHA)
- Use `make migrate` for database migrations
- Run `make lint test` before builds
