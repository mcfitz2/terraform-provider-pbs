# Datastore Immutability Test

This directory contains an HCL-based Terraform test that validates the fix for issue #18.

## Test: s3_datastore_immutability.tftest.hcl

**Issue**: [#18 - Datastore Backend Fields Should Be Immutable](https://github.com/mcfitz2/terraform-provider-pbs/issues/18)

**Purpose**: Validates that the datastore resource correctly handles immutable backend fields.

**What it tests**:
1. **Initial Creation**: S3 datastore can be created with backend fields (path, s3_client, s3_bucket)
2. **Idempotent Apply**: Running `terraform apply` again without changes succeeds (no 400 error)
3. **Mutable Field Updates**: Updating comment field works without recreating the resource
4. **Immutable Field Detection**: Changing backend fields (s3_bucket) triggers resource replacement

**Background**: 
Before the fix, attempting to run `terraform apply` a second time would fail with:
```
Error: Error Updating Datastore
Could not update datastore, unexpected error: failed to update datastore
s3-backup: API request failed with status 400: parameter verification
failed - 'backend': schema does not allow additional properties
```

This occurred because the provider was sending immutable backend configuration fields (path, s3_client, s3_bucket, backing_device, removable) in update API requests, which PBS rejects.

**The Fix**:
1. Added `RequiresReplace()` plan modifiers to immutable backend fields
2. Created `planToDatastoreForUpdate()` method that excludes backend fields from update requests
3. Modified `Update()` method to use the new update-safe converter

**How to run**:
```bash
cd test/tftest/datastore_immutability
terraform init
terraform test
```

**Prerequisites**:
- PBS server running and accessible
- Valid S3 endpoint configuration (can be dummy values for test)
- Provider configured with valid credentials

**Expected behavior**:
- All run blocks should pass
- No 400 errors on subsequent applies
- Comment updates don't trigger replacement
- S3 bucket changes show replacement in plan
