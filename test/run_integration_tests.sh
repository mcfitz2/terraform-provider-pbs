#!/usr/bin/env bash

# Integration Test Runner for PBS Terraform Provider
# This script sets up and runs integration tests against a PBS instance

set -e

# Ensure Node.js is available in PATH for Terraform operations
export PATH="/usr/local/bin:/usr/bin:$PATH"
export NODE_PATH="/usr/local/bin:/usr/bin"

# Try to find node binary and add it to PATH if available
if command -v node >/dev/null 2>&1; then
    NODE_BIN=$(command -v node)
    NODE_DIR=$(dirname "$NODE_BIN")
    export PATH="$NODE_DIR:$PATH"
    echo "Found Node.js at: $NODE_BIN"
elif [ -x "/usr/local/bin/node" ]; then
    export PATH="/usr/local/bin:$PATH"
    echo "Using Node.js at: /usr/local/bin/node"
elif [ -x "/usr/bin/node" ]; then
    export PATH="/usr/bin:$PATH"
    echo "Using Node.js at: /usr/bin/node"
else
    echo "Warning: Node.js not found in PATH - Terraform JSON operations may fail"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERBOSE=false
COVERAGE=false
TIMEOUT="30m"
TEST_PATTERN=""

# Help function
show_help() {
    cat << EOF
PBS Terraform Provider Integration Test Runner

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    -v, --verbose       Enable verbose output
    -c, --coverage      Enable coverage reporting
    -t, --timeout       Test timeout (default: 30m)
    -p, --pattern       Run specific test pattern
    --setup-only        Only setup test environment, don't run tests
    --cleanup           Clean up test artifacts and exit

ENVIRONMENT VARIABLES (Required):
    PBS_ADDRESS         PBS server address (e.g., https://pbs.example.com:8007)
    PBS_USERNAME        PBS username
    PBS_PASSWORD        PBS password

ENVIRONMENT VARIABLES (Optional):
    PBS_INSECURE_TLS    Skip TLS verification (default: false)

MULTI-PROVIDER S3 TESTING VARIABLES (Optional):
    AWS_ACCESS_KEY_ID       AWS access key for S3 testing
    AWS_SECRET_ACCESS_KEY   AWS secret key for S3 testing
    AWS_REGION             AWS region (default: us-east-1)
    
    B2_ACCESS_KEY_ID       Backblaze B2 key ID for S3 testing
    B2_SECRET_ACCESS_KEY   Backblaze B2 application key for S3 testing
    B2_REGION              B2 region (default: us-west-004)
    
    SCALEWAY_ACCESS_KEY    Scaleway access key for S3 testing
    SCALEWAY_SECRET_KEY    Scaleway secret key for S3 testing
    SCALEWAY_REGION        Scaleway region (default: fr-par)

EXAMPLES:
    # Run all integration tests
    PBS_ADDRESS=https://pbs.local:8007 PBS_USERNAME=admin PBS_PASSWORD=secret $0

    # Run specific test with verbose output
    PBS_ADDRESS=https://pbs.local:8007 PBS_USERNAME=admin PBS_PASSWORD=secret $0 -v -p TestS3Endpoint

    # Run with coverage
    PBS_ADDRESS=https://pbs.local:8007 PBS_USERNAME=admin PBS_PASSWORD=secret $0 -c
    
    # Run multi-provider S3 tests with AWS and Backblaze
    PBS_ADDRESS=https://pbs.local:8007 PBS_USERNAME=admin PBS_PASSWORD=secret \\
    AWS_ACCESS_KEY_ID=your-key AWS_SECRET_ACCESS_KEY=your-secret \\
    B2_ACCESS_KEY_ID=your-b2-key B2_SECRET_ACCESS_KEY=your-b2-secret \\
    $0 -v -p TestS3EndpointMultiProvider

    # Cleanup test artifacts
    $0 --cleanup

EOF
}

# Logging functions
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

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -p|--pattern)
            TEST_PATTERN="$2"
            shift 2
            ;;
        --setup-only)
            SETUP_ONLY=true
            shift
            ;;
        --cleanup)
            CLEANUP_ONLY=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Cleanup function
cleanup_test_artifacts() {
    log_info "Cleaning up test artifacts..."
    
    # Remove terraform state files
    find . -name "terraform.tfstate*" -delete 2>/dev/null || true
    find . -name ".terraform" -type d -exec rm -rf {} + 2>/dev/null || true
    find . -name "*.tfplan" -delete 2>/dev/null || true
    
    # Remove coverage files
    rm -f coverage.out coverage.html 2>/dev/null || true
    
    # Remove test binary
    rm -f terraform-provider-pbs 2>/dev/null || true
    
    log_success "Test artifacts cleaned up"
}

# Handle cleanup-only mode
if [ "$CLEANUP_ONLY" = true ]; then
    cleanup_test_artifacts
    exit 0
fi

