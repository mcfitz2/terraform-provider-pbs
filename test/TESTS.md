# Terraform Provider PBS - Test Suite Documentation

This document describes the comprehensive test suite for the Terraform Provider for Proxmox Backup Server.

## Test Structure

The test suite is organized into several categories, each testing different resource types:

### 1. Datastore Tests (`datastore_test.go`)
- **TestDatastoreDirectoryIntegration**: Tests directory datastore lifecycle (create, read, update, delete)
- **TestDatastoreZFSIntegration**: Tests ZFS datastore functionality
- Tests datastore-specific features like gc_schedule

### 2. S3 Provider Tests (`s3_providers_test.go`)
- **TestS3DatastoreMultiProvider**: Tests S3 datastore with multiple cloud providers (AWS, B2, Scaleway)
- **TestS3EndpointMultiProvider**: Tests S3 endpoint configuration with various providers
- Tests provider-specific configurations and authentication

### 3. Job Tests (`jobs_test.go`)
- **TestPruneJobIntegration**: Tests prune job lifecycle with retention policies
  - Tests: keep_last, keep_daily, keep_weekly, keep_monthly, keep_yearly
  - Validates schedule configuration and comment updates
- **TestPruneJobWithFilters**: Tests prune jobs with backup filtering
  - namespace regex filtering
  - backup_type filtering (vm/ct/host)
  - max_depth configuration
- **TestSyncJobIntegration**: Tests sync job lifecycle
  - Remote server and datastore configuration
  - remove_vanished option
  - rate limiting (rate_limit_kbps)
- **TestSyncJobWithGroupFilter**: Tests sync jobs with group filters
  - Multiple regex pattern filtering
  - Namespace filtering
- **TestVerifyJobIntegration**: Tests verify job lifecycle
  - ignore_verified option
  - outdated_after configuration
  - max_depth traversal
- **TestGCJobIntegration**: Tests garbage collection job lifecycle
  - Simple schedule-based GC
  - Comment and disable options

### 4. Notification Tests (`notifications_test.go`)

#### Notification Targets
- **TestSMTPNotificationIntegration**: SMTP notification target
  - Server, port, authentication configuration
  - mailto list, from/author fields
  - Password handling
- **TestGotifyNotificationIntegration**: Gotify notification target
  - Server URL and token configuration
- **TestSendmailNotificationIntegration**: Sendmail notification target
  - Local sendmail configuration
  - mailto, from, author fields
- **TestWebhookNotificationIntegration**: Webhook notification target
  - URL and HTTP method configuration
  - Custom webhook endpoints

#### Notification Routing
- **TestNotificationEndpointIntegration**: Notification endpoints (target groups)
  - Groups multiple notification targets
  - Target list management
- **TestNotificationMatcherIntegration**: Notification matchers (routing rules)
  - match_severity filtering (error, warning, info, notice)
  - match_field custom field matching
  - Target routing to endpoints
- **TestNotificationMatcherModes**: Tests matcher modes (all vs any)
  - "all" mode: All conditions must match
  - "any" mode: At least one condition must match
- **TestNotificationMatcherWithCalendar**: Calendar-based filtering
  - systemd calendar event format
  - Business hours filtering example
- **TestNotificationMatcherInvertMatch**: Inverted matching
  - Notify when conditions DON'T match
  - Useful for exclusion patterns

### 5. Metrics Tests (`metrics_test.go`)
- **TestMetricsServerInfluxDBHTTPIntegration**: InfluxDB HTTP metrics export
  - URL, organization, bucket configuration
  - Token authentication
  - HTTPS/TLS configuration
- **TestMetricsServerInfluxDBUDPIntegration**: InfluxDB UDP metrics export
  - Host and port configuration
  - UDP/TCP protocol selection
- **TestMetricsServerMTU**: Custom MTU configuration
  - Network packet size optimization
- **TestMetricsServerVerifyCertificate**: TLS certificate verification
  - verify_tls option testing
- **TestMetricsServerDisabled**: Disabled server configuration
  - enable/disable toggle
- **TestMetricsServerMaxBodySize**: Custom max body size
  - HTTP request size limits
- **TestMetricsServerTimeout**: Custom timeout configuration
  - HTTP request timeout handling
- **TestMetricsServerTypeChange**: Type change behavior
  - Validates that type changes trigger resource replacement

## Running Tests

### Prerequisites
1. Build the provider binary:
   ```bash
   go build .
   ```

2. Set up environment variables:
   ```bash
   export PBS_ADDRESS="https://your-pbs-server:8007"
   export PBS_USERNAME="root@pam"
   export PBS_PASSWORD="your-password"
   ```

3. For S3 tests, also set:
   ```bash
   export AWS_ACCESS_KEY_ID="your-aws-key"
   export AWS_SECRET_ACCESS_KEY="your-aws-secret"
   export AWS_REGION="us-west-2"
   # And/or B2, Scaleway credentials
   ```

### Run All Tests
```bash
cd test
go test -v -timeout 30m
```

