# Integration Tests for Proxmox Backup Server Terraform Provider

This directory contains integration tests that validate the Terraform provider against a real Proxmox Backup Server (PBS) instance.

## Prerequisites

1. **Go 1.21+** - Required for running the tests
2. **Terraform** - Must be installed and available in PATH
3. **PBS Instance** - A running Proxmox Backup Server instance for testing

## Setup Options

### Option 1: Docker Container (Recommended for Development)

The easiest way to run tests is using the included Docker setup:

```bash
# Run all tests with Docker PBS container
./test/run_docker_tests.sh

# Run specific tests with verbose output
./test/run_docker_tests.sh -v TestQuickSmoke

# Start container only (for manual testing)
./test/run_docker_tests.sh --start-only

# Clean up containers
./test/run_docker_tests.sh --cleanup
```

**Docker Prerequisites:**
- Docker and Docker Compose installed
- No additional PBS instance required

**Default Docker Credentials:**
- URL: `https://localhost:8007`
- Username: `admin@pam`
- Password: `password123`

### Option 2: External PBS Instance

#### Environment Variables

For testing against an external PBS instance:

```bash
export PBS_ADDRESS="https://your-pbs-server:8007"     # PBS server address
export PBS_USERNAME="your-username@pam"              # PBS username (e.g., admin@pam)
export PBS_PASSWORD="your-password"                  # PBS password
export PBS_INSECURE_TLS="true"                       # For self-signed certificates
```

### Multi-Provider S3 Testing Environment Variables

For multi-provider S3 endpoint testing (optional - enables real S3 bucket testing):

#### AWS S3
```bash
export AWS_ACCESS_KEY_ID="your-aws-access-key"       # AWS access key
export AWS_SECRET_ACCESS_KEY="your-aws-secret-key"   # AWS secret key
export AWS_REGION="us-east-1"                        # AWS region (optional, default: us-east-1)
```

#### Backblaze B2
```bash
export B2_ACCESS_KEY_ID="your-b2-key-id"            # Backblaze B2 key ID
export B2_SECRET_ACCESS_KEY="your-b2-application-key" # Backblaze B2 application key
export B2_REGION="us-west-004"                       # B2 region (optional, default: us-west-004)
```

#### Scaleway Object Storage
```bash
export SCALEWAY_ACCESS_KEY="your-scaleway-access-key"    # Scaleway access key
export SCALEWAY_SECRET_KEY="your-scaleway-secret-key"    # Scaleway secret key
export SCALEWAY_REGION="fr-par"                          # Scaleway region (optional, default: fr-par)
```

### Optional Environment Variables

```bash
export PBS_INSECURE_TLS="true"                      # Skip TLS certificate verification (default: false)
```

### Multi-Provider S3 Setup

The multi-provider S3 tests are optional but provide comprehensive validation of S3 endpoint functionality with real cloud providers. Set up credentials for the providers you want to test:

#### AWS S3 Setup
1. Create an AWS account and obtain access keys
2. Ensure your AWS user has S3 permissions: `s3:CreateBucket`, `s3:DeleteBucket`, `s3:ListBucket`, `s3:PutObject`, `s3:GetObject`
3. Set the AWS environment variables as shown above

#### Backblaze B2 Setup
1. Create a Backblaze account and generate application keys
2. Create an application key with read/write permissions
3. Note: B2 uses path-style addressing by default
4. Set the B2 environment variables as shown above

#### Scaleway Object Storage Setup
1. Create a Scaleway account and generate API keys
2. Enable Object Storage in your Scaleway project
3. Generate access keys with Object Storage permissions
4. Set the Scaleway environment variables as shown above

**Note:** Multi-provider tests will automatically skip providers that don't have credentials configured, so you can run tests with any subset of providers configured.

## Running Tests

### Quick Start with Docker (Recommended)

```bash
# Run all tests with Docker PBS (zero configuration)
./test/run_docker_tests.sh

# Run specific tests with verbose output
./test/run_docker_tests.sh -v TestQuickSmoke

# Keep container running for debugging
./test/run_docker_tests.sh -k TestDatastoreDirectoryIntegration
```

### Quick Start with External PBS

```bash
# Set environment variables
export PBS_ADDRESS="https://pbs.local:8007"
export PBS_USERNAME="admin@pam"
export PBS_PASSWORD="secret"

# Run integration tests
make test-integration
```

### Individual Test Commands

```bash
# Run all tests (unit + integration)
make test

# Run only unit tests (no PBS instance required)
make test-unit

# Run only integration tests
make test-integration

# Run specific integration test
go test -v ./test/ -run TestS3EndpointIntegration

# Run multi-provider S3 tests (requires S3 credentials)
go test -v ./test/ -run TestS3EndpointMultiProvider

# Run single provider test
go test -v ./test/ -run TestS3EndpointMultiProvider/AWS

# Run tests with verbose output and timeout
go test -v -timeout 30m ./test/...
```

## Test Structure

### Test Files

- `setup.go` - Common test setup and utility functions
- `integration_test.go` - Main test entry point with organized test suites
- `datastore_test.go` - Complete datastore lifecycle tests (directory, S3)
- `s3_providers.go` - Multi-provider S3 configuration and utilities  
- `s3_providers_test.go` - Multi-provider S3 integration tests (AWS, Backblaze B2, Scaleway)
- `Makefile` - Build and test automation
- `run_integration_tests.sh` - Shell script for running integration tests

### Test Categories

1. **Integration Suite** (`TestIntegration`) - Complete provider functionality
   - **Datastore Tests** - Directory and S3 datastore lifecycle
   - **S3 Endpoint Tests** - Multi-provider S3 endpoint management
