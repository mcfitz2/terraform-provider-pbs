#!/usr/bin/env bash

# Local Docker PBS Test Runner
# This script starts a PBS container and runs integration tests against it

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

cleanup() {
    if [[ "$CLEANUP_ON_EXIT" == "true" ]]; then
        log_info "Cleaning up PBS container..."
        docker-compose -f "$COMPOSE_FILE" down --volumes --remove-orphans >/dev/null 2>&1 || true
    fi
}

# Set up cleanup trap
CLEANUP_ON_EXIT="true"
trap cleanup EXIT

show_help() {
    cat << EOF
Local Docker PBS Test Runner

USAGE:
    $0 [OPTIONS] [TEST_PATTERN]

OPTIONS:
    -h, --help          Show this help message
    -k, --keep          Keep the PBS container running after tests
    -v, --verbose       Enable verbose test output
    -s, --start-only    Only start the PBS container, don't run tests
    --cleanup           Stop and remove existing PBS container
    --logs              Show PBS container logs

EXAMPLES:
    $0                              # Run all integration tests
    $0 -v TestQuickSmoke            # Run smoke tests with verbose output
    $0 -k TestS3EndpointMultiProvider  # Run S3 tests and keep container
    $0 --start-only                 # Only start PBS container
    $0 --cleanup                    # Clean up containers and exit

ENVIRONMENT VARIABLES:
    PBS_IMAGE           Docker image to use (default: ayufan/proxmox-backup-server:latest)
    PBS_TAG             Docker tag to use (default: latest)
    
    # Optional S3 provider credentials for S3 tests:
    AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION
    B2_ACCESS_KEY_ID, B2_SECRET_ACCESS_KEY, B2_REGION  
    SCALEWAY_ACCESS_KEY, SCALEWAY_SECRET_KEY, SCALEWAY_REGION

EOF
}

# Parse command line arguments
VERBOSE=""
KEEP_CONTAINER=""
START_ONLY=""
TEST_PATTERN=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -k|--keep)
            KEEP_CONTAINER="true"
            CLEANUP_ON_EXIT=""
            shift
            ;;
        -s|--start-only)
            START_ONLY="true"
            CLEANUP_ON_EXIT=""
            shift
            ;;
        --cleanup)
            log_info "Cleaning up existing PBS containers..."
            docker-compose -f "$COMPOSE_FILE" down --volumes --remove-orphans
            log_success "Cleanup completed"
            exit 0
            ;;
        --logs)
            docker-compose -f "$COMPOSE_FILE" logs -f pbs
            exit 0
            ;;
        *)
            if [[ -z "$TEST_PATTERN" ]]; then
                TEST_PATTERN="$1"
            else
                log_error "Unknown option: $1"
                exit 1
            fi
            shift
            ;;
    esac
done

# Check if Docker is available
if ! command -v docker >/dev/null 2>&1; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose >/dev/null 2>&1; then
    log_error "Docker Compose is not installed or not in PATH"
    exit 1
fi

# Start PBS container
log_info "Starting PBS container..."
docker-compose -f "$COMPOSE_FILE" up -d

# Wait for PBS to be ready
log_info "Waiting for PBS to be ready..."
timeout=300  # 5 minutes
elapsed=0
while [[ $elapsed -lt $timeout ]]; do
    if curl -k -f -s https://localhost:8007 >/dev/null 2>&1; then
        log_success "PBS is ready!"
        break
    fi
    
    if [[ $((elapsed % 30)) -eq 0 ]]; then
        log_info "Still waiting for PBS... ($elapsed/${timeout}s)"
    fi
    
    sleep 5
    elapsed=$((elapsed + 5))
done

if [[ $elapsed -ge $timeout ]]; then
    log_error "PBS failed to start within $timeout seconds"
    log_info "Container logs:"
    docker-compose -f "$COMPOSE_FILE" logs --tail=50 pbs
    exit 1
fi

# Check S3 credentials
S3_PROVIDERS_AVAILABLE=()
if [[ -n "$AWS_ACCESS_KEY_ID" && -n "$AWS_SECRET_ACCESS_KEY" ]]; then
    S3_PROVIDERS_AVAILABLE+=("AWS")
fi
if [[ -n "$B2_ACCESS_KEY_ID" && -n "$B2_SECRET_ACCESS_KEY" ]]; then
    S3_PROVIDERS_AVAILABLE+=("Backblaze B2")
fi
if [[ -n "$SCALEWAY_ACCESS_KEY" && -n "$SCALEWAY_SECRET_KEY" ]]; then
    S3_PROVIDERS_AVAILABLE+=("Scaleway")
fi

# If start-only mode, show connection info and exit
if [[ "$START_ONLY" == "true" ]]; then
    log_success "PBS container is running!"
    echo
    echo "Connection details:"
    echo "  URL:      https://localhost:8007"
    echo "  Username: admin@pbs"  
    echo "  Password: password123"
    echo
    if [[ ${#S3_PROVIDERS_AVAILABLE[@]} -gt 0 ]]; then
        echo "S3 providers configured: ${S3_PROVIDERS_AVAILABLE[*]}"
    else
        echo "Note: No S3 provider credentials found. S3 tests will be skipped."
        echo "      To enable S3 tests, set AWS/B2/Scaleway credentials."
    fi
    echo
    echo "To run tests manually:"
    echo "  export PBS_ADDRESS=https://localhost:8007"
    echo "  export PBS_USERNAME=admin@pbs"
    echo "  export PBS_PASSWORD=password123"
    echo "  export PBS_INSECURE_TLS=true"
    echo "  go test $VERBOSE ./test -run \"\${TEST_PATTERN:-.*}\""
    echo
    echo "To stop the container:"
    echo "  $0 --cleanup"
    echo "  # or"
    echo "  docker-compose -f $COMPOSE_FILE down"
    exit 0
fi

# Set up environment for tests
export PBS_ADDRESS="https://localhost:8007"
export PBS_USERNAME="admin@pbs"
export PBS_PASSWORD="password123"
export PBS_INSECURE_TLS="true"

# Show S3 provider status
if [[ ${#S3_PROVIDERS_AVAILABLE[@]} -gt 0 ]]; then
    log_info "S3 providers configured: ${S3_PROVIDERS_AVAILABLE[*]}"
else
    log_warning "No S3 provider credentials found. S3 tests will be skipped."
    log_info "To enable S3 tests, set environment variables:"
    log_info "  AWS: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION"
    log_info "  B2: B2_ACCESS_KEY_ID, B2_SECRET_ACCESS_KEY, B2_REGION"
    log_info "  Scaleway: SCALEWAY_ACCESS_KEY, SCALEWAY_SECRET_KEY, SCALEWAY_REGION"
fi

# Run tests
log_info "Running integration tests..."
cd "$SCRIPT_DIR/.."

if [[ -n "$TEST_PATTERN" ]]; then
    log_info "Running tests matching pattern: $TEST_PATTERN"
    go test $VERBOSE ./test -run "$TEST_PATTERN"
else
    log_info "Running all integration tests"
    go test $VERBOSE ./test
fi

test_exit_code=$?

if [[ $test_exit_code -eq 0 ]]; then
    log_success "All tests passed!"
else
    log_error "Some tests failed (exit code: $test_exit_code)"
fi

# Keep container running if requested
if [[ "$KEEP_CONTAINER" == "true" ]]; then
    log_info "Keeping PBS container running (use '$0 --cleanup' to stop)"
    CLEANUP_ON_EXIT=""
fi

exit $test_exit_code