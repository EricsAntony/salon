# Salon Platform - Root Makefile
# Manages user-service, salon-service, booking-service, payment-service, and notification-service

.PHONY: help build test clean deploy docker-build docker-push local-up local-down

# Default target
help:
	@echo "Salon Platform - Available Commands:"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  make build           Build all services"
	@echo "  make build-user      Build user-service only"
	@echo "  make build-salon     Build salon-service only"
	@echo "  make build-booking   Build booking-service only"
	@echo "  make build-payment   Build payment-service only"
	@echo "  make build-notification Build notification-service only"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  make test            Run tests for all services"
	@echo "  make test-user       Run user-service tests"
	@echo "  make test-salon      Run salon-service tests"
	@echo "  make test-booking    Run booking-service tests"
	@echo "  make test-payment    Run payment-service tests"
	@echo "  make test-notification Run notification-service tests"
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
	@echo ""
	@echo "⚙️  Configuration Commands:"
	@echo "  make config-list     List all environment configurations"
	@echo "  make config-validate ENV=dev  Validate environment configs"
	@echo "  make config-diff ENV1=dev ENV2=prod  Compare configs"

# Build commands
build: build-user build-salon build-booking build-payment build-notification

build-user:
	@echo "🏗️ Building user-service..."
	cd user-service && go build -o ../bin/user-service ./cmd

build-salon:
	@echo "🏗️ Building salon-service..."
	cd salon-service && go build -o ../bin/salon-service ./cmd

build-booking:
	@echo "🏗️ Building booking-service..."
	cd booking-service && go build -o ../bin/booking-service ./cmd

build-payment:
	@echo "🏗️ Building payment-service..."
	cd payment-service && go build -o ../bin/payment-service ./cmd

build-notification:
	@echo "🏗️ Building notification-service..."
	cd notification-service && go build -o ../bin/notification-service ./cmd

# Test commands
test: test-user test-salon test-booking test-payment test-notification

test-user:
	@echo "🧪 Testing user-service..."
	cd user-service && go test ./...

test-salon:
	@echo "🧪 Testing salon-service..."
	cd salon-service && go test ./...

test-booking:
	@echo "🧪 Testing booking-service..."
	cd booking-service && go test ./...

test-payment:
	@echo "🧪 Testing payment-service..."
	cd payment-service && go test ./...

test-notification:
	@echo "🧪 Testing notification-service..."
	cd notification-service && go test ./...

test-shared:
	@echo "🧪 Testing salon-shared..."
	cd salon-shared && go test ./...

# Docker commands
docker-build:
	@echo "🐳 Building Docker images..."
	docker build -f Dockerfile.user-service -t salon/user-service:latest .
	docker build -f Dockerfile.salon-service -t salon/salon-service:latest .
	docker build -f Dockerfile.booking-service -t salon/booking-service:latest .
	docker build -f Dockerfile.payment-service -t salon/payment-service:latest .
	docker build -f Dockerfile.notification-service -t salon/notification-service:latest .

docker-push:
	@echo "🐳 Pushing Docker images..."
	docker push salon/user-service:latest
	docker push salon/salon-service:latest
	docker push salon/booking-service:latest
	docker push salon/payment-service:latest
	docker push salon/notification-service:latest

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

deploy-booking-dev:
	./deploy-multi-env.sh booking-service dev

deploy-booking-stage:
	./deploy-multi-env.sh booking-service stage

deploy-booking-prod:
	./deploy-multi-env.sh booking-service prod --validate-only

deploy-payment-dev:
	./deploy-multi-env.sh payment-service dev

deploy-payment-stage:
	./deploy-multi-env.sh payment-service stage

deploy-payment-prod:
	./deploy-multi-env.sh payment-service prod --validate-only

deploy-notification-dev:
	./deploy-multi-env.sh notification-service dev

deploy-notification-stage:
	./deploy-multi-env.sh notification-service stage

deploy-notification-prod:
	./deploy-multi-env.sh notification-service prod --validate-only

# Utility commands
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	cd user-service && go clean
	cd salon-service && go clean
	cd booking-service && go clean
	cd payment-service && go clean
	cd notification-service && go clean

tidy:
	@echo "🧹 Running go mod tidy..."
	cd user-service && go mod tidy
	cd salon-service && go mod tidy
	cd booking-service && go mod tidy
	cd payment-service && go mod tidy
	cd notification-service && go mod tidy
	cd salon-shared && go mod tidy

fmt:
	@echo "🎨 Formatting code..."
	cd user-service && go fmt ./...
	cd salon-service && go fmt ./...
	cd booking-service && go fmt ./...
	cd payment-service && go fmt ./...
	cd notification-service && go fmt ./...
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

migrate-up-booking:
	@echo "📊 Running booking-service migrations..."
	cd booking-service && migrate -path ./migrations -database "$(BOOKING_DB_URL)" up

migrate-up-payment:
	@echo "📊 Running payment-service migrations..."
	cd payment-service && make migrate-up

migrate-up-notification:
	@echo "📊 Running notification-service migrations..."
	cd notification-service && make migrate-up

migrate-down-user:
	@echo "📊 Rolling back user-service migrations..."
	cd user-service && make migrate-down

migrate-down-salon:
	@echo "📊 Rolling back salon-service migrations..."
	cd salon-service && migrate -path ./migrations -database "$(SALON_DB_URL)" down 1

migrate-down-booking:
	@echo "📊 Rolling back booking-service migrations..."
	cd booking-service && migrate -path ./migrations -database "$(BOOKING_DB_URL)" down 1

migrate-down-payment:
	@echo "📊 Rolling back payment-service migrations..."
	cd payment-service && make migrate-down

migrate-down-notification:
	@echo "📊 Rolling back notification-service migrations..."
	cd notification-service && make migrate-down

# Development helpers
dev-user:
	@echo "🔧 Starting user-service in development mode..."
	cd user-service && go run ./cmd

dev-salon:
	@echo "🔧 Starting salon-service in development mode..."
	cd salon-service && go run ./cmd

dev-booking:
	@echo "🔧 Starting booking-service in development mode..."
	cd booking-service && go run ./cmd

dev-payment:
	@echo "🔧 Starting payment-service in development mode..."
	cd payment-service && go run ./cmd

dev-notification:
	@echo "🔧 Starting notification-service in development mode..."
	cd notification-service && go run ./cmd

# Security scan
security-scan:
	@echo "🔒 Running security scan..."
	docker run --rm -v $(PWD):/app securecodewarrior/docker-security-scanner /app

# Generate API documentation
docs:
	@echo "📚 Generating API documentation..."
	@echo "User Service API: http://localhost:8080/docs"
	@echo "Salon Service API: http://localhost:8081/docs"
	@echo "Booking Service API: http://localhost:8082/docs"

# Configuration management
config-list:
	@echo "📋 Listing all environment configurations..."
	./config-manager.sh list

config-validate:
	@echo "🔍 Validating $(ENV) environment configurations..."
	./config-manager.sh validate $(ENV)

config-diff:
	@echo "🔍 Comparing $(ENV1) vs $(ENV2) configurations..."
	./config-manager.sh diff $(ENV1) $(ENV2)

config-show:
	@echo "📄 Showing $(ENV) $(SERVICE) configuration..."
	./config-manager.sh show $(ENV) $(SERVICE)