# Check required environment variables
check_env_vars() {
    log_info "Checking environment variables..."
    
    if [ -z "$PBS_ADDRESS" ]; then
        log_error "PBS_ADDRESS environment variable is required"
        echo "Example: export PBS_ADDRESS=https://pbs.local:8007"
        exit 1
    fi
    
    if [ -z "$PBS_USERNAME" ]; then
        log_error "PBS_USERNAME environment variable is required"
        echo "Example: export PBS_USERNAME=admin"
        exit 1
    fi
    
    if [ -z "$PBS_PASSWORD" ]; then
        log_error "PBS_PASSWORD environment variable is required"
        echo "Example: export PBS_PASSWORD=secret"
        exit 1
    fi
    
    log_success "Environment variables validated"
    log_info "PBS Address: $PBS_ADDRESS"
    log_info "PBS Username: $PBS_USERNAME"
    log_info "TLS Insecure: ${PBS_INSECURE_TLS:-false}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed or not in PATH"
        exit 1
    fi
    
    TF_VERSION=$(terraform version -json | grep -o '"terraform_version":"[^"]*' | cut -d'"' -f4)
    log_info "Terraform version: $TF_VERSION"
    
    log_success "Prerequisites validated"
}

# Setup test environment
setup_test_env() {
    log_info "Setting up test environment..."
    
    # Clean up any previous artifacts
    cleanup_test_artifacts
    
    # Update go dependencies
    log_info "Updating Go dependencies..."
    go mod tidy
    go mod download
    
    # Build the provider
    log_info "Building terraform provider..."
    go build -o terraform-provider-pbs .
    
    if [ ! -f "terraform-provider-pbs" ]; then
        log_error "Failed to build terraform provider"
        exit 1
    fi
    
    log_success "Test environment setup complete"
}

# Test PBS connectivity
test_pbs_connectivity() {
    log_info "Testing PBS connectivity..."
    
    # Create a simple Go program to test connectivity with authentication
    cat > /tmp/test_pbs_connection.go << 'EOF'
package main

import (
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "time"
)

type AuthResponse struct {
    Data struct {
        Ticket   string `json:"ticket"`
        CSRFToken string `json:"CSRFPreventionToken"`
    } `json:"data"`
}

func main() {
    address := os.Getenv("PBS_ADDRESS")
    username := os.Getenv("PBS_USERNAME")
    password := os.Getenv("PBS_PASSWORD")
    
    // Create HTTP client with timeout
    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: os.Getenv("PBS_INSECURE_TLS") == "true",
            },
        },
    }
    
    // First, try to get the version without auth (some PBS instances allow this)
    resp, err := client.Get(address + "/api2/json/version")
    if err != nil {
        fmt.Printf("Connection failed: %v\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 200 {
        fmt.Println("PBS connectivity test successful (no auth required)")
        return
    }
    
    // If version endpoint requires auth, try to authenticate
    fmt.Printf("Version endpoint requires authentication (status: %d), testing login...\n", resp.StatusCode)
    
    // Prepare login data
    loginData := url.Values{
        "username": {username},
        "password": {password},
    }
    
    // Attempt login
    loginResp, err := client.PostForm(address + "/api2/json/access/ticket", loginData)
    if err != nil {
        fmt.Printf("Login request failed: %v\n", err)
        os.Exit(1)
    }
    defer loginResp.Body.Close()
    
    if loginResp.StatusCode != 200 {
        body, _ := io.ReadAll(loginResp.Body)
        fmt.Printf("Login failed with status %d: %s\n", loginResp.StatusCode, string(body))
        os.Exit(1)
    }
    
    // Parse login response
    var authResp AuthResponse
    body, err := io.ReadAll(loginResp.Body)
    if err != nil {
        fmt.Printf("Failed to read login response: %v\n", err)
        os.Exit(1)
    }
    
    if err := json.Unmarshal(body, &authResp); err != nil {
        fmt.Printf("Failed to parse login response: %v\n", err)
        os.Exit(1)
    }
    
    if authResp.Data.Ticket == "" {
        fmt.Printf("Login successful but no ticket received\n")
        os.Exit(1)
    }
    
    fmt.Println("PBS authentication test successful")
}
EOF
    
    if go run /tmp/test_pbs_connection.go; then
        log_success "PBS connectivity verified"
    else
        log_error "Failed to connect to PBS instance"
        exit 1
    fi
    
    # Clean up
    rm -f /tmp/test_pbs_connection.go
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    # Build test command
    TEST_CMD="go test"
    
    if [ "$VERBOSE" = true ]; then
        TEST_CMD="$TEST_CMD -v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
    fi
    
    TEST_CMD="$TEST_CMD -timeout $TIMEOUT"
    
    if [ -n "$TEST_PATTERN" ]; then
        TEST_CMD="$TEST_CMD -run $TEST_PATTERN"
    fi
    
    TEST_CMD="$TEST_CMD ./test/..."
    
    log_info "Running: $TEST_CMD"
    
    # Run tests
    if eval $TEST_CMD; then
        log_success "Integration tests completed successfully"
        
        # Generate coverage report if enabled
        if [ "$COVERAGE" = true ] && [ -f "coverage.out" ]; then
            log_info "Generating coverage report..."
            go tool cover -html=coverage.out -o coverage.html
            COVERAGE_PERCENT=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            log_info "Test coverage: $COVERAGE_PERCENT"
            log_info "Coverage report: coverage.html"
        fi
        
        return 0
    else
        log_error "Integration tests failed"
        return 1
    fi
}

# Main execution
main() {
    echo "=================================================="
    echo "PBS Terraform Provider Integration Test Runner"
    echo "=================================================="
    
    check_env_vars
    check_prerequisites
    setup_test_env
    
    if [ "$SETUP_ONLY" = true ]; then
        log_success "Setup completed. Skipping test execution."
        exit 0
    fi
    
    test_pbs_connectivity
    
    if run_integration_tests; then
        log_success "All integration tests passed!"
        exit 0
    else
        log_error "Integration tests failed!"
        exit 1
    fi
}

# Run main function
main "$@"