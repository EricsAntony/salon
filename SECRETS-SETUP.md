# 🔐 JWT Secrets Setup Guide

This guide explains how to configure JWT secrets for each environment in your Render deployment.

## 📋 Generated Secret Files

The deployment script has generated environment-specific secret files:

- `render-secrets-dev.env` - Development environment secrets
- `render-secrets-stage.env` - Staging environment secrets  
- `render-secrets-prod.env` - Production environment secrets

## 🔧 Setting Up Secrets in Render

### Step 1: Access Render Dashboard
1. Go to [Render Dashboard](https://dashboard.render.com)
2. Navigate to your service (e.g., `user-service-dev`)
3. Go to **Environment** tab

### Step 2: Add Environment Variables

For each service in each environment, add these secrets:

#### Development Environment (`user-service-dev`, `salon-service-dev`)
```bash
# Copy from render-secrets-dev.env
USER_SERVICE_JWT_ACCESSSECRET=dev-user-access-[generated-hash]
USER_SERVICE_JWT_REFRESHSECRET=dev-user-refresh-[generated-hash]
SALON_SERVICE_JWT_ACCESSSECRET=dev-salon-access-[generated-hash]
SALON_SERVICE_JWT_REFRESHSECRET=dev-salon-refresh-[generated-hash]
```

#### Staging Environment (`user-service-stage`, `salon-service-stage`)
```bash
# Copy from render-secrets-stage.env
USER_SERVICE_JWT_ACCESSSECRET=stage-user-access-[generated-hash]
USER_SERVICE_JWT_REFRESHSECRET=stage-user-refresh-[generated-hash]
SALON_SERVICE_JWT_ACCESSSECRET=stage-salon-access-[generated-hash]
SALON_SERVICE_JWT_REFRESHSECRET=stage-salon-refresh-[generated-hash]
```

#### Production Environment (`user-service-prod`, `salon-service-prod`)
```bash
# Copy from render-secrets-prod.env
USER_SERVICE_JWT_ACCESSSECRET=prod-user-access-[generated-hash]
USER_SERVICE_JWT_REFRESHSECRET=prod-user-refresh-[generated-hash]
SALON_SERVICE_JWT_ACCESSSECRET=prod-salon-access-[generated-hash]
SALON_SERVICE_JWT_REFRESHSECRET=prod-salon-refresh-[generated-hash]
```

## 🛡️ Security Best Practices

### 1. Secret Rotation
- **Development**: Rotate monthly
- **Staging**: Rotate bi-weekly  
- **Production**: Rotate weekly

### 2. Access Control
- Only authorized team members should have access to production secrets
- Use Render's team management features to control access
- Never commit secrets to version control

### 3. Monitoring
- Monitor for unusual JWT token usage patterns
- Set up alerts for failed authentication attempts
- Log JWT token generation and validation events

## 🔄 Rotating Secrets

To rotate JWT secrets for any environment:

```bash
# Generate new secrets for specific environment
./deploy-multi-env.sh all [env] --generate-secrets

# Example: Rotate production secrets
./deploy-multi-env.sh all prod --generate-secrets
```

Then update the secrets in Render dashboard and redeploy the services.

## 📊 Environment-Specific Configurations

### Development Environment
- **Longer token TTL**: For easier development and testing
- **Debug logging**: Enabled for troubleshooting
- **Relaxed security**: Suitable for development workflow

### Staging Environment  
- **Production-like TTL**: Matches production settings
- **Info logging**: Balanced logging for testing
- **Enhanced security**: Production-like security measures

### Production Environment
- **Short token TTL**: Maximum security
- **Minimal logging**: Only warnings and errors
- **Maximum security**: All security features enabled

## 🚨 Emergency Procedures

### If Secrets Are Compromised

1. **Immediate Actions**:
   ```bash
   # Generate new secrets immediately
   ./deploy-multi-env.sh all [env] --generate-secrets
   
   # Update secrets in Render dashboard
   # Redeploy affected services
   ```

2. **Invalidate Existing Tokens**:
   - All existing JWT tokens will become invalid with new secrets
   - Users will need to re-authenticate
   - Monitor for suspicious activity

3. **Post-Incident**:
   - Review access logs
   - Update security procedures
   - Consider additional security measures

## 📝 Checklist for Each Environment

### ✅ Development Setup
- [ ] Set `USER_SERVICE_JWT_ACCESSSECRET` in user-service-dev
- [ ] Set `USER_SERVICE_JWT_REFRESHSECRET` in user-service-dev
- [ ] Set `SALON_SERVICE_JWT_ACCESSSECRET` in salon-service-dev
- [ ] Set `SALON_SERVICE_JWT_REFRESHSECRET` in salon-service-dev
- [ ] Verify services start successfully
- [ ] Test authentication endpoints

### ✅ Staging Setup
- [ ] Set `USER_SERVICE_JWT_ACCESSSECRET` in user-service-stage
- [ ] Set `USER_SERVICE_JWT_REFRESHSECRET` in user-service-stage
- [ ] Set `SALON_SERVICE_JWT_ACCESSSECRET` in salon-service-stage
- [ ] Set `SALON_SERVICE_JWT_REFRESHSECRET` in salon-service-stage
- [ ] Verify services start successfully
- [ ] Test authentication endpoints
- [ ] Run integration tests

### ✅ Production Setup
- [ ] Set `USER_SERVICE_JWT_ACCESSSECRET` in user-service-prod
- [ ] Set `USER_SERVICE_JWT_REFRESHSECRET` in user-service-prod
- [ ] Set `SALON_SERVICE_JWT_ACCESSSECRET` in salon-service-prod
- [ ] Set `SALON_SERVICE_JWT_REFRESHSECRET` in salon-service-prod
- [ ] Verify services start successfully
- [ ] Test authentication endpoints
- [ ] Monitor for 24 hours post-deployment
- [ ] Set up secret rotation schedule

## 🔗 Related Documentation

- [Render Environment Variables](https://render.com/docs/environment-variables)
- [JWT Best Practices](https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/)
- [Salon Platform Security Guide](./SECURITY.md)

---

**⚠️ Important**: Never share these secret files or commit them to version control. Each environment should have unique, strong secrets.
