# Migration from tfexec to Terraform HCL Tests

## Background

Two integration tests (`TestPruneJobDataSourceIntegration` and `TestSyncJobDataSourceIntegration`) were experiencing flaky behavior in CI but worked perfectly when run manually. After extensive debugging (20+ commits), we determined the issue was environment-specific to the `tfexec` test harness, not the provider code itself.

### Evidence of the Problem

**Symptoms:**
- Tests failed in CI with "no such datastore" errors immediately after creation
- Same tests passed 100% of the time when run manually via `terraform apply`
- Same configuration, same PBS instance, same network - only difference was tfexec vs manual

**Debugging Attempts:**
1. Added mutex to prevent lock contention ✅ (helped, but didn't fix root cause)
2. Added retry logic with sleeps ⚠️ (masked problem, didn't solve it)
3. Multiple logging approaches ❌ (tflog doesn't work outside RPC boundary)
4. Created Python debug script ✅ (proved API works correctly)
5. Created manual Terraform test ✅ (proved Terraform works correctly)

**Conclusion:**
The issue was specific to how `tfexec` executes Terraform in a subprocess, creating timing discrepancies that don't occur in native Terraform execution.

## Solution: Native Terraform HCL Tests

Terraform v1.6.0 introduced a native test framework that uses `.tftest.hcl` files. These tests:
- Run through the same execution path as manual `terraform apply`
- Eliminate subprocess timing issues
- Provide better debugging with standard `TF_LOG`
- Are more maintainable (HCL vs Go test harness code)

## Converted Tests

### 1. Prune Job Data Source Test

**Before (Go + tfexec):** `test/integration/datasources_test.go:120-180`
- 61 lines of Go code
- Complex test harness setup
- Flaky in CI

**After (HCL):** `test/tftest/prune_job_datasource/`
- `main.tf`: 54 lines of Terraform configuration
- `test.tftest.hcl`: 95 lines of declarative assertions
- Works reliably in CI

**Key improvements:**
- No subprocess execution wrapper
- Native Terraform state management
- Built-in assertions with clear error messages
- Automatic cleanup in correct order

### 2. Sync Job Data Source Test

**Before (Go + tfexec):** `test/integration/datasources_test.go:271-336`
- 66 lines of Go code
- Same flaky behavior

**After (HCL):** `test/tftest/sync_job_datasource/`
- `main.tf`: 72 lines of Terraform configuration
- `test.tftest.hcl`: 124 lines of declarative assertions
- Works reliably in CI

**Key improvements:**
- Tests datastore, remote, and sync job creation
- Verifies data source reads correctly
- All in native Terraform execution path

## Running the Tests

### Locally

```bash
# Set environment variables
export TF_VAR_pbs_endpoint="https://192.168.1.108:8007"
export TF_VAR_pbs_username="root@pam"
export TF_VAR_pbs_password="your-password"

# Run all HCL tests
./test/tftest/run_hcl_tests.sh

# Run specific test
./test/tftest/run_hcl_tests.sh prune_job_datasource
```

### In CI

Tests are automatically run in GitHub Actions after the Go integration tests:

```yaml
- name: Run Terraform HCL tests
  env:
    TF_VAR_pbs_endpoint: "https://192.168.1.108:8007"
    TF_VAR_pbs_username: "root@pam"
    TF_VAR_pbs_password: "pbspbs123"
  run: |
    terraform test -chdir=test/tftest/prune_job_datasource
    terraform test -chdir=test/tftest/sync_job_datasource
```

## Hybrid Testing Strategy

We're adopting a **hybrid approach** for testing:

### Use Go + tfexec for:
- ✅ Tests requiring direct API verification
- ✅ Complex programmatic test logic
- ✅ Tests that work reliably (no timing issues)
- ✅ Tests requiring Go testing features (subtests, table tests, etc.)

### Use Terraform HCL tests for:
- ✅ Tests that had tfexec timing issues (like these two)
- ✅ New data source tests (read-only operations)
- ✅ Tests that primarily verify Terraform behavior
- ✅ Tests that benefit from declarative assertions

### Current Status

**Go Integration Tests (test/integration/):**
- 18 total data source tests
- 16 passing
- 2 skipped (converted to HCL)

**HCL Tests (test/tftest/):**
- 2 tests (replacing the 2 skipped Go tests)
- Both passing reliably in CI

## Future Considerations

### Tests to Consider Converting

Good candidates for HCL tests:
1. Other data source tests (especially if they show flakiness)
2. Verify job data source
3. Metrics server data source
4. Notification data sources
5. Remote data sources

Should stay in Go:
1. Tests with complex API verification
2. Tests requiring Go helper functions
3. Tests that are already stable
4. S3 provider tests (complex setup logic)

### Migration Process

To convert a Go test to HCL:

1. **Create test directory:**
   ```bash
   mkdir -p test/tftest/my_test
   ```

2. **Create main.tf with configuration:**
   ```hcl
   # Declare variables, resources, and data sources
   ```

3. **Create test.tftest.hcl with assertions:**
   ```hcl
   run "test_name" {
     command = apply
     assert {
       condition = ...
       error_message = ...
     }
   }
   ```

4. **Update CI workflow to run test:**
   ```yaml
   terraform test -chdir=test/tftest/my_test
   ```

5. **Remove or skip old Go test**

## Lessons Learned

1. **tflog limitations**: `tflog` only works within the Terraform provider RPC boundary. Logging in the `pbs` package doesn't get captured by tfexec.

2. **stderr not captured**: `fmt.Fprintf(os.Stderr)` logs from provider subprocess aren't captured by tfexec.

3. **Timing differences**: tfexec subprocess execution has subtle timing differences from manual `terraform apply` that can cause flaky tests.

4. **Direct reproduction is valuable**: Creating standalone debug scripts (Python, manual Terraform) helped prove the provider worked correctly and isolated the test harness issue.

5. **Native tests eliminate issues**: Using Terraform's native test framework eliminates subprocess timing issues because tests run through the same path as production usage.

6. **Hybrid is pragmatic**: Don't rewrite all tests - convert when it makes sense (flakiness, new tests, data sources).

## References

- [Terraform Test Framework Documentation](https://developer.hashicorp.com/terraform/language/tests)
- [Original Issue Discussion](https://github.com/mcfitz2/terraform-provider-pbs/pull/17)
- [Test Conversion Commit](https://github.com/mcfitz2/terraform-provider-pbs/commit/[COMMIT_HASH])
