# Makefile for Go project
# Usage: make help

APP_NAME := {{APP_NAME}}
MODULE := {{MODULE}}
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X $(MODULE)/internal/app.ServiceVersion=$(VERSION)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Directories
BIN_DIR := bin
CMD_DIR := cmd/app

.PHONY: all build run test lint clean docker docker-push migrate help

all: lint test build ## Run lint, test, and build

build: ## Build the application
	@echo "Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)

run: ## Run the application locally
	$(GOCMD) run $(LDFLAGS) ./$(CMD_DIR)

test: ## Run tests with race detection
	$(GOTEST) -race -cover -coverprofile=coverage.out ./...

test-short: ## Run short tests only
	$(GOTEST) -short ./...

test-integration: ## Run integration tests
	$(GOTEST) -tags=integration -race ./...

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	$(GOCMD) fmt ./...
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

tidy: ## Tidy and verify dependencies
	$(GOMOD) tidy
	$(GOMOD) verify

clean: ## Clean build artifacts
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out

# Docker targets
docker: ## Build Docker image
	docker build \
		--build-arg SERVICE_VERSION=$(VERSION) \
		-t $(APP_NAME):$(VERSION) \
		-t $(APP_NAME):latest \
		.

docker-push: ## Push Docker image to registry
	docker push $(APP_NAME):$(VERSION)
	docker push $(APP_NAME):latest

# Database migrations
migrate-up: ## Run database migrations
	@which goose > /dev/null || go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations postgres "$(DB_DSN)" up

migrate-down: ## Rollback last migration
	goose -dir migrations postgres "$(DB_DSN)" down

migrate-status: ## Show migration status
	goose -dir migrations postgres "$(DB_DSN)" status

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	goose -dir migrations create $(NAME) sql

# Development helpers
dev: ## Run with hot reload (requires air)
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air

deps: ## Download dependencies
	$(GOMOD) download

tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/air-verse/air@latest
	go install golang.org/x/tools/cmd/goimports@latest

help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
