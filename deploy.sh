#!/bin/bash

# Salon Platform Multi-Environment Deployment Script for Render
# Usage: ./deploy.sh [user-service|salon-service|all] [dev|stage|prod] [--auto-deploy]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
RENDER_API_URL="https://api.render.com/v1"
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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if render.yaml exists
    if [ ! -f "$RENDER_YAML" ]; then
        log_error "render.yaml not found in current directory"
        exit 1
    fi
    
    # Check if Render CLI is installed
    if ! command -v render &> /dev/null; then
        log_warning "Render CLI not found. Install it from: https://render.com/docs/cli"
        log_info "Continuing with manual deployment instructions..."
    fi
    
    # Check if required environment variables are set
    if [ -z "$RENDER_API_KEY" ]; then
        log_warning "RENDER_API_KEY not set. You'll need to configure secrets manually in Render dashboard."
    fi
    
    log_success "Prerequisites check completed"
}

# Validate build
validate_build() {
    log_info "Validating build configuration..."
    
    # Check if all services can build
    log_info "Testing user-service build..."
    if ! docker build -f Dockerfile.user-service -t salon/user-service:test . > /dev/null 2>&1; then
        log_error "user-service Docker build failed"
        exit 1
    fi
    
    log_info "Testing salon-service build..."
    if ! docker build -f Dockerfile.salon-service -t salon/salon-service:test . > /dev/null 2>&1; then
        log_error "salon-service Docker build failed"
        exit 1
    fi
    
    # Clean up test images
    docker rmi salon/user-service:test salon/salon-service:test > /dev/null 2>&1 || true
    
    log_success "Build validation completed"
}

# Deploy specific service
deploy_service() {
    local service_name=$1
    local auto_deploy=${2:-false}
    
    log_info "Deploying $service_name..."
    
    case $service_name in
        "user-service")
            log_info "User Service will be available at: https://user-service-<your-app-id>.onrender.com"
            ;;
        "salon-service")
            log_info "Salon Service will be available at: https://salon-service-<your-app-id>.onrender.com"
            ;;
        *)
            log_error "Unknown service: $service_name"
            exit 1
            ;;
    esac
    
    if [ "$auto_deploy" = true ]; then
        log_info "Auto-deployment enabled for $service_name"
    else
        log_info "Manual deployment - remember to trigger deployment in Render dashboard"
    fi
}

# Deploy all services
deploy_all() {
    local auto_deploy=${1:-false}
    
    log_info "Deploying all services..."
    
    # Deploy database first
    log_info "Database 'salon-db' will be created automatically"
    
    # Deploy services
    deploy_service "user-service" $auto_deploy
    deploy_service "salon-service" $auto_deploy
    
    log_success "All services deployment initiated"
}

# Generate environment variables template
generate_env_template() {
    log_info "Generating environment variables template..."
    
    cat > render-secrets.env << EOF
# Render Environment Variables Template
# Copy these to your Render service settings

# JWT Secrets (Generate strong random strings)
USER_SERVICE_JWT_ACCESSSECRET=your-user-access-secret-here
USER_SERVICE_JWT_REFRESHSECRET=your-user-refresh-secret-here
SALON_SERVICE_JWT_ACCESSSECRET=your-salon-access-secret-here
SALON_SERVICE_JWT_REFRESHSECRET=your-salon-refresh-secret-here

# Optional: Override default values
# USER_SERVICE_JWT_ACCESSTTLMINUTES=15
# USER_SERVICE_JWT_REFRESHTTLDAYS=7
# USER_SERVICE_OTP_EXPIRYMINUTES=5
# SALON_SERVICE_JWT_ACCESSTTLMINUTES=15
# SALON_SERVICE_JWT_REFRESHTTLDAYS=7
# SALON_SERVICE_OTP_EXPIRYMINUTES=5

# Database connection strings are automatically provided by Render
EOF
    
    log_success "Environment template created: render-secrets.env"
    log_warning "Remember to set these secrets in Render dashboard for each service"
}

# Show deployment status
show_status() {
    log_info "Deployment Status:"
    echo ""
    echo "📋 Services to deploy:"
    echo "  • user-service  (Customer API - Port 8080)"
    echo "  • salon-service (Business API - Port 8081)"
    echo "  • salon-db      (PostgreSQL Database)"
    echo ""
    echo "🔗 After deployment:"
    echo "  • User Service:  https://user-service-<id>.onrender.com"
    echo "  • Salon Service: https://salon-service-<id>.onrender.com"
    echo ""
    echo "⚙️  Configuration files:"
    echo "  • render.yaml           - Render deployment config"
    echo "  • Dockerfile.user-service   - User service container"
    echo "  • Dockerfile.salon-service  - Salon service container"
    echo "  • render-secrets.env    - Environment variables template"
}

# Main deployment logic
main() {
    local service=${1:-"all"}
    local auto_deploy_flag=""
    
    # Parse arguments
    for arg in "$@"; do
        case $arg in
            --auto-deploy)
                auto_deploy_flag="true"
                ;;
        esac
    done
    
    echo "🚀 Salon Platform Deployment Script"
    echo "===================================="
    
    check_prerequisites
    validate_build
    generate_env_template
    
    case $service in
        "user-service"|"salon-service")
            deploy_service $service $auto_deploy_flag
            ;;
        "all")
            deploy_all $auto_deploy_flag
            ;;
        "status")
            show_status
            ;;
        *)
            echo "Usage: $0 [user-service|salon-service|all|status] [--auto-deploy]"
            echo ""
            echo "Commands:"
            echo "  user-service    Deploy only user service"
            echo "  salon-service   Deploy only salon service"  
            echo "  all            Deploy all services (default)"
            echo "  status         Show deployment status"
            echo ""
            echo "Options:"
            echo "  --auto-deploy  Enable automatic deployment on git push"
            exit 1
            ;;
    esac
    
    show_status
    
    echo ""
    log_success "Deployment script completed!"
    echo ""
    echo "📝 Next Steps:"
    echo "1. Push your code to GitHub"
    echo "2. Connect your GitHub repo to Render"
    echo "3. Create services using render.yaml blueprint"
    echo "4. Set environment variables from render-secrets.env"
    echo "5. Trigger deployment"
    echo ""
    echo "📚 Documentation: https://render.com/docs/deploy-from-github"
}

# Run main function with all arguments
main "$@"
