#!/bin/bash

# JWT Secrets Copy Helper
# Usage: ./copy-secrets.sh [environment] [service]
# Helps copy environment-specific JWT secrets for Render setup

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
    echo "Services: user, salon, all"
    echo ""
    echo "Examples:"
    echo "  $0 dev user     # Show user-service secrets for dev"
    echo "  $0 prod salon   # Show salon-service secrets for prod"
    echo "  $0 stage all    # Show all secrets for stage"
}

extract_secrets() {
    local env=$1
    local service=$2
    local secrets_file="config/environments/${env}/secrets.env"
    
    if [ ! -f "$secrets_file" ]; then
        echo -e "${RED}Error: Secrets file not found: $secrets_file${NC}"
        echo "Run: ./deploy-multi-env.sh all $env --generate-secrets"
        exit 1
    fi
    
    echo -e "${BLUE}📋 JWT Secrets for ${env^^} Environment${NC}"
    echo "================================================"
    echo ""
    
    case $service in
        "user")
            echo -e "${GREEN}🔑 User Service Secrets (copy to user-service-${env} in Render):${NC}"
            echo ""
            grep "USER_SERVICE_JWT" "$secrets_file" | while read line; do
                echo "  $line"
            done
            ;;
        "salon")
            echo -e "${GREEN}🔑 Salon Service Secrets (copy to salon-service-${env} in Render):${NC}"
            echo ""
            grep "SALON_SERVICE_JWT" "$secrets_file" | while read line; do
                echo "  $line"
            done
            ;;
        "all")
            echo -e "${GREEN}🔑 User Service Secrets (copy to user-service-${env}):${NC}"
            echo ""
            grep "USER_SERVICE_JWT" "$secrets_file" | while read line; do
                echo "  $line"
            done
            echo ""
            echo -e "${GREEN}🔑 Salon Service Secrets (copy to salon-service-${env}):${NC}"
            echo ""
            grep "SALON_SERVICE_JWT" "$secrets_file" | while read line; do
                echo "  $line"
            done
            ;;
        *)
            echo -e "${RED}Error: Invalid service: $service${NC}"
            show_usage
            exit 1
            ;;
    esac
    
    echo ""
    echo -e "${YELLOW}📝 Additional Configuration (optional):${NC}"
    echo ""
    grep -E "(TTL|EXPIRY)" "$secrets_file" | while read line; do
        echo "  $line"
    done
    
    echo ""
    echo -e "${BLUE}🔗 Render Dashboard Links:${NC}"
    case $service in
        "user"|"all")
            echo "  User Service: https://dashboard.render.com/web/srv-[your-user-service-${env}-id]/env"
            ;;
    esac
    case $service in
        "salon"|"all")
            echo "  Salon Service: https://dashboard.render.com/web/srv-[your-salon-service-${env}-id]/env"
            ;;
    esac
    
    echo ""
    echo -e "${GREEN}✅ Next Steps:${NC}"
    echo "1. Copy the secrets above to your Render service environment variables"
    echo "2. Redeploy the service(s) to apply the new secrets"
    echo "3. Test authentication endpoints to verify secrets are working"
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
    user|salon|all)
        ;;
    *)
        echo -e "${RED}Error: Invalid service: $service${NC}"
        show_usage
        exit 1
        ;;
esac

extract_secrets $env $service
