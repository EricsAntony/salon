# Booking Service Makefile

# Variables
SERVICE_NAME := booking-service
DOCKER_IMAGE := $(SERVICE_NAME)
DOCKER_TAG := latest
DB_URL := postgres://postgres:password@localhost:5432/salon?sslmode=disable

# Go commands
.PHONY: build test clean run dev docker-build docker-run

# Build the application
build:
	@echo "Building $(SERVICE_NAME)..."
	@go build -o $(SERVICE_NAME) ./cmd

# Run tests
test:
	@echo "Running tests for $(SERVICE_NAME)..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(SERVICE_NAME)
	@go clean

# Run the service locally
run: build
	@echo "Starting $(SERVICE_NAME)..."
	@./$(SERVICE_NAME)

# Run in development mode with live reload (requires air)
dev:
	@echo "Starting $(SERVICE_NAME) in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi

# Database commands
.PHONY: db-migrate db-reset db-seed

# Run database migrations
db-migrate:
	@echo "Running database migrations..."
	@BOOKING_SERVICE_DB_URL="$(DB_URL)" go run cmd/main.go --migrate-only

# Reset database (drop and recreate)
db-reset:
	@echo "Resetting database..."
	@psql "$(DB_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) db-migrate

# Seed database with test data
db-seed:
	@echo "Seeding database with test data..."
	@# Add seed data commands here

# Docker commands
.PHONY: docker-build docker-run docker-stop docker-logs

# Build Docker image
docker-build:
	@echo "Building Docker image for $(SERVICE_NAME)..."
	@cd .. && docker build -f Dockerfile.$(SERVICE_NAME) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container
docker-run:
	@echo "Running $(SERVICE_NAME) in Docker..."
	@docker run -d \
		--name $(SERVICE_NAME) \
		-p 8082:8082 \
		-e BOOKING_SERVICE_ENV=development \
		-e BOOKING_SERVICE_DB_URL="$(DB_URL)" \
		-e BOOKING_SERVICE_JWT_ACCESSSECRET="dev-access-secret" \
		-e BOOKING_SERVICE_JWT_REFRESHSECRET="dev-refresh-secret" \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Stop Docker container
docker-stop:
	@echo "Stopping $(SERVICE_NAME) Docker container..."
	@docker stop $(SERVICE_NAME) || true
	@docker rm $(SERVICE_NAME) || true

# View Docker logs
docker-logs:
	@docker logs -f $(SERVICE_NAME)

# Development commands
.PHONY: deps fmt lint vet

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Quality checks
quality: fmt vet lint test

# API testing commands
.PHONY: test-api test-health

# Test health endpoint
test-health:
	@echo "Testing health endpoint..."
	@curl -f http://localhost:8082/health || echo "Health check failed"

# Test API endpoints (requires running service)
test-api:
	@echo "Testing API endpoints..."
	@echo "Testing health..."
	@curl -f http://localhost:8082/health
	@echo "\nTesting ready..."
	@curl -f http://localhost:8082/ready

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the service locally"
	@echo "  dev           - Run in development mode with live reload"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-reset      - Reset database schema"
	@echo "  db-seed       - Seed database with test data"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"
	@echo "  docker-stop   - Stop Docker container"
	@echo "  docker-logs   - View Docker logs"
	@echo ""
	@echo "Development:"
	@echo "  deps          - Download dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  vet           - Vet code"
	@echo "  quality       - Run all quality checks"
	@echo ""
	@echo "Testing:"
	@echo "  test-health   - Test health endpoint"
	@echo "  test-api      - Test API endpoints"

# Default target
.DEFAULT_GOAL := help
