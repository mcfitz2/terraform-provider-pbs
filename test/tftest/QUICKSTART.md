# Terraform HCL Tests - Quick Start

## What Are These?

These are **native Terraform tests** using the `.tftest.hcl` format introduced in Terraform v1.6.0. They replace two flaky Go integration tests that had timing issues with the `tfexec` test harness.

## Why Convert?

The converted tests (`TestPruneJobDataSourceIntegration` and `TestSyncJobDataSourceIntegration`) were failing in CI with "datastore not found" errors but worked perfectly when run manually. After extensive debugging, we determined this was a tfexec-specific timing issue, not a provider bug. Native Terraform tests eliminate this issue by using the same execution path as production usage.

## Prerequisites

- **Terraform v1.6.0+** (HCL tests require v1.6+)
- **Built provider binary** (`go build .`)
- **PBS server** running and accessible

## Quick Start

### 1. Set Environment Variables

```bash
export TF_VAR_pbs_endpoint="https://192.168.1.108:8007"
export TF_VAR_pbs_username="root@pam"
export TF_VAR_pbs_password="your-password"
```

### 2. Run Tests

```bash
# Run all HCL tests
./test/tftest/run_hcl_tests.sh

# Run specific test
./test/tftest/run_hcl_tests.sh prune_job_datasource
```

## Test Coverage

✅ **Prune Job Data Source** (`prune_job_datasource/`)
- Creates datastore and prune job
- Reads job via data source
- Verifies all attributes match

✅ **Sync Job Data Source** (`sync_job_datasource/`)
- Creates datastore, remote, and sync job
- Reads job via data source  
- Verifies all attributes match

## Documentation

- **[README.md](README.md)** - Detailed usage instructions
- **[TFEXEC_TO_HCL_MIGRATION.md](../../docs/TFEXEC_TO_HCL_MIGRATION.md)** - Full migration story and lessons learned

## CI Integration

These tests run automatically in GitHub Actions after the Go integration tests:

```yaml
- name: Run Terraform HCL tests
  run: |
    terraform test -chdir=test/tftest/prune_job_datasource
    terraform test -chdir=test/tftest/sync_job_datasource
```

## Debugging

Enable detailed logging:

```bash
export TF_LOG=DEBUG
terraform test -chdir=test/tftest/prune_job_datasource
```

## Status

- ✅ Both HCL tests passing reliably in CI
- ✅ Replace 2 flaky Go tests (now skipped)
- ✅ Same test coverage, better reliability
