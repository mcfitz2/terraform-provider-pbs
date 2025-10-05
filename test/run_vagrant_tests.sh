#!/usr/bin/env bash

# Vagrant-based PBS VM Test Runner
# This script manages PBS VM lifecycle using Vagrant and runs integration tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

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
        log_info "Cleaning up PBS VM..."
        vagrant destroy -f >/dev/null 2>&1 || true
        rm -f zfs-disk.vdi 2>/dev/null || true
    fi
}

# Set up cleanup trap
CLEANUP_ON_EXIT="true"
trap cleanup EXIT

show_help() {
    cat << EOF
Vagrant-based PBS VM Test Runner

USAGE:
    $0 [OPTIONS] [TEST_PATTERN]

OPTIONS:
    -h, --help          Show this help message
    -k, --keep          Keep the PBS VM running after tests
    -v, --verbose       Enable verbose test output
    -s, --start-only    Only start the PBS VM, don't run tests
    --provision         Force re-provisioning of the VM
    --cleanup           Destroy existing PBS VM and exit
    --status            Show PBS VM status
    --ssh               SSH into the PBS VM
    --logs              Show PBS service logs from VM

EXAMPLES:
    $0                              # Run all integration tests
    $0 -v TestDatastoreZFS          # Run ZFS tests with verbose output
    $0 -k TestIntegration           # Run integration tests and keep VM
    $0 --start-only                 # Only start PBS VM
    $0 --ssh                        # SSH into running PBS VM
    $0 --cleanup                    # Destroy VM and cleanup

ENVIRONMENT VARIABLES:
    # Optional S3 provider credentials for S3 tests:
    AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION
    B2_ACCESS_KEY_ID, B2_SECRET_ACCESS_KEY, B2_REGION  
    SCALEWAY_ACCESS_KEY, SCALEWAY_SECRET_KEY, SCALEWAY_REGION

NOTES:
    - Requires Vagrant and VirtualBox to be installed
    - VM provisioning takes 5-10 minutes on first run
    - Subsequent runs use the cached VM unless --provision is used
    - ZFS tests require the VM (Docker container doesn't support ZFS)

EOF
}

check_requirements() {
    local missing_deps=()
    
    if ! command -v vagrant >/dev/null 2>&1; then
        missing_deps+=("vagrant")
    fi
    
    if ! command -v VBoxManage >/dev/null 2>&1; then
        missing_deps+=("virtualbox")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        echo
        echo "Installation instructions:"
        echo "  macOS:   brew install vagrant virtualbox"
        echo "  Linux:   Install from https://www.vagrantup.com/ and https://www.virtualbox.org/"
        echo
        exit 1
    fi
}

# Parse command line arguments
VERBOSE=""
KEEP_VM=""
START_ONLY=""
FORCE_PROVISION=""
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
            KEEP_VM="true"
            CLEANUP_ON_EXIT=""
            shift
            ;;
        -s|--start-only)
            START_ONLY="true"
            CLEANUP_ON_EXIT=""
            shift
            ;;
        --provision)
            FORCE_PROVISION="--provision"
            shift
            ;;
        --cleanup)
            log_info "Cleaning up existing PBS VM..."
            vagrant destroy -f
            rm -f zfs-disk.vdi
            log_success "Cleanup completed"
            exit 0
            ;;
        --status)
            vagrant status
            exit 0
            ;;
        --ssh)
            vagrant ssh
            exit 0
            ;;
        --logs)
            log_info "Fetching PBS service logs..."
            vagrant ssh -c "sudo journalctl -u proxmox-backup.service -n 50 --no-pager"
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

# Check requirements
check_requirements

# Check if VM is already running
VM_STATUS=$(vagrant status --machine-readable | grep ",state," | cut -d',' -f4)

if [[ "$VM_STATUS" == "running" && -z "$FORCE_PROVISION" ]]; then
    log_info "PBS VM is already running"
