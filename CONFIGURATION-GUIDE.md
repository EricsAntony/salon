# 📁 Configuration Management Guide

This guide covers the organized configuration structure for the salon platform's multi-environment deployment.

## 🏗️ New Organized Structure

```
salon/
├── config/
│   ├── environments/
│   │   ├── dev/
│   │   │   ├── render.yaml           # Render deployment config
│   │   │   ├── user-service.yaml     # User service configuration
│   │   │   ├── salon-service.yaml    # Salon service configuration
│   │   │   └── secrets.env           # Environment secrets (gitignored)
│   │   ├── stage/
│   │   │   ├── render.yaml
│   │   │   ├── user-service.yaml
│   │   │   ├── salon-service.yaml
│   │   │   └── secrets.env
│   │   └── prod/
│   │       ├── render.yaml
│   │       ├── user-service.yaml
│   │       ├── salon-service.yaml
│   │       └── secrets.env
│   └── README.md                     # Configuration documentation
├── config-manager.sh                 # Configuration management tool
├── deploy-multi-env.sh               # Multi-environment deployment
└── copy-secrets.sh                   # Secret management helper
```

## ✅ **Benefits of Organized Structure**

### **🎯 Environment Isolation**
- **Clear separation** of dev, stage, and prod configurations
- **Easy identification** of environment-specific settings
- **Reduced configuration errors** through organized structure

### **🔧 Maintainability**
- **Centralized configuration** management in `config/` directory
- **Consistent naming** across all environments
- **Version-controlled** configuration changes

### **🚀 Deployment Efficiency**
- **Automated configuration** selection based on environment
- **Validation tools** to ensure configuration correctness
- **Comparison tools** to diff configurations between environments

## 📋 **Available Commands**

### **Configuration Management**
```bash
# List all configurations
make config-list
./config-manager.sh list

# Show specific configuration
./config-manager.sh show dev user-service
./config-manager.sh show dev secrets      # Show environment secrets

# Validate environment configurations
make config-validate ENV=dev
./config-manager.sh validate prod

# Compare configurations between environments
make config-diff ENV1=dev ENV2=prod
./config-manager.sh diff stage prod

# Copy configurations from one environment to another
./config-manager.sh copy dev stage
```

### **Deployment with New Structure**
```bash
# Deploy using organized configs
make deploy-dev     # Uses config/environments/dev/render.yaml
make deploy-stage   # Uses config/environments/stage/render.yaml
make deploy-prod    # Uses config/environments/prod/render.yaml

# Show deployment status
make deploy-status
./deploy-multi-env.sh status
```

## 🔧 **Environment-Specific Features**

### **Development Environment (`config/environments/dev/`)**
- **Debug logging** enabled for troubleshooting
- **Longer JWT TTL** (60 minutes) for easier development
- **Relaxed rate limiting** (5 OTP requests per minute)
- **Local database** connections
- **Auto-deployment** from `develop` branch

### **Staging Environment (`config/environments/stage/`)**
- **Production-like settings** for realistic testing
- **Info level logging** for balanced visibility
- **Standard JWT TTL** (15 minutes)
- **External database** with SSL
- **Manual deployment** from `main` branch

### **Production Environment (`config/environments/prod/`)**
- **Minimal logging** (warn level only) for performance
- **Short JWT TTL** (15 minutes) for security
- **Strict rate limiting** (3 OTP requests per minute)
- **Secure database** connections with SSL
- **Manual deployment** with validation

## 🔐 **Security Considerations**

### **Configuration Security**
- **No secrets in config files** - use environment variables
- **SSL/TLS enforced** in staging and production
- **Environment-specific** database credentials
- **Separate JWT secrets** per environment

### **Access Control**
- **Development**: Relaxed for ease of development
- **Staging**: Production-like security for testing
- **Production**: Maximum security settings

## 📊 **Configuration Validation**

### **Automatic Validation**
The `config-manager.sh` script provides:
- **YAML syntax validation** using Python YAML parser
- **Required file checks** for each environment
- **Configuration completeness** verification

### **Manual Validation**
```bash
# Validate all environments
./config-manager.sh validate dev
./config-manager.sh validate stage
./config-manager.sh validate prod

# Compare configurations
./config-manager.sh diff dev prod
```

## 🔄 **Migration from Old Structure**

### **What Changed**
- ✅ **Moved**: `render-dev.yaml` → `config/environments/dev/render.yaml`
- ✅ **Moved**: `render-stage.yaml` → `config/environments/stage/render.yaml`
- ✅ **Moved**: `render-prod.yaml` → `config/environments/prod/render.yaml`
- ✅ **Added**: Service-specific configuration files per environment
- ✅ **Updated**: Deployment scripts to use new paths

### **Backward Compatibility**
- **Deployment scripts** automatically use new paths
- **Make commands** work with organized structure
- **Secret management** remains unchanged

## 🛠️ **Adding New Environments**

### **Step-by-Step Process**
1. **Create directory**: `mkdir -p config/environments/new-env`
2. **Copy base configs**: `./config-manager.sh copy dev new-env`
3. **Update environment values**: Edit configs to match new environment
4. **Update deployment scripts**: Add new environment to validation
5. **Test configuration**: `./config-manager.sh validate new-env`

### **Example: Adding QA Environment**
```bash
# Create QA environment
mkdir -p config/environments/qa

# Copy from staging as base
./config-manager.sh copy stage qa

# Update QA-specific values
# Edit config/environments/qa/*.yaml files

# Validate
./config-manager.sh validate qa
```

## 📝 **Best Practices**

### **Configuration Management**
- **Always validate** configurations before deployment
- **Use environment variables** for secrets and sensitive data
- **Keep configurations** in version control
- **Document changes** in configuration files

### **Environment Consistency**
- **Use staging** to test production-like configurations
- **Keep development** flexible for rapid iteration
- **Maintain security** standards in production
- **Regular configuration** reviews and updates

## 🔗 **Related Documentation**

- [Configuration README](./config/README.md) - Detailed configuration guide
- [Secrets Setup Guide](./SECRETS-SETUP.md) - JWT secrets management
- [Deployment Guide](./README.md) - Main deployment documentation

---

**✅ The salon platform now has a clean, organized, and maintainable configuration structure that supports professional multi-environment deployment!**
