#!/bin/bash

# Configuration Management Script for Salon Platform
# Usage: ./config-manager.sh [command] [environment] [service]

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

CONFIG_DIR="config/environments"

show_usage() {
    echo "Configuration Manager for Salon Platform"
    echo ""
    echo "Usage: $0 [command] [environment] [service]"
    echo ""
    echo "Commands:"
    echo "  list                    List all available configurations"
    echo "  show [env] [service]    Show configuration for specific environment and service"
    echo "  validate [env]          Validate all configurations for environment"
    echo "  copy [from-env] [to-env] Copy configuration from one environment to another"
    echo "  diff [env1] [env2]      Compare configurations between environments"
    echo ""
    echo "Environments: dev, stage, prod"
    echo "Services: user-service, salon-service, render, secrets"
    echo ""
    echo "Examples:"
    echo "  $0 list                           # List all configurations"
    echo "  $0 show dev user-service          # Show dev user-service config"
    echo "  $0 validate prod                  # Validate all prod configs"
    echo "  $0 copy dev stage                 # Copy dev configs to stage"
    echo "  $0 diff dev prod                  # Compare dev vs prod configs"
}

list_configs() {
    echo -e "${BLUE}📋 Available Configurations:${NC}"
    echo ""
    
    for env in dev stage prod; do
        if [ -d "$CONFIG_DIR/$env" ]; then
            echo -e "${GREEN}🔧 $env Environment:${NC}"
            for config in "$CONFIG_DIR/$env"/*.yaml; do
                if [ -f "$config" ]; then
                    basename=$(basename "$config" .yaml)
                    echo "  • $basename.yaml"
                fi
            done
            # Check for secrets file
            if [ -f "$CONFIG_DIR/$env/secrets.env" ]; then
                echo "  • secrets.env"
            fi
            echo ""
        fi
    done
}

show_config() {
    local env=$1
    local service=$2
    
    if [ -z "$env" ] || [ -z "$service" ]; then
        echo -e "${RED}Error: Environment and service required${NC}"
        show_usage
        exit 1
    fi
    
    local config_file="$CONFIG_DIR/$env/$service.yaml"
    
    # Handle special case for secrets
    if [ "$service" = "secrets" ]; then
        config_file="$CONFIG_DIR/$env/secrets.env"
        if [ ! -f "$config_file" ]; then
            echo -e "${RED}Error: Secrets file not found: $config_file${NC}"
            echo "Run: ./deploy-multi-env.sh all $env --generate-secrets"
            exit 1
        fi
        echo -e "${BLUE}📄 Secrets: $env/secrets.env${NC}"
        echo "================================================"
        cat "$config_file"
        return
    fi
    
    if [ ! -f "$config_file" ]; then
        echo -e "${RED}Error: Configuration not found: $config_file${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}📄 Configuration: $env/$service.yaml${NC}"
    echo "================================================"
    cat "$config_file"
}

validate_config() {
    local env=$1
    
    if [ -z "$env" ]; then
        echo -e "${RED}Error: Environment required${NC}"
        show_usage
        exit 1
    fi
    
    if [ ! -d "$CONFIG_DIR/$env" ]; then
        echo -e "${RED}Error: Environment directory not found: $CONFIG_DIR/$env${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}🔍 Validating $env environment configurations...${NC}"
    echo ""
    
    local errors=0
    
    # Check required files
    for service in user-service salon-service render; do
        local config_file="$CONFIG_DIR/$env/$service.yaml"
        if [ -f "$config_file" ]; then
            echo -e "${GREEN}✅ $service.yaml${NC} - Found"
            
            # Basic YAML syntax check
            if command -v python3 &> /dev/null; then
                if python3 -c "import yaml; yaml.safe_load(open('$config_file'))" 2>/dev/null; then
                    echo -e "   ${GREEN}✅ Valid YAML syntax${NC}"
                else
                    echo -e "   ${RED}❌ Invalid YAML syntax${NC}"
                    errors=$((errors + 1))
                fi
            fi
        else
            echo -e "${RED}❌ $service.yaml${NC} - Missing"
            errors=$((errors + 1))
        fi
    done
    
    echo ""
    if [ $errors -eq 0 ]; then
        echo -e "${GREEN}✅ All configurations valid for $env environment${NC}"
    else
        echo -e "${RED}❌ Found $errors error(s) in $env environment${NC}"
        exit 1
    fi
}

copy_configs() {
    local from_env=$1
    local to_env=$2
    
    if [ -z "$from_env" ] || [ -z "$to_env" ]; then
        echo -e "${RED}Error: Source and target environments required${NC}"
        show_usage
        exit 1
    fi
    
    if [ ! -d "$CONFIG_DIR/$from_env" ]; then
        echo -e "${RED}Error: Source environment not found: $from_env${NC}"
        exit 1
    fi
    
    if [ ! -d "$CONFIG_DIR/$to_env" ]; then
        echo -e "${YELLOW}Creating target environment directory: $CONFIG_DIR/$to_env${NC}"
        mkdir -p "$CONFIG_DIR/$to_env"
    fi
    
    echo -e "${BLUE}📋 Copying configurations from $from_env to $to_env...${NC}"
    echo ""
    
    for config in "$CONFIG_DIR/$from_env"/*.yaml; do
        if [ -f "$config" ]; then
            local basename=$(basename "$config")
            local target="$CONFIG_DIR/$to_env/$basename"
            
            cp "$config" "$target"
            echo -e "${GREEN}✅ Copied $basename${NC}"
            
            # Update environment-specific values
            sed -i "s/env: $from_env/env: $to_env/g" "$target"
            sed -i "s/$from_env-/$to_env-/g" "$target"
            sed -i "s/_$from_env/_$to_env/g" "$target"
        fi
    done
    
    echo ""
    echo -e "${GREEN}✅ Configuration copy completed${NC}"
    echo -e "${YELLOW}⚠️  Remember to update environment-specific values manually${NC}"
}

diff_configs() {
    local env1=$1
    local env2=$2
    
    if [ -z "$env1" ] || [ -z "$env2" ]; then
        echo -e "${RED}Error: Two environments required for comparison${NC}"
        show_usage
        exit 1
    fi
    
    if [ ! -d "$CONFIG_DIR/$env1" ] || [ ! -d "$CONFIG_DIR/$env2" ]; then
        echo -e "${RED}Error: One or both environments not found${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}🔍 Comparing configurations: $env1 vs $env2${NC}"
    echo "================================================"
    
    for service in user-service salon-service render; do
        local config1="$CONFIG_DIR/$env1/$service.yaml"
        local config2="$CONFIG_DIR/$env2/$service.yaml"
        
        if [ -f "$config1" ] && [ -f "$config2" ]; then
            echo ""
            echo -e "${GREEN}📄 $service.yaml${NC}"
            echo "----------------------------------------"
            if command -v diff &> /dev/null; then
                diff -u "$config1" "$config2" || true
            else
                echo "diff command not available"
            fi
        elif [ -f "$config1" ]; then
            echo -e "${YELLOW}⚠️  $service.yaml exists only in $env1${NC}"
        elif [ -f "$config2" ]; then
            echo -e "${YELLOW}⚠️  $service.yaml exists only in $env2${NC}"
        else
            echo -e "${RED}❌ $service.yaml missing in both environments${NC}"
        fi
    done
}

# Main script
case "${1:-}" in
    "list")
        list_configs
        ;;
    "show")
        show_config "$2" "$3"
        ;;
    "validate")
        validate_config "$2"
        ;;
    "copy")
        copy_configs "$2" "$3"
        ;;
    "diff")
        diff_configs "$2" "$3"
        ;;
    *)
        show_usage
        exit 1
        ;;
esac
