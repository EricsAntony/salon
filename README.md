A comprehensive salon management platform built with microservices architecture, featuring shared components and production-ready deployment.

##  Architecture

### Services
- **user-service** (Port 8080): Customer-facing API for user registration, authentication, and profile management
- **salon-service** (Port 8081): Business management API for salon operations, staff, and services
- **booking-service** (Port 8082): Booking management API for appointments, availability, and scheduling
- **payment-service** (Port 8083): Payment processing API with Stripe and Razorpay integration
- **notification-service** (Port 8084): Multi-channel notification API for email, SMS, and push notifications
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
# User Service:         http://localhost:8080
# Salon Service:        http://localhost:8081
# Booking Service:      http://localhost:8082
# Payment Service:      http://localhost:8083
# Notification Service: http://localhost:8084
# PostgreSQL:           localhost:5432 (database: salon)
```

### Individual Service Development
```bash
# User Service
cd user-service && go run ./cmd

# Salon Service  
cd salon-service && go run ./cmd

# Booking Service
cd booking-service && go run ./cmd

# Payment Service
cd payment-service && go run ./cmd

# Notification Service
cd notification-service && go run ./cmd
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
make deploy-user         # Deploy user-service only
make deploy-salon        # Deploy salon-service only
make deploy-booking      # Deploy booking-service only
make deploy-payment      # Deploy payment-service only
make deploy-notification # Deploy notification-service only
make deploy-status       # Show deployment status

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

#### User Service
- **DEV**: `https://user-service-dev-<id>.onrender.com`
- **STAGE**: `https://user-service-stage-<id>.onrender.com`
- **PROD**: `https://user-service-prod-<id>.onrender.com`

#### Salon Service
- **DEV**: `https://salon-service-dev-<id>.onrender.com`
- **STAGE**: `https://salon-service-stage-<id>.onrender.com`
- **PROD**: `https://salon-service-prod-<id>.onrender.com`

#### Booking Service
- **DEV**: `https://booking-service-dev-<id>.onrender.com`
- **STAGE**: `https://booking-service-stage-<id>.onrender.com`
- **PROD**: `https://booking-service-prod-<id>.onrender.com`

#### Payment Service
- **DEV**: `https://payment-service-dev-<id>.onrender.com`
- **STAGE**: `https://payment-service-stage-<id>.onrender.com`
- **PROD**: `https://payment-service-prod-<id>.onrender.com`

#### Notification Service
- **DEV**: `https://notification-service-dev-<id>.onrender.com`
- **STAGE**: `https://notification-service-stage-<id>.onrender.com`
- **PROD**: `https://notification-service-prod-<id>.onrender.com`

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

#### Booking Service
```bash
BOOKING_SERVICE_ENV=prod
BOOKING_SERVICE_DB_URL=<postgres-connection-string>
BOOKING_SERVICE_JWT_ACCESSSECRET=<secret>
BOOKING_SERVICE_JWT_REFRESHSECRET=<secret>
BOOKING_SERVICE_JWT_ACCESSTTLMINUTES=15
BOOKING_SERVICE_JWT_REFRESHTTLDAYS=7
USER_SERVICE_URL=https://user-service-prod-<id>.onrender.com
SALON_SERVICE_URL=https://salon-service-prod-<id>.onrender.com
PAYMENT_SERVICE_URL=https://payment-service-prod-<id>.onrender.com
NOTIFICATION_SERVICE_URL=https://notification-service-prod-<id>.onrender.com
```

#### Payment Service
```bash
PAYMENT_SERVICE_ENV=prod
PAYMENT_SERVICE_DB_URL=<postgres-connection-string>
PAYMENT_SERVICE_STRIPE_SECRET_KEY=<stripe-live-secret>
PAYMENT_SERVICE_STRIPE_WEBHOOK_SECRET=<stripe-webhook-secret>
PAYMENT_SERVICE_RAZORPAY_KEY_ID=<razorpay-live-key-id>
PAYMENT_SERVICE_RAZORPAY_KEY_SECRET=<razorpay-live-secret>
PAYMENT_SERVICE_RAZORPAY_WEBHOOK_SECRET=<razorpay-webhook-secret>
```

