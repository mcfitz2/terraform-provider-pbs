# Datastore Immutability Test

This directory contains tests that validate the fix for issue #18: Datastore Backend Fields Should Be Immutable.

## Issue Background

Issue #18 identified a bug where updating a datastore with backend fields would fail with:
```
400 - schema does not allow additional properties
```

This occurred because the provider was incorrectly sending immutable backend fields in update requests to the PBS API, even when those fields hadn't changed.

## Test Purpose

This test validates that:
1. Directory datastores can be created successfully
2. Subsequent applies without changes don't fail with 400 errors (the core issue)
3. Mutable fields (like `comment`, `gc_schedule`) can be updated without recreation
4. Immutable fields (like `path`) trigger replacement when changed

**Note:** This test uses directory datastores instead of S3 to avoid external dependencies. The immutability behavior applies to all datastore types.

## Test Scenarios

The test suite includes 5 run blocks:

1. **setup** - Plans initial configuration
2. **create_directory_datastore** - Creates the datastore and verifies all fields
3. **reapply_without_changes** - Applies the same config again (tests issue #18 fix)
4. **update_mutable_fields** - Updates `comment` and `gc_schedule` fields without recreation
5. **plan_immutable_field_change** - Verifies changing `path` triggers replacement

## Running the Test

From this directory:
```bash
terraform test
```

With a specific test ID to avoid name conflicts:
```bash
TF_VAR_test_id="mytest" terraform test
```

## Related Files

- `main.tf` - Terraform configuration with directory datastore resource
- `directory_datastore_immutability.tftest.hcl` - Test suite with assertions

## CI Integration

This test is automatically run as part of the CI workflow in `.github/workflows/vm-integration-tests.yml`.

## S3 Datastore Testing

For S3-specific testing including bucket lifecycle management, see `test/tftest/s3_providers/`.
