#!/bin/bash
set -e

# Script to run all Terraform HCL tests locally
# Requires: terraform 1.6+, built provider binary, PBS server access

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==========================================="
echo "Terraform HCL Test Runner"
echo "==========================================="
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check Terraform version
if ! command -v terraform &> /dev/null; then
    echo "❌ Error: terraform not found in PATH"
    exit 1
fi

TF_VERSION=$(terraform version -json | jq -r '.terraform_version')
echo "✅ Terraform version: $TF_VERSION"

# Check if provider binary exists
if [ ! -f "$PROJECT_ROOT/terraform-provider-pbs" ]; then
    echo "❌ Error: Provider binary not found. Run 'make build' first."
    exit 1
fi
echo "✅ Provider binary found"

# Check required environment variables
REQUIRED_VARS=("TF_VAR_pbs_endpoint" "TF_VAR_pbs_username" "TF_VAR_pbs_password")
MISSING_VARS=()

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        MISSING_VARS+=("$var")
    fi
done

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo "❌ Error: Missing required environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        echo "  - $var"
    done
    echo ""
    echo "Example:"
    echo "  export TF_VAR_pbs_endpoint='https://192.168.1.108:8007'"
    echo "  export TF_VAR_pbs_username='root@pam'"
    echo "  export TF_VAR_pbs_password='your-password'"
    exit 1
fi

echo "✅ Environment variables set"
echo ""

# Test directories in order
TEST_DIRS=(
    "datastores_datasource"
    "prune_job_datasource"
    "prune_jobs_datasource"
    "sync_job_datasource"
    "datastores"
    "jobs"
    "remotes"
    "metrics"
    "notifications"
    "datasources"
)

PASSED=0
FAILED=0
SKIPPED=0

echo "Running tests..."
echo ""

for dir in "${TEST_DIRS[@]}"; do
    TEST_PATH="$PROJECT_ROOT/test/tftest/$dir"
    
    if [ ! -d "$TEST_PATH" ]; then
        echo "⚠️  $dir - Directory not found (skipping)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi
    
    echo "→ Testing $dir..."
    
    if (cd "$TEST_PATH" && terraform init -input=false > /dev/null 2>&1 && terraform test); then
        echo "✅ $dir - PASSED"
        PASSED=$((PASSED + 1))
    else
        echo "❌ $dir - FAILED"
        FAILED=$((FAILED + 1))
    fi
    
    echo ""
done

# Summary
echo "==========================================="
echo "Test Summary"
echo "==========================================="
echo "Passed:  $PASSED"
echo "Failed:  $FAILED"
echo "Skipped: $SKIPPED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi
