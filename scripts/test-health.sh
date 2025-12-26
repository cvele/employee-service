#!/bin/bash

# Health Check Test Script
# Tests the /health/live and /health/ready endpoints

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8000}"
HEALTH_LIVE="$BASE_URL/health/live"
HEALTH_READY="$BASE_URL/health/ready"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Test liveness endpoint
test_liveness() {
    print_info "Testing liveness endpoint: $HEALTH_LIVE"
    
    response=$(curl -s -w "\n%{http_code}" "$HEALTH_LIVE")
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    if [ "$status" -eq 200 ]; then
        print_success "Liveness check passed (HTTP $status)"
        echo "  Response: $body"
    else
        print_error "Liveness check failed (HTTP $status)"
        echo "  Response: $body"
        return 1
    fi
}

# Test readiness endpoint
test_readiness() {
    print_info "Testing readiness endpoint: $HEALTH_READY"
    
    response=$(curl -s -w "\n%{http_code}" "$HEALTH_READY")
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    if [ "$status" -eq 200 ]; then
        print_success "Readiness check passed (HTTP $status)"
        echo "  Response: $body"
    elif [ "$status" -eq 503 ]; then
        print_error "Service not ready (HTTP $status)"
        echo "  Response: $body"
        echo "  This is expected if database or other dependencies are not available"
        return 1
    else
        print_error "Readiness check failed with unexpected status (HTTP $status)"
        echo "  Response: $body"
        return 1
    fi
}

# Test health endpoints with detailed output
test_health_verbose() {
    print_info "Testing liveness endpoint with details"
    curl -v "$HEALTH_LIVE" 2>&1 | grep -E "(< HTTP|< Content-Type|^OK|Service not live)"
    echo ""
    
    print_info "Testing readiness endpoint with details"
    curl -v "$HEALTH_READY" 2>&1 | grep -E "(< HTTP|< Content-Type|^OK|Service not ready)"
    echo ""
}

# Main execution
echo "=========================================="
echo "  Employee Service - Health Check Tests"
echo "=========================================="
echo ""

# Check if service is running
print_info "Checking if service is accessible at $BASE_URL"
if ! curl -s -f -o /dev/null "$BASE_URL/metrics" 2>/dev/null; then
    print_error "Service is not accessible. Make sure the service is running."
    echo ""
    echo "To start the service:"
    echo "  1. Start dependencies: make docker-up"
    echo "  2. Run the service: ./bin/employee-service -conf ./configs"
    exit 1
fi
print_success "Service is accessible"
echo ""

# Run tests
test_liveness
echo ""

test_readiness
echo ""

# Show verbose output if requested
if [ "$1" = "-v" ] || [ "$1" = "--verbose" ]; then
    echo "=========================================="
    echo "  Verbose Output"
    echo "=========================================="
    echo ""
    test_health_verbose
fi

echo "=========================================="
print_success "All health checks completed"
echo "=========================================="

