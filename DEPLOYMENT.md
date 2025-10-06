# Salon Platform - Deployment Guide

This guide covers the deployment setup for all services in the Salon Platform microservices architecture.

## 🏗️ Architecture Overview

The Salon Platform consists of 5 microservices:

- **user-service** (Port 8080): Customer-facing API for registration, authentication, profile management
- **salon-service** (Port 8081): Business management API for salon operations, staff, services
- **booking-service** (Port 8082): Booking management and coordination service
- **payment-service** (Port 8083): Payment processing and gateway integration
- **notification-service** (Port 8084): Multi-channel notification delivery
- **salon-db**: Shared PostgreSQL database

## 🚀 Quick Start

### Local Development

1. **Start all services locally:**
   ```bash
   make local-up
   ```

2. **Build all services:**
   ```bash
   make build
   ```

3. **Run tests:**
   ```bash
   make test
   ```

### Individual Service Development

```bash
# Start individual services in development mode
make dev-user          # User service on port 8080
make dev-salon         # Salon service on port 8081
make dev-booking       # Booking service on port 8082
make dev-payment       # Payment service on port 8083
make dev-notification  # Notification service on port 8084
```

## 🐳 Docker Deployment

### Build Docker Images

```bash
# Build all service images
make docker-build

# Build individual service images
docker build -f Dockerfile.user-service -t salon/user-service:latest .
docker build -f Dockerfile.salon-service -t salon/salon-service:latest .
docker build -f Dockerfile.booking-service -t salon/booking-service:latest .
docker build -f Dockerfile.payment-service -t salon/payment-service:latest .
docker build -f Dockerfile.notification-service -t salon/notification-service:latest .
```

### Local Docker Compose

The platform uses a centralized docker-compose.yml for all services:

```bash
# Start all services with dependencies
make local-up

# Or use docker-compose directly
docker-compose up -d

# View logs
make local-logs

# Stop all services
make local-down
```

