#  Salon Platform

A comprehensive salon management platform built with microservices architecture, featuring shared components and production-ready deployment.

##  Architecture

### Services
- **user-service** (Port 8080): Customer-facing API for user registration, authentication, and profile management
- **salon-service** (Port 8081): Business management API for salon operations, staff, and services  
- **salon-shared**: Common functionality shared across all services

### Shared Components
All services leverage the `salon-shared` module for:
-  **Authentication & JWT management** with user type validation
-  **Middleware** (audit logging, rate limiting, authorization)
-  **Validation utilities** (phone, OTP, common fields)
-  **Error handling** with consistent HTTP responses
-  **HTTP utilities** and crypto functions

##  Quick Start

### Local Development
```bash
# Start all services with Docker Compose
make local-up

# Services available at:
# User Service:  http://localhost:8080
# Salon Service: http://localhost:8081
# PostgreSQL:    localhost:5432 (database: salon)
```

### Individual Service Development
```bash
# User Service
cd user-service && go run ./cmd

# Salon Service  
cd salon-service && go run ./cmd
```

##  Available Commands

```bash
# Build & Test
make build          # Build all services
make test           # Run all tests
make tidy           # Update dependencies

# Docker & Local Development
make docker-build   # Build Docker images
make local-up       # Start local environment
make local-down     # Stop local environment

# Deployment
make deploy         # Deploy all to Render
make deploy-user    # Deploy user-service only
make deploy-salon   # Deploy salon-service only
make deploy-status  # Show deployment status

# Configuration Management
make config-list    # List all environment configurations
make config-validate ENV=dev  # Validate environment configs
make config-diff ENV1=dev ENV2=prod  # Compare configs
```

##  Multi-Environment Deployment

### Environment Strategy
- **🔧 DEV**: Development environment (auto-deploy from `develop` branch)
- **🧪 STAGE**: Staging environment (manual deploy from `main` branch)
- **🚀 PROD**: Production environment (manual deploy with validation)

### Quick Deployment Commands
```bash
# Deploy to development (auto-deploy enabled)
make deploy-dev

# Deploy to staging (manual approval)
make deploy-stage

# Validate production deployment
make deploy-prod

# Show all environment status
make deploy-status
```

### Environment-Specific Configurations

#### Development Environment
```bash
./deploy-multi-env.sh all dev
# • Branch: develop
# • Auto-deploy: enabled
# • Plan: free
# • Log level: debug
# • Database: salon_dev
```

#### Staging Environment
```bash
./deploy-multi-env.sh all stage
# • Branch: main
# • Auto-deploy: disabled
# • Plan: starter
# • Log level: info
# • Database: salon_stage
```

#### Production Environment
```bash
./deploy-multi-env.sh all prod --validate-only
# • Branch: main
# • Auto-deploy: disabled
# • Plan: standard
# • Log level: warn
# • Database: salon_prod
```

### Service URLs (After Deployment)
- **DEV**: `https://user-service-dev-<id>.onrender.com`
- **STAGE**: `https://user-service-stage-<id>.onrender.com`
- **PROD**: `https://user-service-prod-<id>.onrender.com`

### Deployment Workflow

#### 1. Development Workflow
```bash
# Push to develop branch (auto-deploys to dev)
git checkout develop
git push origin develop

# Services automatically deploy to dev environment
```

#### 2. Staging Workflow
```bash
# Merge to main and manually deploy to stage
git checkout main
git merge develop
git push origin main

# Deploy to staging
make deploy-stage
```

#### 3. Production Workflow
```bash
# Validate production deployment first
make deploy-prod  # Runs validation only

# After validation, deploy manually
./deploy-multi-env.sh all prod
```

##  Configuration

### Environment Variables

#### User Service
```bash
USER_SERVICE_ENV=prod
USER_SERVICE_DB_URL=<postgres-connection-string>
USER_SERVICE_JWT_ACCESSSECRET=<secret>
USER_SERVICE_JWT_REFRESHSECRET=<secret>
USER_SERVICE_JWT_ACCESSTTLMINUTES=15
USER_SERVICE_JWT_REFRESHTTLDAYS=7
USER_SERVICE_OTP_EXPIRYMINUTES=5
```

#### Salon Service
```bash
SALON_SERVICE_ENV=prod
SALON_SERVICE_DB_URL=<postgres-connection-string>
SALON_SERVICE_JWT_ACCESSSECRET=<secret>
SALON_SERVICE_JWT_REFRESHSECRET=<secret>
SALON_SERVICE_JWT_ACCESSTTLMINUTES=15
SALON_SERVICE_JWT_REFRESHTTLDAYS=7
SALON_SERVICE_OTP_EXPIRYMINUTES=5
```

##  Database Schema

Both services share a single PostgreSQL database named `salon`:

- **User Service Tables**: `users`, `otps`, `tokens`
- **Salon Service Tables**: `salons`, `branches`, `categories`, `services`, `staff`, `staff_auth`, `staff_services`

##  Security Features

- **JWT Token Type Validation**: Separate token types for customers and salon staff
- **Rate Limiting**: Configurable OTP request limits per service
- **Audit Logging**: Comprehensive logging of sensitive operations
- **Scoped Authorization**: Users/staff can only access their own data
- **Input Validation**: Unified validation across all services

##  API Documentation

### User Service Endpoints
- `POST /otp/request` - Request OTP for phone number
- `POST /user/register` - Register new user with OTP
- `POST /user/authenticate` - Authenticate with phone + OTP
- `GET /user/{id}` - Get user profile (protected)
- `PUT /user/{id}` - Update user profile (protected)
- `POST /user/refresh` - Refresh access token

### Salon Service Endpoints
- `POST /otp/staff/request` - Request staff OTP
- `POST /staff/authenticate` - Staff authentication
- `GET /salons` - List salons
- `POST /salons` - Create salon
- `GET /salons/{id}` - Get salon details
- Staff, services, categories, and branch management endpoints

##  Development

### Project Structure
salon/
├── user-service/              # Customer API service
├── salon-service/             # Business management API
├── salon-shared/              # Shared components
├── config/                    # Configuration management
│   ├── environments/          # Environment-specific configs
│   │   ├── dev/              # Development environment
│   │   │   ├── render.yaml   # Render deployment config
│   │   │   ├── *.yaml        # Service configurations
│   │   │   └── secrets.env   # Environment secrets
│   │   ├── stage/            # Staging environment
│   │   └── prod/             # Production environment
│   └── README.md             # Configuration guide
├── Dockerfile.user-service    # User service container
├── Dockerfile.salon-service   # Salon service container
├── docker-compose.yml        # Local development
├── deploy-multi-env.sh       # Multi-environment deployment
├── config-manager.sh         # Configuration management
├── copy-secrets.sh           # Secret management helper
└── Makefile                  # Build automation

### Adding New Services

1. Create service directory
3. Add service to `docker-compose.yml`
4. Update `render.yaml` for production
5. Add build targets to `Makefile`

##  Service Documentation

- [User Service Documentation](./user-service/README.md)
- [Salon Service Documentation](./salon-service/README.md)

##  Contributing

1. Fork the repository
2. Create feature branch
3. Use shared components from `salon-shared`
4. Add tests for new functionality
5. Update documentation
6. Submit pull request

##  License

This project is licensed under the MIT License.