else
    # Start or create PBS VM
    log_info "Starting PBS VM (this may take 5-10 minutes on first run)..."
    
    if [[ -n "$FORCE_PROVISION" ]]; then
        vagrant up $FORCE_PROVISION
    else
        vagrant up
    fi
    
    # Wait a bit for services to stabilize
    log_info "Waiting for PBS services to stabilize..."
    sleep 10
fi

# Verify PBS is accessible
log_info "Verifying PBS API accessibility..."
timeout=60
elapsed=0
while [[ $elapsed -lt $timeout ]]; do
    if curl -k -f -s https://localhost:8007 >/dev/null 2>&1; then
        log_success "PBS API is accessible!"
        break
    fi
    
    if [[ $elapsed -ge $timeout ]]; then
        log_error "PBS API is not accessible after ${timeout}s"
        log_info "Checking VM status..."
        vagrant status
        log_info "Checking PBS service status..."
        vagrant ssh -c "sudo systemctl status proxmox-backup.service --no-pager" || true
        exit 1
    fi
    
    sleep 2
    elapsed=$((elapsed + 2))
done

# Display VM information
log_info "=== PBS VM Information ==="
vagrant ssh -c "sudo proxmox-backup-manager version --verbose" 2>/dev/null | head -5
echo ""
log_info "ZFS Pools:"
vagrant ssh -c "sudo zpool list 2>/dev/null" || echo "  No ZFS pools configured"
echo ""
log_info "Connection Details:"
echo "  URL:      https://localhost:8007"
echo "  Username: admin@pbs"
echo "  Password: password123"
echo "========================="

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

if [[ ${#S3_PROVIDERS_AVAILABLE[@]} -gt 0 ]]; then
    log_info "S3 providers configured: ${S3_PROVIDERS_AVAILABLE[*]}"
else
    log_warning "No S3 provider credentials found. S3 tests will be skipped."
fi

# If start-only mode, show info and exit
if [[ "$START_ONLY" == "true" ]]; then
    log_success "PBS VM is running and ready!"
    echo
    echo "To run tests manually:"
    echo "  export PBS_ADDRESS=https://localhost:8007"
    echo "  export PBS_USERNAME=admin@pbs"
    echo "  export PBS_PASSWORD=password123"
    echo "  export PBS_INSECURE_TLS=true"
    echo "  cd .."
    echo "  go test $VERBOSE ./test/integration -run \"\${TEST_PATTERN:-.*}\""
    echo
    echo "Useful commands:"
    echo "  $0 --ssh              # SSH into VM"
    echo "  $0 --logs             # View PBS logs"
    echo "  $0 --cleanup          # Destroy VM"
    exit 0
fi

# Set up environment for tests
export PBS_ADDRESS="https://localhost:8007"
export PBS_USERNAME="admin@pbs"
export PBS_PASSWORD="password123"
export PBS_INSECURE_TLS="true"

# Run tests
log_info "Running integration tests..."
cd "$SCRIPT_DIR/.."

if [[ -n "$TEST_PATTERN" ]]; then
    log_info "Running tests matching pattern: $TEST_PATTERN"
    go test $VERBOSE ./test/integration -timeout 30m -run "$TEST_PATTERN"
else
    log_info "Running all integration tests (including ZFS)"
    go test $VERBOSE ./test/integration -timeout 45m
fi

test_exit_code=$?

if [[ $test_exit_code -eq 0 ]]; then
    log_success "All tests passed!"
else
    log_error "Some tests failed (exit code: $test_exit_code)"
    log_info "To view PBS logs: $0 --logs"
fi

# Keep VM running if requested
if [[ "$KEEP_VM" == "true" ]]; then
    log_info "Keeping PBS VM running"
    log_info "  Use '$0 --ssh' to access VM"
    log_info "  Use '$0 --cleanup' to destroy VM"
    CLEANUP_ON_EXIT=""
fi

exit $test_exit_code