#### Notification Service
```bash
NOTIFICATION_SERVICE_ENV=prod
NOTIFICATION_SERVICE_DB_URL=<postgres-connection-string>
NOTIFICATION_SERVICE_SMTP_HOST=<smtp-host>
NOTIFICATION_SERVICE_SMTP_PORT=587
NOTIFICATION_SERVICE_SMTP_USERNAME=<smtp-username>
NOTIFICATION_SERVICE_SMTP_PASSWORD=<smtp-password>
NOTIFICATION_SERVICE_SMTP_FROM=noreply@salon.com
NOTIFICATION_SERVICE_TWILIO_ACCOUNT_SID=<twilio-account-sid>
NOTIFICATION_SERVICE_TWILIO_AUTH_TOKEN=<twilio-auth-token>
NOTIFICATION_SERVICE_TWILIO_FROM=<twilio-phone-number>
NOTIFICATION_SERVICE_FCM_SERVER_KEY=<fcm-server-key>
NOTIFICATION_SERVICE_FCM_PROJECT_ID=<fcm-project-id>

##  Database Schema

All services share a single PostgreSQL database named `salon`:

- **User Service Tables**: `users`, `otps`, `tokens`
- **Salon Service Tables**: `salons`, `branches`, `categories`, `services`, `staff`, `staff_auth`, `staff_services`
- **Booking Service Tables**: `bookings`, `booking_services`, `booking_history`, `branch_configurations`
- **Payment Service Tables**: `payments`, `payment_methods`, `transactions`, `refunds`
- **Notification Service Tables**: `notifications`, `notification_templates`, `notification_logs`, `notification_preferences`

##  Security Features

- **JWT Token Type Validation**: Separate token types for customers and salon staff
- **Rate Limiting**: Configurable OTP request limits per service
- **Audit Logging**: Comprehensive logging of sensitive operations
- **Scoped Authorization**: Users/staff can only access their own data
- **Input Validation**: Unified validation across all services

##  API Documentation
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

### Booking Service Endpoints
- `POST /bookings/initiate` - Initiate new booking (protected)
- `POST /bookings/confirm` - Confirm booking with payment (protected)
- `GET /bookings/{id}` - Get booking details (protected)
- `GET /bookings/user/{userId}` - Get user bookings (protected)
- `PATCH /bookings/{id}/cancel` - Cancel booking (protected)
- `PATCH /bookings/{id}/reschedule` - Reschedule booking (protected)
- `GET /stylists/{id}/availability` - Get stylist availability
- `POST /bookings/summary` - Calculate booking pricing
- `GET /branches/{id}/config` - Get branch configuration

##  Development

### Project Structure
salon/
├── user-service/              # Customer API service
├── salon-service/             # Business management API
├── booking-service/           # Booking management API
├── payment-service/           # Payment processing API
├── notification-service/      # Multi-channel notification API
├── salon-shared/              # Shared components
├── config/                    # Configuration management
│   ├── environments/          # Environment-specific configs
│   │   ├── dev/              # Development environment
│   │   │   ├── render.yaml   # Render deployment config
│   │   │   ├── *.yaml        # Service configurations
│   │   │   └── *.env.sample  # Service-specific env templates
│   │   ├── stage/            # Staging environment
│   │   └── prod/             # Production environment
│   └── README.md             # Configuration guide
├── Dockerfile.user-service    # User service container
├── Dockerfile.salon-service   # Salon service container
├── Dockerfile.booking-service # Booking service container
├── Dockerfile.payment-service # Payment service container
├── Dockerfile.notification-service # Notification service container
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
- [Booking Service Documentation](./booking-service/README.md)
- [Payment Service Documentation](./payment-service/README.md)
- [Notification Service Documentation](./notification-service/README.md)
- [Shared Components Documentation](./salon-shared/README.md)

##  Contributing

1. Fork the repository
2. Create feature branch
3. Use shared components from `salon-shared`
4. Add tests for new functionality
5. Update documentation
6. Submit pull request

##  License

This project is licensed under the MIT License.