2. **Smoke Tests** (`TestQuickSmoke`) - Fast basic connectivity tests
3. **Multi-Provider Tests** - Real cloud provider integration (AWS, Backblaze B2, Scaleway)
4. **Provider Quirks** - Backblaze B2 compatibility with `skip-if-none-match-header`

## Test Scenarios

### Basic S3 Endpoint Tests

#### TestS3EndpointIntegration
- Creates an S3 endpoint via Terraform
- Verifies creation through both Terraform state and direct API calls
- Updates the endpoint configuration
- Verifies updates through API
- Cleans up resources

#### TestS3EndpointValidation
- Tests invalid configurations
- Validates proper error handling
- Ensures required fields are enforced

#### TestS3EndpointConcurrency
- Creates multiple S3 endpoints simultaneously
- Verifies concurrent operations work correctly
- Tests resource isolation

### Multi-Provider S3 Endpoint Tests

#### TestS3EndpointMultiProvider
- Tests S3 endpoint creation with real cloud providers
- Creates actual S3 buckets for each configured provider
- Validates PBS S3 endpoint connectivity to real S3 services
- Tests with AWS S3, Backblaze B2, and Scaleway Object Storage
- Automatically cleans up S3 buckets after testing
- Runs in parallel for faster execution

#### TestS3EndpointProviderSpecificFeatures
- Tests provider-specific S3 configurations
- Validates path-style vs virtual-hosted-style addressing
- Tests region-specific endpoint configurations
- Verifies compatibility across different S3 implementations

#### TestS3EndpointConcurrentProviders
- Creates multiple S3 endpoints across different providers simultaneously
- Tests concurrent access to different S3 services
- Validates provider isolation and independence

## Adding New Tests

### Test File Template

```go
package test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewResource(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    tc := SetupTest(t)
    defer tc.DestroyTerraform(t)

    // Your test logic here
}
```

### Best Practices

1. **Use unique names** - Always use `GenerateTestName()` for resource names
2. **Clean up resources** - Use `defer tc.DestroyTerraform(t)` in each test
3. **Verify via API** - Always verify Terraform changes through direct API calls
4. **Handle errors gracefully** - Use `require.NoError()` for critical operations
5. **Test edge cases** - Include validation and error scenarios

## Troubleshooting

### Common Issues

#### PBS Connection Issues
```
Error: Failed to create PBS API client: connection refused
```
- Verify PBS_ADDRESS is correct and accessible
- Check if PBS instance is running
- Verify firewall allows connections on port 8007

#### Authentication Issues
```
Error: authentication failed
```
- Verify PBS_USERNAME and PBS_PASSWORD are correct
- Check if user has necessary permissions
- Verify user account is not disabled

#### TLS Issues
```
Error: x509: certificate signed by unknown authority
```
- Set `PBS_INSECURE_TLS=true` for self-signed certificates
- Or provide proper CA certificates

#### Resource Already Exists
```
Error: resource already exists
```
- Previous test may not have cleaned up properly
- Check PBS admin interface for leftover test resources
- Manually clean up resources with test names

#### S3 Provider Issues

##### AWS S3 Issues
```
Error: InvalidAccessKeyId: The AWS Access Key Id you provided does not exist
```
- Verify AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are correct
- Check if AWS user has sufficient S3 permissions
- Verify AWS region is supported

##### Backblaze B2 Issues
```
Error: 401 Unauthorized
```
- Verify B2_ACCESS_KEY_ID and B2_SECRET_ACCESS_KEY are correct
- Check if application key has bucket create/delete permissions
- Note: B2 requires path-style addressing

##### Scaleway Issues
```
Error: SignatureDoesNotMatch
```
- Verify SCALEWAY_ACCESS_KEY and SCALEWAY_SECRET_KEY are correct
- Check if keys have Object Storage permissions
- Verify region is correctly set (fr-par, nl-ams, pl-waw)

##### General S3 Issues
```
Error: BucketAlreadyExists
```
- S3 bucket names must be globally unique
- Tests use randomized names to avoid conflicts
- If tests fail to cleanup, manually delete test buckets

```
Error: region mismatch
```
- Ensure region environment variables match your provider configuration
- Some providers require specific regions

### Debug Mode

Run tests with additional debugging:

```bash
# Enable Go test verbose mode
go test -v ./test/...

# Run with race detection
go test -race ./test/...

# Run with coverage
go test -cover ./test/...
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - name: Run Integration Tests
      env:
        PBS_ADDRESS: ${{ secrets.PBS_ADDRESS }}
        PBS_USERNAME: ${{ secrets.PBS_USERNAME }}
        PBS_PASSWORD: ${{ secrets.PBS_PASSWORD }}
      run: make test-integration
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any
    environment {
        PBS_ADDRESS = credentials('pbs-address')
        PBS_USERNAME = credentials('pbs-username')  
        PBS_PASSWORD = credentials('pbs-password')
    }
    stages {
        stage('Integration Tests') {
            steps {
                sh 'make test-integration'
            }
        }
    }
}
```

## Safety Considerations

- Integration tests create and delete real resources on the PBS instance
- Use a dedicated test PBS instance, not production
- Test resources use predictable naming patterns (`*-test-*`)
- All tests clean up resources automatically
- Failed tests may leave resources that need manual cleanup

## Contributing

When adding new integration tests:

1. Follow the existing test patterns
2. Add comprehensive error handling
3. Include both positive and negative test cases
4. Update this README with new test descriptions
5. Ensure tests are isolated and don't interfere with each other