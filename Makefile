# Salon Platform - Root Makefile
# Manages both user-service and salon-service

.PHONY: help build test clean deploy docker-build docker-push local-up local-down

# Default target
help:
	@echo "Salon Platform - Available Commands:"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  make build           Build all services"
	@echo "  make build-user      Build user-service only"
	@echo "  make build-salon     Build salon-service only"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  make test            Run tests for all services"
	@echo "  make test-user       Run user-service tests"
	@echo "  make test-salon      Run salon-service tests"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  make docker-build    Build Docker images"
	@echo "  make docker-push     Push images to registry"
	@echo "  make local-up        Start local development environment"
	@echo "  make local-down      Stop local development environment"
	@echo ""
	@echo "🚀 Deployment Commands:"
	@echo "  make deploy          Deploy all services to dev (default)"
	@echo "  make deploy-dev      Deploy all services to dev"
	@echo "  make deploy-stage    Deploy all services to stage"
	@echo "  make deploy-prod     Deploy all services to prod"
	@echo "  make deploy-status   Show multi-environment status"
	@echo ""
	@echo "🧹 Utility Commands:"
	@echo "  make clean           Clean build artifacts"
	@echo "  make tidy            Run go mod tidy for all services"
	@echo "  make fmt             Format code for all services"

# Build commands
build: build-user build-salon

build-user:
	@echo "🏗️ Building user-service..."
	cd user-service && go build -o ../bin/user-service ./cmd

build-salon:
	@echo "🏗️ Building salon-service..."
	cd salon-service && go build -o ../bin/salon-service ./cmd

# Test commands
test: test-user test-salon

test-user:
	@echo "🧪 Testing user-service..."
	cd user-service && go test ./...

test-salon:
	@echo "🧪 Testing salon-service..."
	cd salon-service && go test ./...

test-shared:
	@echo "🧪 Testing salon-shared..."
	cd salon-shared && go test ./...

# Docker commands
docker-build:
	@echo "🐳 Building Docker images..."
	docker build -f Dockerfile.user-service -t salon/user-service:latest .
	docker build -f Dockerfile.salon-service -t salon/salon-service:latest .

docker-push:
	@echo "🐳 Pushing Docker images..."
	docker push salon/user-service:latest
	docker push salon/salon-service:latest

# Local development
local-up:
	@echo "🚀 Starting local development environment..."
	docker-compose -f docker-compose.yml up -d --build

local-down:
	@echo "🛑 Stopping local development environment..."
	docker-compose -f docker-compose.yml down

local-logs:
	@echo "📋 Showing local development logs..."
	docker-compose -f docker-compose.yml logs -f

# Deployment commands
deploy: deploy-dev

deploy-dev:
	@echo "🚀 Deploying all services to DEV..."
	./deploy-multi-env.sh all dev

deploy-stage:
	@echo "🚀 Deploying all services to STAGE..."
	./deploy-multi-env.sh all stage

deploy-prod:
	@echo "🚀 Deploying all services to PROD..."
	./deploy-multi-env.sh all prod --validate-only
	@echo "⚠️  Production deployment requires manual approval"
	@echo "Run: ./deploy-multi-env.sh all prod (after validation)"

deploy-status:
	@echo "📊 Multi-environment status..."
	./deploy-multi-env.sh status

# Environment-specific service deployments
deploy-user-dev:
	./deploy-multi-env.sh user-service dev

deploy-user-stage:
	./deploy-multi-env.sh user-service stage

deploy-user-prod:
	./deploy-multi-env.sh user-service prod --validate-only

deploy-salon-dev:
	./deploy-multi-env.sh salon-service dev

deploy-salon-stage:
	./deploy-multi-env.sh salon-service stage

deploy-salon-prod:
	./deploy-multi-env.sh salon-service prod --validate-only

# Utility commands
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	cd user-service && go clean
	cd salon-service && go clean

tidy:
	@echo "🧹 Running go mod tidy..."
	cd user-service && go mod tidy
	cd salon-service && go mod tidy
	cd salon-shared && go mod tidy

fmt:
	@echo "🎨 Formatting code..."
	cd user-service && go fmt ./...
	cd salon-service && go fmt ./...
	cd salon-shared && go fmt ./...

# Create bin directory
bin:
	mkdir -p bin

# Database migrations
migrate-up-user:
	@echo "📊 Running user-service migrations..."
	cd user-service && make migrate-up

migrate-up-salon:
	@echo "📊 Running salon-service migrations..."
	cd salon-service && migrate -path ./migrations -database "$(SALON_DB_URL)" up

migrate-down-user:
	@echo "📊 Rolling back user-service migrations..."
	cd user-service && make migrate-down

migrate-down-salon:
	@echo "📊 Rolling back salon-service migrations..."
	cd salon-service && migrate -path ./migrations -database "$(SALON_DB_URL)" down 1

# Development helpers
dev-user:
	@echo "🔧 Starting user-service in development mode..."
	cd user-service && go run ./cmd

dev-salon:
	@echo "🔧 Starting salon-service in development mode..."
	cd salon-service && go run ./cmd

# Security scan
security-scan:
	@echo "🔒 Running security scan..."
	docker run --rm -v $(PWD):/app securecodewarrior/docker-security-scanner /app

# Generate API documentation
docs:
	@echo "📚 Generating API documentation..."
	@echo "User Service API: http://localhost:8080/docs"
	@echo "Salon Service API: http://localhost:8081/docs"
