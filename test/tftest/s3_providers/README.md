# S3 Provider Tests

This directory contains HCL-based Terraform tests that validate S3-backed datastore functionality with multiple S3-compatible storage providers.

## Overview

These tests create real S3 buckets, PBS S3 endpoints, and datastores for each supported provider:
- **AWS S3** - Amazon's native S3 service
- **Backblaze B2** - S3-compatible object storage with specific quirks
- **Scaleway Object Storage** - European S3-compatible storage

Each test validates the complete lifecycle:
1. Create S3 bucket (managed by Terraform's AWS provider)
2. Create PBS S3 endpoint with provider-specific configuration
3. Create PBS datastore using the S3 backend
4. Verify all resources and relationships
5. Test idempotency (no drift)
6. Clean up all resources (bucket automatically deleted)

## Provider-Specific Configuration

### AWS S3
```hcl
s3_endpoint      = "s3.{region}.amazonaws.com"
path_style       = true
provider_quirks  = []
```

### Backblaze B2
```hcl
s3_endpoint      = "s3.{region}.backblazeb2.com"
path_style       = true  # REQUIRED
provider_quirks  = ["skip-if-none-match-header"]  # REQUIRED
```
Backblaze B2's S3 API has limited compatibility. The `skip-if-none-match-header` quirk prevents PBS from sending If-None-Match headers that B2 doesn't support (would return 501 errors).

### Scaleway Object Storage
```hcl
s3_endpoint      = "s3.{region}.scw.cloud"
path_style       = true
provider_quirks  = []
```

## Running the Tests

### Prerequisites

Each test requires specific environment variables for the corresponding provider. Tests are automatically skipped if the required credentials are not present.

**AWS S3:**
```bash
export AWS_ACCESS_KEY_ID="your-aws-access-key"
export AWS_SECRET_ACCESS_KEY="your-aws-secret-key"
export AWS_REGION="us-west-2"  # optional, defaults to us-west-2
```

**Backblaze B2:**
```bash
export B2_ACCESS_KEY_ID="your-b2-application-key-id"
export B2_SECRET_ACCESS_KEY="your-b2-application-key"
export B2_REGION="us-west-004"  # optional, defaults to us-west-004
```

**Scaleway:**
```bash
export SCALEWAY_ACCESS_KEY="your-scaleway-access-key"
export SCALEWAY_SECRET_KEY="your-scaleway-secret-key"
export SCALEWAY_REGION="fr-par"  # optional, defaults to fr-par
```

### Run All S3 Provider Tests

From this directory:
```bash
# Set PBS connection details
export TF_VAR_pbs_endpoint="https://your-pbs-server:8007"
export TF_VAR_pbs_username="root@pam"
export TF_VAR_pbs_password="your-password"

# Set test ID (unique per run)
export TF_VAR_test_id="$(date +%s)"

# Set S3 provider credentials (see above)

# Run tests
terraform test
```

### Run Individual Provider Tests

**AWS only:**
```bash
terraform test -filter=tests/aws.tftest.hcl
```

**Backblaze only:**
```bash
terraform test -filter=tests/backblaze.tftest.hcl
```

**Scaleway only:**
```bash
terraform test -filter=tests/scaleway.tftest.hcl
```

## CI Integration

These tests are automatically run in CI when the corresponding provider credentials are available:
- Configured in `.github/workflows/vm-integration-tests.yml`
- Each provider test runs independently
- Tests are skipped if credentials are not set
- S3 buckets are automatically cleaned up after each test

## Test Structure

Each test file follows the same pattern:

1. **setup_{provider}** - Validates configuration
2. **create_{provider}_s3_infrastructure** - Creates bucket, endpoint, datastore
3. **verify_{provider}_no_drift** - Ensures no configuration drift
4. **update/test_{provider}_*** - Provider-specific validation

## Bucket Lifecycle Management

S3 buckets are managed by Terraform's AWS provider with:
- **Unique naming**: `pbs-test-{provider}-{test_id}`
- **force_destroy**: Allows deletion of non-empty buckets
- **Automatic cleanup**: Terraform destroy removes buckets
- **Tags**: All buckets tagged with test metadata

## Troubleshooting

### Bucket Already Exists
If a test fails midway, the bucket may still exist. Either:
1. Manually delete the bucket via provider console
2. Use a different test_id
3. Wait for eventual consistency (some providers delay bucket deletion)

### Backblaze 501 Errors
Ensure `provider_quirks = ["skip-if-none-match-header"]` is set. Without this, PBS will send unsupported headers to B2.

### Permission Errors
Ensure S3 credentials have full bucket permissions:
- CreateBucket, DeleteBucket
- PutObject, GetObject, DeleteObject
- ListBucket

### Credential Issues
- AWS: Use IAM user credentials, not root account
- Backblaze: Use Application Keys, not Master Application Key
- Scaleway: Use API keys with Object Storage permissions

## Related Tests

- **Directory Datastore Tests**: `test/tftest/datastores/`
- **Immutability Tests**: `test/tftest/datastore_immutability/`
- **Legacy Go Tests**: `test/integration/s3_providers_test.go` (being replaced)

## Migration from Go Tests

These HCL tests replace the Go-based tfexec tests in `test/integration/s3_providers_test.go`:
- ✅ Better bucket lifecycle management (Terraform-native)
- ✅ No external bucket creation scripts needed
- ✅ Automatic cleanup via Terraform destroy
- ✅ Native Terraform test framework (v1.6+)
- ✅ Easier to maintain and extend
