# Notification Service Makefile

.PHONY: build test clean run docker-build docker-run dev lint fmt vet deps migrate-up migrate-down help

# Variables
APP_NAME := notification-service
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_CONTAINER := $(APP_NAME)-container
GO_VERSION := 1.21
PORT := 8084

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) ./cmd

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || true

# Run the application locally
run: build
	@echo "Running $(APP_NAME) on port $(PORT)..."
	@./bin/$(APP_NAME)

# Run in development mode with hot reload
dev:
	@echo "Running $(APP_NAME) in development mode..."
	@go run ./cmd

# Docker operations are handled by root Makefile
docker-info:
	@echo "ℹ️  Docker operations are centralized in root Makefile:"
	@echo "  cd .. && make docker-build           # Build all service images"
	@echo "  cd .. && make local-up               # Start local development"
	@echo "  cd .. && make build-notification     # Build notification service only"

# Lint the code
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet the code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	@migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/salon?sslmode=disable" up

migrate-down:
	@echo "Running database migrations down..."
	@migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/salon?sslmode=disable" down

# Create new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations $$name

# Database setup for development
db-setup:
	@echo "Setting up development database..."
	@docker run -d \
		--name notification-postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=salon \
		-p 5432:5432 \
		postgres:15-alpine

# Stop development database
db-stop:
	@echo "Stopping development database..."
	@docker stop notification-postgres 2>/dev/null || true
	@docker rm notification-postgres 2>/dev/null || true

# Setup message queue for development
queue-setup:
	@echo "Setting up development message queue..."
	@docker run -d \
		--name notification-redis \
		-p 6379:6379 \
		redis:7-alpine

# Stop message queue
queue-stop:
	@echo "Stopping development message queue..."
	@docker stop notification-redis 2>/dev/null || true
	@docker rm notification-redis 2>/dev/null || true

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@swag init -g ./cmd/main.go -o ./docs

# Security scan
security:
	@echo "Running security scan..."
	@gosec ./...

# Performance benchmark
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Check for outdated dependencies
deps-check:
	@echo "Checking for outdated dependencies..."
	@go list -u -m all

# Full CI pipeline
ci: deps fmt vet lint test security

# Production build
prod-build:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/$(APP_NAME) ./cmd

# Deployment operations are handled by root Makefile
deploy-info:
	@echo "ℹ️  Deployment operations are centralized in root Makefile:"
	@echo "  cd .. && make deploy-dev         # Deploy to development"
	@echo "  cd .. && make deploy-stage       # Deploy to staging"
	@echo "  cd .. && make deploy-prod        # Deploy to production"

# Start all development services
dev-up: db-setup queue-setup
	@echo "Starting all development services..."
	@sleep 5
	@make dev

# Stop all development services
dev-down: db-stop queue-stop
	@echo "Stopping all development services..."

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the application locally"
	@echo "  dev           - Run in development mode"
	@echo "  docker-info   - Show Docker command info"
	@echo "  deploy-info   - Show deployment command info"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  deps          - Download dependencies"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  migrate-create- Create new migration"
	@echo "  db-setup      - Setup development database"
	@echo "  db-stop       - Stop development database"
	@echo "  queue-setup   - Setup development message queue"
	@echo "  queue-stop    - Stop development message queue"
	@echo "  install-tools - Install development tools"
	@echo "  docs          - Generate API documentation"
	@echo "  security      - Run security scan"
	@echo "  bench         - Run benchmarks"
	@echo "  deps-check    - Check for outdated dependencies"
	@echo "  ci            - Run full CI pipeline"
	@echo "  prod-build    - Build for production"
	@echo "  dev-up        - Start all development services"
	@echo "  dev-down      - Stop all development services"
	@echo "  help          - Show this help message"
