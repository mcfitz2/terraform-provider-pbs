#!/bin/bash
set -e

# Integration test runner with Docker services
# This script starts all required services and runs the integration tests

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Service configuration
INFLUXDB_PORT=8086
NFS_PORT=2049
CIFS_PORT=445
GOTIFY_PORT=8080
WEBHOOK_PORT=8081

# PBS configuration (from environment or defaults)
PBS_ADDRESS="${PBS_ADDRESS:-https://192.168.1.108:8007}"
PBS_USERNAME="${PBS_USERNAME:-root@pam}"
PBS_PASSWORD="${PBS_PASSWORD:-pbspbs123}"

# Get the PBS server's hostname/IP for calculating which local IP to use
PBS_HOST=$(echo "$PBS_ADDRESS" | sed -E 's#https?://([^:/]+).*#\1#')

# Function to get the local IP on the same subnet as PBS server
get_local_ip() {
    local pbs_ip=$1
    
    # If TEST_HOST_IP is already set, use it
    if [ -n "$TEST_HOST_IP" ]; then
        echo "$TEST_HOST_IP"
        return
    fi
    
    # Try to resolve PBS hostname to IP if it's a hostname
    if ! [[ $pbs_ip =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        pbs_ip=$(getent hosts "$pbs_ip" | awk '{ print $1 ; exit }' || echo "")
    fi
    
    if [ -z "$pbs_ip" ]; then
        echo -e "${YELLOW}Warning: Could not resolve PBS host, using 127.0.0.1${NC}" >&2
        echo "127.0.0.1"
        return
    fi
    
    # Get the network prefix (first 3 octets)
    local pbs_subnet=$(echo "$pbs_ip" | cut -d. -f1-3)
    
    # Find a local IP on the same subnet
    local local_ip=$(ip addr show 2>/dev/null | grep -oP "inet \K${pbs_subnet}\.\d+" | head -1)
    
    # macOS fallback
    if [ -z "$local_ip" ]; then
        local_ip=$(ifconfig 2>/dev/null | grep -oE "inet ${pbs_subnet}\.[0-9]+" | awk '{print $2}' | head -1)
    fi
    
    # If still not found, try to find any non-loopback IP
    if [ -z "$local_ip" ]; then
        local_ip=$(ip addr show 2>/dev/null | grep -oP 'inet \K[0-9.]+' | grep -v '^127\.' | head -1)
    fi
    
    # macOS fallback for any IP
    if [ -z "$local_ip" ]; then
        local_ip=$(ifconfig 2>/dev/null | grep -oE 'inet [0-9.]+' | awk '{print $2}' | grep -v '^127\.' | head -1)
    fi
    
    # Final fallback
    if [ -z "$local_ip" ]; then
        echo -e "${YELLOW}Warning: Could not determine local IP, using 127.0.0.1${NC}" >&2
        echo "127.0.0.1"
    else
        echo "$local_ip"
    fi
}

LOCAL_IP=$(get_local_ip "$PBS_HOST")

echo -e "${GREEN}Starting integration test infrastructure...${NC}"

# Function to check if a service is ready
wait_for_service() {
    local service=$1
    local host=$2
    local port=$3
    local max_attempts=30
    local attempt=1

    echo -e "${YELLOW}Waiting for ${service} to be ready...${NC}"
    while ! nc -z ${host} ${port} 2>/dev/null; do
        if [ $attempt -ge $max_attempts ]; then
            echo -e "${RED}${service} failed to start after ${max_attempts} attempts${NC}"
            return 1
        fi
        echo "  Attempt $attempt/$max_attempts..."
        sleep 2
        attempt=$((attempt + 1))
    done
    echo -e "${GREEN}${service} is ready!${NC}"
}

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up Docker containers...${NC}"
    docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" down -v
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

# Start Docker services
echo -e "${GREEN}Starting Docker services...${NC}"
docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" up -d

# Wait for services to be ready
wait_for_service "InfluxDB" "localhost" ${INFLUXDB_PORT}
wait_for_service "NFS" "localhost" ${NFS_PORT}
wait_for_service "CIFS" "localhost" ${CIFS_PORT}

# Give services extra time to fully initialize
echo -e "${YELLOW}Waiting for services to fully initialize...${NC}"
sleep 5

# Set environment variables for tests
export PBS_ADDRESS="${PBS_ADDRESS}"
export PBS_USERNAME="${PBS_USERNAME}"
export PBS_PASSWORD="${PBS_PASSWORD}"
export TF_ACC=1

# Service endpoints for tests (PBS needs to reach these, so use local IP not localhost)
export TEST_HOST_IP="${LOCAL_IP}"
export TEST_INFLUXDB_HOST="${LOCAL_IP}"
export TEST_INFLUXDB_PORT="${INFLUXDB_PORT}"
export TEST_INFLUXDB_ORG="testorg"
export TEST_INFLUXDB_BUCKET="pbs-metrics"
export TEST_INFLUXDB_TOKEN="test-token-for-pbs-provider"
export TEST_NFS_HOST="${LOCAL_IP}"
export TEST_NFS_EXPORT="/exports/test"
export TEST_CIFS_HOST="${LOCAL_IP}"
export TEST_CIFS_SHARE="testshare"
export TEST_CIFS_USERNAME="testuser"
export TEST_CIFS_PASSWORD="testpass"
export TEST_GOTIFY_HOST="${LOCAL_IP}"
export TEST_GOTIFY_PORT="${GOTIFY_PORT}"
export TEST_WEBHOOK_HOST="${LOCAL_IP}"
export TEST_WEBHOOK_PORT="${WEBHOOK_PORT}"

# Check for AWS credentials (for S3 tests)
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
    echo -e "${GREEN}AWS credentials detected - S3 tests will run${NC}"
    export AWS_S3_AVAILABLE=true
else
    echo -e "${YELLOW}AWS credentials not found - S3 tests will be skipped${NC}"
    export AWS_S3_AVAILABLE=false
fi

echo -e "${GREEN}Environment configured:${NC}"
echo -e "  ${BLUE}PBS Server:${NC} ${PBS_ADDRESS} (${PBS_HOST})"
echo -e "  ${BLUE}Local IP:${NC} ${LOCAL_IP} (for PBS to reach test services)"
echo -e "  ${BLUE}Test Services:${NC}"
echo "    - InfluxDB: ${TEST_INFLUXDB_HOST}:${TEST_INFLUXDB_PORT}"
echo "    - Gotify: ${TEST_GOTIFY_HOST}:${TEST_GOTIFY_PORT}"
echo "    - Webhook: ${TEST_WEBHOOK_HOST}:${TEST_WEBHOOK_PORT}"
echo "    - NFS: ${TEST_NFS_HOST}:${TEST_NFS_EXPORT}"
echo "    - CIFS: ${TEST_CIFS_HOST}/${TEST_CIFS_SHARE}"

# Run tests
echo -e "${GREEN}Running integration tests...${NC}"
cd "${PROJECT_ROOT}"

if [ -n "$1" ]; then
    # Run specific test if provided
    echo -e "${YELLOW}Running test: $1${NC}"
    go test ./test/integration/... -v -p 1 -timeout 30m -run "$1"
else
    # Run all integration tests sequentially to avoid PBS resource contention
    go test ./test/integration/... -v -p 1 -timeout 30m
fi

exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
else
    echo -e "${RED}Tests failed with exit code ${exit_code}${NC}"
fi

exit $exit_code