**Included Services:**
- All 5 microservices (user, salon, booking, payment, notification)
- PostgreSQL database with health checks
- Redis for caching and message queuing
- MailHog for email testing (Web UI: http://localhost:8025)
- Adminer for database management (Web UI: http://localhost:8090)

## ☁️ Cloud Deployment (Render.com)

### Environment Configuration

The platform supports multiple environments:
- **dev**: Development environment with auto-deploy
- **stage**: Staging environment for testing
- **prod**: Production environment with manual approval

### Deployment Commands

```bash
# Deploy to development
make deploy-dev

# Deploy to staging
make deploy-stage

# Deploy to production (validation only)
make deploy-prod

# Check deployment status
make deploy-status
```

### Environment Variables

#### Payment Service
```bash
PAYMENT_SERVICE_ENV=prod
PAYMENT_SERVICE_DB_URL=<auto-provided>
PAYMENT_SERVICE_STRIPE_SECRET_KEY=<set-manually>
PAYMENT_SERVICE_STRIPE_WEBHOOK_SECRET=<set-manually>
PAYMENT_SERVICE_RAZORPAY_KEY_ID=<set-manually>
PAYMENT_SERVICE_RAZORPAY_KEY_SECRET=<set-manually>
```

#### Notification Service
```bash
NOTIFICATION_SERVICE_ENV=prod
NOTIFICATION_SERVICE_DB_URL=<auto-provided>
NOTIFICATION_SERVICE_SMTP_HOST=<set-manually>
NOTIFICATION_SERVICE_SMTP_USERNAME=<set-manually>
NOTIFICATION_SERVICE_SMTP_PASSWORD=<set-manually>
NOTIFICATION_SERVICE_TWILIO_ACCOUNT_SID=<set-manually>
NOTIFICATION_SERVICE_TWILIO_AUTH_TOKEN=<set-manually>
NOTIFICATION_SERVICE_FCM_SERVER_KEY=<set-manually>
```

## 🔧 Service Configuration

### Payment Service Features
- **Payment Gateways**: Stripe, Razorpay support
- **Webhook Handling**: Secure webhook processing
- **Refund Management**: Full and partial refunds
- **Idempotency**: Safe retry mechanisms
- **Rate Limiting**: 100 requests/minute
- **Centralized Configuration**: Uses root-level Dockerfile and config

### Notification Service Features
- **Multi-Channel**: Email, SMS, Push notifications
- **Message Queuing**: Redis support for reliable delivery
- **Template Engine**: Dynamic notification templates
- **Retry Logic**: Configurable retry with backoff
- **Rate Limiting**: 200 requests/minute
- **Development Tools**: MailHog integration for email testing
- **Centralized Configuration**: Uses root-level Dockerfile and config

### Database Schema

All services share the `salon` database with service-specific tables:

- **User Service**: `users`, `otps`, `tokens`
- **Salon Service**: `salons`, `branches`, `categories`, `services`, `staff`, `staff_auth`, `staff_services`
- **Booking Service**: `bookings`, `booking_services`, `booking_history`
- **Payment Service**: `payments`, `refunds`, `payment_gateways`
- **Notification Service**: `notifications`, `notification_templates`, `notification_providers`

## 🔐 Security Features

### Container Security
- **Non-root users**: All containers run as `appuser` (uid: 10001)
- **Minimal base images**: Alpine Linux 3.19
- **Security updates**: Latest security patches applied
- **Health checks**: Container and application level monitoring

### Application Security
- **JWT Token Types**: Service-specific token validation
- **Rate Limiting**: Per-service rate limiting configuration
- **Audit Logging**: Comprehensive operation logging
- **Input Validation**: Unified validation across services
- **CORS Configuration**: Configurable CORS policies

## 📊 Monitoring & Health Checks

### Health Endpoints
All services expose standard health endpoints:
- `GET /health` - Basic health check
- `GET /ready` - Readiness check with database connectivity

### Service URLs (Production)
- **User Service**: `https://user-service-prod.onrender.com`
- **Salon Service**: `https://salon-service-prod.onrender.com`
- **Booking Service**: `https://booking-service-prod.onrender.com`
- **Payment Service**: `https://payment-service-prod.onrender.com`
- **Notification Service**: `https://notification-service-prod.onrender.com`

## 🛠️ Development Tools

### Available Make Commands

```bash
# Build Commands
make build                    # Build all services
make build-payment           # Build payment service only
make build-notification      # Build notification service only

# Test Commands
make test                    # Run tests for all services
make test-payment           # Run payment service tests
make test-notification      # Run notification service tests

# Docker Commands
make docker-build           # Build Docker images
make docker-push            # Push images to registry

# Development Commands
make dev-payment            # Start payment service in dev mode
make dev-notification       # Start notification service in dev mode

# Utility Commands
make clean                  # Clean build artifacts
make tidy                   # Run go mod tidy for all services
make fmt                    # Format code for all services
```

### Database Migrations

```bash
# Payment service migrations
cd payment-service && make migrate-up
cd payment-service && make migrate-down

# Notification service migrations
cd notification-service && make migrate-up
cd notification-service && make migrate-down
```

## 🚨 Troubleshooting

### Common Issues

1. **Port Conflicts**: Ensure ports 8080-8084 are available
2. **Database Connection**: Check database URL and credentials
3. **Service Discovery**: Verify service URLs in configuration
4. **Gateway Credentials**: Ensure payment gateway credentials are set
5. **SMTP Configuration**: Verify email provider settings

### Logs and Debugging

```bash
# View local development logs
make local-logs

# Check individual service logs
docker-compose -f payment-service/docker-compose.yml logs -f
docker-compose -f notification-service/docker-compose.yml logs -f
```

### Health Check Validation

```bash
# Check service health
curl http://localhost:8083/health  # Payment service
curl http://localhost:8084/health  # Notification service

# Check readiness
curl http://localhost:8083/ready   # Payment service
curl http://localhost:8084/ready   # Notification service
```

## 📚 API Documentation

### Payment Service Endpoints
- `POST /api/v1/payments/initiate` - Initiate payment
- `POST /api/v1/payments/confirm` - Confirm payment
- `POST /api/v1/payments/{id}/refund` - Process refund
- `GET /api/v1/bookings/{id}/payments` - Get booking payments

### Notification Service Endpoints
- `POST /api/v1/notifications/send` - Send notification
- `GET /api/v1/notifications/{id}` - Get notification status
- `GET /api/v1/notifications/templates` - List templates

## 🔄 Integration Flow

### Booking with Payment Flow
1. User initiates booking → Booking service creates booking
2. Payment initiated → Payment service processes payment
3. Payment confirmed → Booking service updates status
4. Notifications sent → Notification service delivers confirmations

### Service Communication
- **Booking ↔ Payment**: HTTP REST API calls
- **Booking ↔ Notification**: HTTP REST API calls
- **Payment ↔ Gateways**: Webhook callbacks
- **Notification ↔ Providers**: Provider-specific APIs

## 📈 Scaling Considerations

### Horizontal Scaling
- Each service can be scaled independently
- Database connection pooling configured per service
- Stateless service design enables easy scaling

### Performance Optimization
- Connection pooling for database and external APIs
- Async notification processing
- Caching for frequently accessed data
- Rate limiting to prevent abuse

## 🔒 Production Checklist

- [ ] All environment variables configured
- [ ] Payment gateway credentials set
- [ ] SMTP/SMS provider credentials configured
- [ ] Database migrations applied
- [ ] Health checks passing
- [ ] SSL certificates configured
- [ ] Monitoring and alerting set up
- [ ] Backup strategy implemented
- [ ] Security scan completed

---

For more detailed information, refer to individual service documentation in their respective directories.
