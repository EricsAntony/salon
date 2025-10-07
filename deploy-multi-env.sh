#!/bin/bash

# Salon Platform Multi-Environment Deployment Script
# Usage: ./deploy-multi-env.sh [service] [environment] [options]
# 
# Services: user-service, salon-service, booking-service, payment-service, notification-service, all
# Environments: dev, stage, prod
# Options: --auto-deploy, --validate-only, --generate-secrets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_SERVICE="all"
DEFAULT_ENV="dev"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_env() {
    echo -e "${PURPLE}[${1^^}]${NC} $2"
}

# Validate environment
validate_environment() {
    local env=$1
    case $env in
        dev|stage|prod)
            return 0
            ;;
        *)
            log_error "Invalid environment: $env. Must be dev, stage, or prod"
            return 1
            ;;
    esac
}

# Validate service
validate_service() {
    local service=$1
    case $service in
        user-service|salon-service|booking-service|payment-service|notification-service|all)
            return 0
            ;;
        *)
            log_error "Invalid service: $service. Must be user-service, salon-service, booking-service, payment-service, notification-service, or all"
            return 1
            ;;
    esac
}

# Get render config file for environment
get_render_config() {
    local env=$1
    echo "config/environments/${env}/render.yaml"
}

# Check if render config exists
check_render_config() {
    local env=$1
    local config_file=$(get_render_config $env)
    
    if [ ! -f "$config_file" ]; then
        log_error "Render config not found: $config_file"
        return 1
    fi
    
    log_success "Found render config: $config_file"
    return 0
}

# Validate build for environment
validate_build() {
    local env=$1
    log_info "Validating build for environment: $env"
    
    # Set environment-specific build args
    local build_args=""
    case $env in
        dev)
            build_args="--build-arg LOG_LEVEL=debug"
            ;;
        stage)
            build_args="--build-arg LOG_LEVEL=info"
            ;;
        prod)
            build_args="--build-arg LOG_LEVEL=warn"
            ;;
    esac
    
    # Test user-service build
    log_info "Testing user-service build for $env..."
    if ! docker build $build_args -f Dockerfile.user-service -t salon/user-service:$env-test . > /dev/null 2>&1; then
        log_error "user-service Docker build failed for $env"
        return 1
    fi
    
    # Test salon-service build
    log_info "Testing salon-service build for $env..."
    if ! docker build $build_args -f Dockerfile.salon-service -t salon/salon-service:$env-test . > /dev/null 2>&1; then
        log_error "salon-service Docker build failed for $env"
        return 1
    fi
    
    # Test booking-service build
    log_info "Testing booking-service build for $env..."
    if ! docker build $build_args -f Dockerfile.booking-service -t salon/booking-service:$env-test . > /dev/null 2>&1; then
        log_error "booking-service Docker build failed for $env"
        return 1
    fi
    
    # Test payment-service build
    log_info "Testing payment-service build for $env..."
    if ! docker build $build_args -f Dockerfile.payment-service -t salon/payment-service:$env-test . > /dev/null 2>&1; then
        log_error "payment-service Docker build failed for $env"
        return 1
    fi
    
    # Test notification-service build
    log_info "Testing notification-service build for $env..."
    if ! docker build $build_args -f Dockerfile.notification-service -t salon/notification-service:$env-test . > /dev/null 2>&1; then
        log_error "notification-service Docker build failed for $env"
        return 1
    fi
    
    # Clean up test images
    docker rmi salon/user-service:$env-test salon/salon-service:$env-test salon/booking-service:$env-test salon/payment-service:$env-test salon/notification-service:$env-test > /dev/null 2>&1 || true
    
    log_success "Build validation completed for $env"
}

# Generate environment-specific secrets (Updated for service-specific structure)
generate_secrets() {
    local env=$1
    
    log_info "Service-specific environment files already exist for $env environment..."
    log_info "Available service environment files:"
    echo "  • config/environments/${env}/user-service.env.sample"
    echo "  • config/environments/${env}/salon-service.env.sample"
    echo "  • config/environments/${env}/booking-service.env.sample"
    echo "  • config/environments/${env}/payment-service.env.sample"
    echo "  • config/environments/${env}/notification-service.env.sample"
    echo ""
    log_info "Use these files to set environment variables in Render dashboard for each service"
}

