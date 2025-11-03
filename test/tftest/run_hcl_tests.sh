#!/bin/bash
# run_hcl_tests.sh - Helper script to run Terraform HCL tests locally
#
# This script:
# 1. Builds the provider binary
# 2. Installs it to the local Terraform plugin directory
# 3. Runs the HCL tests
#
# Usage:
#   ./test/tftest/run_hcl_tests.sh [test_name]
#
# Examples:
#   ./test/tftest/run_hcl_tests.sh                    # Run all tests
#   ./test/tftest/run_hcl_tests.sh prune_job_datasource  # Run specific test

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Terraform HCL Test Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}❌ Terraform not found. Please install Terraform 1.6.0 or later.${NC}"
    exit 1
fi

# Check Terraform version
TF_VERSION=$(terraform version -json | grep -o '"terraform_version":"[^"]*"' | cut -d'"' -f4)
echo -e "${BLUE}→ Terraform version: ${TF_VERSION}${NC}"

# Verify version is 1.6.0 or later
REQUIRED_VERSION="1.6.0"
if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$TF_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo -e "${RED}❌ Terraform version must be 1.6.0 or later for HCL tests${NC}"
    exit 1
fi

# Check environment variables
if [ -z "$TF_VAR_pbs_endpoint" ] && [ ! -f "$SCRIPT_DIR/terraform.tfvars" ]; then
    echo -e "${YELLOW}⚠️  Warning: TF_VAR_pbs_endpoint not set and terraform.tfvars not found${NC}"
    echo -e "${YELLOW}   Please set environment variables or create terraform.tfvars files:${NC}"
    echo ""
    echo "   export TF_VAR_pbs_endpoint=\"https://192.168.1.108:8007\""
    echo "   export TF_VAR_pbs_username=\"root@pam\""
    echo "   export TF_VAR_pbs_password=\"your-password\""
    echo ""
fi

# Build provider
echo ""
echo -e "${BLUE}→ Building provider binary...${NC}"
cd "$PROJECT_ROOT"
go build -o terraform-provider-pbs .
echo -e "${GREEN}✓ Provider built${NC}"

# Install provider to plugin directory
echo ""
echo -e "${BLUE}→ Installing provider to plugin directory...${NC}"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
esac

# Map OS names
case "$OS" in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        echo -e "${RED}❌ Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/micah/pbs/1.0.0/${OS}_${ARCH}"
mkdir -p "$PLUGIN_DIR"
cp "$PROJECT_ROOT/terraform-provider-pbs" "$PLUGIN_DIR/"
chmod +x "$PLUGIN_DIR/terraform-provider-pbs"
echo -e "${GREEN}✓ Provider installed to $PLUGIN_DIR${NC}"

# Determine which tests to run
TEST_NAME="${1:-}"
if [ -z "$TEST_NAME" ]; then
    # Run all tests
    TESTS=(datastores_datasource prune_job_datasource prune_jobs_datasource sync_job_datasource)
else
    # Run specific test
    TESTS=("$TEST_NAME")
fi

# Run tests
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Running Tests${NC}"
echo -e "${BLUE}========================================${NC}"

FAILED_TESTS=()
for test in "${TESTS[@]}"; do
    TEST_DIR="$SCRIPT_DIR/$test"
    
    if [ ! -d "$TEST_DIR" ]; then
        echo -e "${RED}❌ Test directory not found: $TEST_DIR${NC}"
        FAILED_TESTS+=("$test")
        continue
    fi
    
    echo ""
    echo -e "${BLUE}→ Running test: ${test}${NC}"
    echo ""
    
    if terraform test -chdir="$TEST_DIR"; then
        echo -e "${GREEN}✓ Test passed: ${test}${NC}"
    else
        echo -e "${RED}❌ Test failed: ${test}${NC}"
        FAILED_TESTS+=("$test")
    fi
done

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ ${#FAILED_TESTS[@]} test(s) failed:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "${RED}   - $test${NC}"
    done
    exit 1
fi
