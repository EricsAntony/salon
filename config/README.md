# 📁 Configuration Management

This directory contains environment-specific configuration files for the salon platform.

## 🏗️ Directory Structure

```
config/
├── environments/
│   ├── dev/
│   │   ├── render.yaml           # Render deployment config for dev
│   │   ├── user-service.yaml     # User service config for dev
│   │   └── salon-service.yaml    # Salon service config for dev
│   ├── stage/
│   │   ├── render.yaml           # Render deployment config for stage
│   │   ├── user-service.yaml     # User service config for stage
│   │   └── salon-service.yaml    # Salon service config for stage
│   └── prod/
│       ├── render.yaml           # Render deployment config for prod
│       ├── user-service.yaml     # User service config for prod
│       └── salon-service.yaml    # Salon service config for prod
└── README.md                     # This file
```

## 🔧 Environment Configurations

### Development Environment (`dev/`)
- **Purpose**: Local development and testing
- **Features**: 
  - Debug logging enabled
  - Longer JWT token TTL for easier development
  - More lenient rate limiting
  - Local database connections
- **Auto-deploy**: Enabled from `develop` branch

### Staging Environment (`stage/`)
- **Purpose**: Pre-production testing and validation
- **Features**:
  - Production-like settings
  - Info level logging
  - Standard JWT token TTL
  - External database connections with SSL
- **Auto-deploy**: Disabled (manual deployment)

### Production Environment (`prod/`)
- **Purpose**: Live production environment
- **Features**:
  - Minimal logging (warn level only)
  - Short JWT token TTL for security
  - Strict rate limiting
  - Secure database connections
- **Auto-deploy**: Disabled (manual deployment with validation)

## 🚀 Usage

### Deployment Scripts
The deployment scripts automatically use the correct configuration files based on the environment:

```bash
# Deploy to development
./deploy-multi-env.sh all dev
# Uses: config/environments/dev/render.yaml

# Deploy to staging  
./deploy-multi-env.sh all stage
# Uses: config/environments/stage/render.yaml

# Deploy to production
./deploy-multi-env.sh all prod
# Uses: config/environments/prod/render.yaml
```

### Service Configuration
Services can load environment-specific configurations:

```bash
# Development
export CONFIG_PATH=config/environments/dev/user-service.yaml
./user-service

# Production
export CONFIG_PATH=config/environments/prod/user-service.yaml
./user-service
```

## 🔐 Security Notes

### Development
- Uses placeholder secrets (change for actual development)
- Local database connections
- Relaxed security settings for ease of development

### Staging & Production
- Secrets should be set via environment variables
- Database connections use SSL/TLS
- Enhanced security settings
- External database endpoints

## 📝 Configuration Management

### Adding New Environments
1. Create new directory: `config/environments/{env-name}/`
2. Copy and modify configuration files from existing environment
3. Update deployment scripts to recognize new environment
4. Add environment-specific secrets

### Modifying Configurations
1. **Development**: Edit files directly in `dev/` folder
2. **Staging/Production**: 
   - Update configuration files
   - Test in staging first
   - Deploy to production after validation

### Environment Variables Override
All configuration files support environment variable overrides:
- `SERVICE_NAME_JWT_ACCESSSECRET`
- `SERVICE_NAME_JWT_REFRESHSECRET`
- `SERVICE_NAME_DB_URL`
- etc.

## 🔗 Related Files

- **Deployment Scripts**: `deploy-multi-env.sh`
- **Secret Management**: `copy-secrets.sh`
- **Secret Templates**: `render-secrets-{env}.env`
- **Documentation**: `SECRETS-SETUP.md`

---

**⚠️ Important**: Never commit actual secrets to version control. Use environment variables or secure secret management systems for sensitive data.