# Deploy to specific environment
deploy_environment() {
    local service=$1
    local env=$2
    local auto_deploy=${3:-false}
    
    log_env $env "Deploying $service to $env environment"
    
    local config_file=$(get_render_config $env)
    local suffix=""
    
    case $env in
        dev)
            suffix="-dev"
            ;;
        stage)
            suffix="-stage"
            ;;
        prod)
            suffix="-prod"
            ;;
    esac
    
    case $service in
        "user-service")
            log_env $env "User Service will be available at: https://user-service${suffix}-<id>.onrender.com"
            ;;
        "salon-service")
            log_env $env "Salon Service will be available at: https://salon-service${suffix}-<id>.onrender.com"
            ;;
        "booking-service")
            log_env $env "Booking Service will be available at: https://booking-service${suffix}-<id>.onrender.com"
            ;;
        "payment-service")
            log_env $env "Payment Service will be available at: https://payment-service${suffix}-<id>.onrender.com"
            ;;
        "notification-service")
            log_env $env "Notification Service will be available at: https://notification-service${suffix}-<id>.onrender.com"
            ;;
        "all")
            log_env $env "User Service will be available at: https://user-service${suffix}-<id>.onrender.com"
            log_env $env "Salon Service will be available at: https://salon-service${suffix}-<id>.onrender.com"
            log_env $env "Booking Service will be available at: https://booking-service${suffix}-<id>.onrender.com"
            log_env $env "Payment Service will be available at: https://payment-service${suffix}-<id>.onrender.com"
            log_env $env "Notification Service will be available at: https://notification-service${suffix}-<id>.onrender.com"
            log_env $env "Database: salon-db${suffix} (salon_${env} database)"
            ;;
    esac
    
    if [ "$auto_deploy" = true ]; then
        log_env $env "Auto-deployment enabled for $service"
    else
        log_env $env "Manual deployment - trigger in Render dashboard"
    fi
    
    log_env $env "Use render config: $config_file"
}

# Show environment status
show_environment_status() {
    log_info "Multi-Environment Deployment Status"
    echo ""
    echo "📋 Available Environments:"
    echo ""
    
    echo "🔧 DEV Environment:"
    echo "  • Branch: develop"
    echo "  • Auto-deploy: enabled"
    echo "  • Plan: free"
    echo "  • Log level: debug"
    echo "  • Config: config/environments/dev/render.yaml"
    echo ""
    
    echo "🧪 STAGE Environment:"
    echo "  • Branch: main"
    echo "  • Auto-deploy: disabled (manual)"
    echo "  • Plan: starter"
    echo "  • Log level: info"
    echo "  • Config: config/environments/stage/render.yaml"
    echo ""
    
    echo "🚀 PROD Environment:"
    echo "  • Branch: main"
    echo "  • Auto-deploy: disabled (manual)"
    echo "  • Plan: standard"
    echo "  • Log level: warn"
    echo "  • Config: config/environments/prod/render.yaml"
    echo ""
    
    echo "🌐 Service URLs (after deployment):"
    echo "  User Service:"
    echo "    • DEV:   https://user-service-dev-<id>.onrender.com"
    echo "    • STAGE: https://user-service-stage-<id>.onrender.com"
    echo "    • PROD:  https://user-service-prod-<id>.onrender.com"
    echo ""
    echo "  Salon Service:"
    echo "    • DEV:   https://salon-service-dev-<id>.onrender.com"
    echo "    • STAGE: https://salon-service-stage-<id>.onrender.com"
    echo "    • PROD:  https://salon-service-prod-<id>.onrender.com"
    echo ""
    echo "  Booking Service:"
    echo "    • DEV:   https://booking-service-dev-<id>.onrender.com"
    echo "    • STAGE: https://booking-service-stage-<id>.onrender.com"
    echo "    • PROD:  https://booking-service-prod-<id>.onrender.com"
    echo ""
    echo "  Payment Service:"
    echo "    • DEV:   https://payment-service-dev-<id>.onrender.com"
    echo "    • STAGE: https://payment-service-stage-<id>.onrender.com"
    echo "    • PROD:  https://payment-service-prod-<id>.onrender.com"
    echo ""
    echo "  Notification Service:"
    echo "    • DEV:   https://notification-service-dev-<id>.onrender.com"
    echo "    • STAGE: https://notification-service-stage-<id>.onrender.com"
    echo "    • PROD:  https://notification-service-prod-<id>.onrender.com"
    echo ""
    
    echo "📊 Database Names:"
    echo "  • DEV:   salon_dev"
    echo "  • STAGE: salon_stage"
    echo "  • PROD:  salon_prod"
}