### Run Specific Test Categories
```bash
# Run only datastore tests
go test -v -run TestDatastore

# Run only job tests
go test -v -run TestIntegration/Jobs

# Run only notification tests
go test -v -run TestIntegration/Notifications

# Run only metrics tests
go test -v -run TestIntegration/Metrics
```

### Run Quick Smoke Tests
```bash
go test -v -run TestQuickSmoke -timeout 5m
```

### Run Tests in Short Mode (Skip Integration Tests)
```bash
go test -short -v
```

## Test Coverage

### Resources Tested
- ✅ `pbs_datastore` (directory and S3)
- ✅ `pbs_s3_endpoint`
- ✅ `pbs_prune_job`
- ✅ `pbs_sync_job`
- ✅ `pbs_verify_job`
- ✅ `pbs_gc_job`
- ✅ `pbs_smtp_notification`
- ✅ `pbs_gotify_notification`
- ✅ `pbs_sendmail_notification`
- ✅ `pbs_webhook_notification`
- ✅ `pbs_notification_endpoint`
- ✅ `pbs_notification_matcher`
- ✅ `pbs_metrics_server` (influxdb-http and influxdb-udp)

### CRUD Operations Tested
Each resource test validates:
- ✅ **Create**: Resource creation via Terraform
- ✅ **Read**: State retrieval and API verification
- ✅ **Update**: Configuration changes and updates
- ✅ **Delete**: Resource cleanup (via DestroyTerraform)

### Advanced Features Tested
- Resource dependencies (endpoints depend on targets)
- Complex field configurations (retention policies, filters, etc.)
- Optional vs required fields
- Default values
- Type-specific fields (InfluxDB HTTP vs UDP)
- Validation (backup_type, matcher mode, etc.)
- Inverted matching logic
- Calendar event parsing
- Rate limiting
- TLS/SSL configuration

## Test Utilities

### TestContext
The `TestContext` struct provides:
- PBS API client for direct API calls
- Terraform executor for plan/apply/destroy operations
- Temporary working directory
- Test configuration management

### Helper Functions
- `GenerateTestName(prefix string)`: Creates unique test resource names
- `SetupTest(t *testing.T)`: Initializes test environment
- `WriteMainTF(config string)`: Writes Terraform configuration
- `ApplyTerraform()`: Applies Terraform configuration
- `GetResourceFromState(resource string)`: Retrieves resource from state
- `DestroyTerraform()`: Cleans up all test resources

## Integration Test Strategy

1. **Setup Phase**: Create test context and API client
2. **Configuration Phase**: Write Terraform configuration
3. **Apply Phase**: Run `terraform apply`
4. **Verification Phase**: 
   - Check Terraform state
   - Verify via direct API calls
5. **Update Phase**: Modify configuration and re-apply
6. **Update Verification**: Verify changes took effect
7. **Cleanup Phase**: Run `terraform destroy`

## Best Practices

### Test Naming
- Use descriptive test names that indicate what is being tested
- Follow the pattern: `Test<Resource><Feature>Integration`
- Use subtests for related test cases

### Resource Naming
- Always use `GenerateTestName()` to avoid conflicts
- Include test type in prefix (e.g., "prune-job", "smtp-target")

### Assertions
- Use `require` for critical checks (test should stop if failed)
- Use `assert` for non-critical checks (test continues)
- Always verify both Terraform state AND direct API calls

### Cleanup
- Use `defer tc.DestroyTerraform(t)` immediately after setup
- Set `PBS_DESTROY_DATA_ON_DELETE=true` environment variable
- Ensure unique resource names to avoid conflicts

### Error Handling
- Check for errors after API calls
- Provide descriptive error messages
- Use `require.NoError` for setup operations
- Log informational messages with `t.Logf()`

## Continuous Integration

The tests are designed to run in CI/CD pipelines with:
- Parallel execution support (where applicable)
- Timeout protection (30-minute default)
- Short mode for quick validation
- Environment-based configuration
- Automatic cleanup on failure

## Future Test Additions

Potential areas for expansion:
- [ ] Import testing for all resources
- [ ] Concurrent resource creation/deletion
- [ ] Large-scale tests (many resources)
- [ ] Performance benchmarking
- [ ] Edge case validation (invalid configurations)
- [ ] Network failure simulation
- [ ] API rate limiting handling
- [ ] Backup/restore workflows
- [ ] Multi-datastore sync chains

## Troubleshooting

### Common Issues

1. **Provider binary not found**
   - Run `go build .` in project root

2. **Connection refused**
   - Verify PBS_ADDRESS is correct
   - Check PBS server is running

3. **Authentication failed**
   - Verify PBS_USERNAME and PBS_PASSWORD
   - Ensure user has appropriate permissions

4. **Tests timeout**
   - Increase timeout: `go test -timeout 60m`
   - Check PBS server performance

5. **Resource already exists**
   - Ensure previous test cleanup completed
   - Use unique names with `GenerateTestName()`

6. **Tests fail on cleanup**
   - Check `PBS_DESTROY_DATA_ON_DELETE=true` is set
   - Verify user has delete permissions
