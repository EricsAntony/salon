# Environment-Specific Configuration

This directory contains environment-specific configuration files for all services in the salon platform.

## **Directory Structure**

```
config/environments/
├── dev/                    # Development environment
├── stage/                  # Staging environment
├── prod/                   # Production environment
└── README.md              # This file
```

## **Service-Specific Environment Files**

Each environment directory contains **service-specific** `.env.sample` files:

### **Development Environment (`dev/`)**
- `user-service.env.sample` - User service environment variables
- `salon-service.env.sample` - Salon service environment variables
- `booking-service.env.sample` - Booking service environment variables
- `payment-service.env.sample` - Payment service environment variables
- `notification-service.env.sample` - Notification service environment variables

### **Staging Environment (`stage/`)**
- `user-service.env.sample` - User service environment variables
- `salon-service.env.sample` - Salon service environment variables
- `booking-service.env.sample` - Booking service environment variables
- `payment-service.env.sample` - Payment service environment variables
- `notification-service.env.sample` - Notification service environment variables

### **Production Environment (`prod/`)**
- `user-service.env.sample` - User service environment variables
- `salon-service.env.sample` - Salon service environment variables
- `booking-service.env.sample` - Booking service environment variables
- `payment-service.env.sample` - Payment service environment variables
- `notification-service.env.sample` - Notification service environment variables

## **Usage Instructions**

### **1. For Render Deployment**

1. **Choose your environment** (dev/stage/prod)
2. **Select the service** you want to deploy
3. **Copy the relevant `.env.sample` file** content
4. **Set environment variables** in Render dashboard for that specific service
5. **Replace sample values** with actual credentials

**Example for user-service in production:**
```bash
# Copy content from: config/environments/prod/user-service.env.sample
# Set in Render dashboard for: user-service-prod
```

### **2. For Local Development**

1. **Copy the dev environment file** for your service
2. **Rename to `.env`** (without .sample)
3. **Replace sample values** with actual local credentials
4. **Use with your service**

**Example:**
```bash
cp config/environments/dev/user-service.env.sample user-service/.env
# Edit user-service/.env with actual values
```

### **3. For CI/CD Pipelines**

Use service-specific environment files to set up automated deployments:

```bash
# Deploy user-service to staging
./deploy.sh user-service stage

# Deploy payment-service to production
./deploy.sh payment-service prod
```

## **Environment Variable Categories**

Each service environment file contains:

### **Core Configuration**
- `PORT` - Service port number
- `{SERVICE}_ENV` - Environment name (dev/stage/prod)
- `{SERVICE}_SERVER_PORT` - Server port configuration

### **Database Configuration**
- `{SERVICE}_DB_URL` - Database connection string (auto-provided by Render)

### **JWT Configuration** (where applicable)
- `{SERVICE}_JWT_ACCESSSECRET` - JWT access token secret
- `{SERVICE}_JWT_REFRESHSECRET` - JWT refresh token secret
- `{SERVICE}_JWT_ACCESSTTLMINUTES` - Access token TTL
- `{SERVICE}_JWT_REFRESHTTLDAYS` - Refresh token TTL

### **Service-Specific Configuration**
- **Payment Service**: Stripe, Razorpay credentials
- **Notification Service**: SMTP, Twilio, FCM, AWS SNS credentials
- **Booking Service**: Service integration URLs

### **Logging Configuration**
- `{SERVICE}_LOG_LEVEL` - Logging level (debug/info/warn/error)
- `{SERVICE}_LOG_SERVICENAME` - Service name for logs

## **Security Best Practices**

### **Development Environment**
- ✅ Use test/sandbox credentials
- ✅ Longer JWT TTL for easier testing
- ✅ Debug logging enabled
- ✅ Extended OTP expiry for testing

### **Staging Environment**
- ✅ Production-like settings
- ✅ Test credentials but realistic TTL
- ✅ Info level logging
- ✅ Staging-specific endpoints

### **Production Environment**
- ⚠️ **CRITICAL**: Use LIVE credentials only
- ⚠️ **CRITICAL**: Generate cryptographically secure secrets
- ⚠️ **CRITICAL**: Use warn/error logging only
- ⚠️ **CRITICAL**: Regular secret rotation
- ⚠️ **CRITICAL**: Monitor all access and transactions

## **Benefits of Service-Specific Files**

### **Organization**
- ✅ **Clear Separation**: Each service has its own configuration
- ✅ **Easy Management**: Teams can manage their service secrets independently
- ✅ **Reduced Complexity**: No need to scroll through all services to find specific variables

### **Security**
- ✅ **Principle of Least Privilege**: Services only see their own secrets
- ✅ **Isolated Access**: Different teams can manage different services
- ✅ **Audit Trail**: Clear ownership and responsibility per service

### **Deployment**
- ✅ **Independent Deployment**: Deploy services independently with their own configs
- ✅ **Service-Specific Rollback**: Rollback individual service configurations
- ✅ **Targeted Updates**: Update only the service that needs changes

### **Development**
- ✅ **Developer Friendly**: Developers work only with relevant environment variables
- ✅ **Faster Setup**: Quick setup for individual service development
- ✅ **Clear Documentation**: Service-specific notes and requirements

## **Migration from Consolidated Files**

If you were previously using consolidated `secrets.env.sample` files:

1. **Identify your service** from the old consolidated file
2. **Find the corresponding service-specific file** in this new structure
3. **Copy relevant environment variables** to the new file
4. **Update your deployment scripts** to use service-specific files
5. **Remove old consolidated files** once migration is complete

This new structure provides better organization, security, and maintainability for the salon platform's environment configuration.