# Main deployment logic
main() {
    local service=${1:-$DEFAULT_SERVICE}
    local env=${2:-$DEFAULT_ENV}
    local validate_only=false
    local generate_secrets_only=false
    local auto_deploy_flag=""
    
    # Parse arguments
    for arg in "$@"; do
        case $arg in
            --auto-deploy)
                auto_deploy_flag="true"
                ;;
            --validate-only)
                validate_only=true
                ;;
            --generate-secrets)
                generate_secrets_only=true
                ;;
        esac
    done
    
    echo "🚀 Salon Platform Multi-Environment Deployment"
    echo "=============================================="
    
    # Handle special commands first
    if [ "$service" = "status" ]; then
        show_environment_status
        exit 0
    fi
    
    # Validate inputs
    if ! validate_service $service; then
        exit 1
    fi
    
    if ! validate_environment $env; then
        exit 1
    fi
    
    # Check render config exists
    if ! check_render_config $env; then
        exit 1
    fi
    
    # Generate secrets if requested
    if [ "$generate_secrets_only" = true ]; then
        generate_secrets $env
        exit 0
    fi
    
    # Validate build
    validate_build $env
    
    # Generate secrets
    generate_secrets $env
    
    # Exit if validate only
    if [ "$validate_only" = true ]; then
        log_success "Validation completed for $service in $env environment"
        exit 0
    fi
    
    # Handle special commands
    if [ "$service" = "status" ]; then
        show_environment_status
        exit 0
    fi
    
    # Deploy
    case $service in
        "user-service"|"salon-service"|"booking-service"|"payment-service"|"notification-service"|"all")
            deploy_environment $service $env $auto_deploy_flag
            ;;
    esac
    
    show_environment_status
    
    echo ""
    log_success "Deployment preparation completed for $service in $env environment!"
    echo ""
    echo "📝 Next Steps:"
    echo "1. Push your code to the appropriate branch:"
    echo "   • DEV: git push origin develop"
    echo "   • STAGE/PROD: git push origin main"
    echo "2. Connect your GitHub repo to Render"
    echo "3. Create services using $(get_render_config $env) blueprint"
    echo "4. Set environment variables from config/environments/${env}/<service>.env.sample files"
    echo "5. Trigger deployment (auto for dev, manual for stage/prod)"
    echo ""
    echo "📚 Documentation: https://render.com/docs/deploy-from-github"
}

# Show usage if no arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 [service] [environment] [options]"
    echo ""
    echo "Services:"
    echo "  user-service          Deploy only user service"
    echo "  salon-service         Deploy only salon service"
    echo "  booking-service       Deploy only booking service"
    echo "  payment-service       Deploy only payment service"
    echo "  notification-service  Deploy only notification service"
    echo "  all                   Deploy all services (default)"
    echo "  status          Show environment status"
    echo ""
    echo "Environments:"
    echo "  dev            Development environment (default)"
    echo "  stage          Staging environment"
    echo "  prod           Production environment"
    echo ""
    echo "Options:"
    echo "  --auto-deploy     Enable automatic deployment"
    echo "  --validate-only   Only validate, don't deploy"
    echo "  --generate-secrets Generate secrets template only"
    echo ""
    echo "Examples:"
    echo "  $0 all dev                         # Deploy all to dev"
    echo "  $0 user-service stage              # Deploy user-service to stage"
    echo "  $0 booking-service dev             # Deploy booking-service to dev"
    echo "  $0 payment-service stage           # Deploy payment-service to stage"
    echo "  $0 notification-service dev        # Deploy notification-service to dev"
    echo "  $0 all prod --validate-only        # Validate prod deployment"
    echo "  $0 status                          # Show environment status"
    exit 1
fi

# Run main function with all arguments
main "$@"
