#!/bin/bash

# Service-Specific Environment Variables Copy Helper
# Usage: ./copy-secrets.sh [environment] [service]
# Helps copy environment-specific variables for Render setup

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

show_usage() {
    echo "Usage: $0 [environment] [service]"
    echo ""
    echo "Environments: dev, stage, prod"
    echo "Services: user, salon, booking, payment, notification, all"
    echo ""
    echo "Examples:"
    echo "  $0 dev user          # Show user-service env vars for dev"
    echo "  $0 prod salon        # Show salon-service env vars for prod"
    echo "  $0 stage booking     # Show booking-service env vars for stage"
    echo "  $0 dev payment       # Show payment-service env vars for dev"
    echo "  $0 prod notification # Show notification-service env vars for prod"
    echo "  $0 stage all         # Show all service env vars for stage"
}

show_service_env() {
    local env=$1
    local service_name=$2
    local env_file="config/environments/${env}/${service_name}-service.env.sample"
    
    if [ ! -f "$env_file" ]; then
        echo -e "${RED}Error: Environment file not found: $env_file${NC}"
        return 1
    fi
    
    echo -e "${GREEN}🔑 ${service_name^} Service Environment Variables (copy to ${service_name}-service-${env} in Render):${NC}"
    echo ""
    
    # Show all environment variables from the service-specific file
    grep -v "^#" "$env_file" | grep -v "^$" | while read line; do
        echo "  $line"
    done
    echo ""
    
    # Show Render dashboard link
    echo -e "${BLUE}🔗 Render Dashboard Link:${NC}"
    echo "  ${service_name^} Service: https://dashboard.render.com/web/srv-[your-${service_name}-service-${env}-id]/env"
    echo ""
}

extract_secrets() {
    local env=$1
    local service=$2
    
    echo -e "${BLUE}📋 Environment Variables for ${env^^} Environment${NC}"
    echo "================================================"
    echo ""
    
    case $service in
        "user")
            show_service_env $env "user"
            ;;
        "salon")
            show_service_env $env "salon"
            ;;
        "booking")
            show_service_env $env "booking"
            ;;
        "payment")
            show_service_env $env "payment"
            ;;
        "notification")
            show_service_env $env "notification"
            ;;
        "all")
            show_service_env $env "user"
            show_service_env $env "salon"
            show_service_env $env "booking"
            show_service_env $env "payment"
            show_service_env $env "notification"
            ;;
        *)
            echo -e "${RED}Error: Invalid service: $service${NC}"
            show_usage
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}✅ Next Steps:${NC}"
    echo "1. Copy the environment variables above to your Render service(s)"
    echo "2. Replace sample values with actual credentials"
    echo "3. Redeploy the service(s) to apply the new environment variables"
    echo "4. Test service endpoints to verify configuration is working"
    echo ""
    echo -e "${YELLOW}💡 Tips:${NC}"
    echo "• Use 'sync: false' in render.yaml for sensitive variables"
    echo "• Generate strong random strings for JWT secrets"
    echo "• Use test/sandbox credentials for dev/stage environments"
    echo "• Use live/production credentials only for prod environment"
}

# Main script
if [ $# -lt 2 ]; then
    show_usage
    exit 1
fi

env=$1
service=$2

# Validate environment
case $env in
    dev|stage|prod)
        ;;
    *)
        echo -e "${RED}Error: Invalid environment: $env${NC}"
        show_usage
        exit 1
        ;;
esac

# Validate service
case $service in
    user|salon|booking|payment|notification|all)
        ;;
    *)
        echo -e "${RED}Error: Invalid service: $service${NC}"
        show_usage
        exit 1
        ;;
esac

extract_secrets $env $service
