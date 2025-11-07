# Bug Investigation: Missing Required `type` Attribute Error

## Issue Report Summary
User reported getting an error: "The argument 'type' is required, but no definition was found" when creating an S3 datastore.

## Investigation Results

### Schema Analysis ✅ CORRECT
The `pbs_datastore` resource schema in `fwprovider/resources/datastores/datastore.go` is correctly defined:

- **No root-level `type` attribute exists** - Verified at line 119-400
- **Only nested `type` attribute** - Inside `maintenance_mode` block (line 300)
  - `maintenance_mode` is optional (line 294)
  - `maintenance_mode.type` is required when the block is present (line 302)
- **All validation logic is correct** - validateDatastoreConfig() does not check for `type` field

### Backend Type Inference ✅ WORKING
The code correctly infers datastore backend type (lines 994-1002):
```go
if isRemovable {
    ds.Backend = datastores.FormatBackendString("removable", ...)
} else if ds.S3Client != "" && ds.S3Bucket != "" {
    ds.Backend = datastores.FormatBackendString("s3", ...)
}
// Otherwise: directory datastore (no explicit backend string needed)
```

### Reproduction Attempt ❌ CANNOT REPRODUCE
Created test configuration in `test_type_bug.tf`:
```hcl
resource "pbs_s3_endpoint" "test" {
  id          = "test-endpoint"
  endpoint    = "https://s3.amazonaws.com"
  region      = "us-east-1"
  access_key  = "test"
  secret_key  = "test"
}

resource "pbs_datastore" "s3_backup" {
  name       = "test-s3-ds"
  path       = "/datastore/test-s3-cache"
  s3_client  = pbs_s3_endpoint.test.id
  s3_bucket  = "my-bucket"
  comment    = "Test S3 datastore"
}
```

Result: `terraform validate` succeeds with current codebase (v0.2.0-dev)

## Possible Causes

### 1. ❓ Terraform State Confusion
If the user has stale state from an older provider version, Terraform might be confused about the schema.

**Solution:** Run `terraform state rm` and re-create the resource

### 2. ❓ maintenance_mode Block Issue
If the user accidentally has a `maintenance_mode` block without the required `type` field:

```hcl
# WRONG - will fail
resource "pbs_datastore" "test" {
  name = "test"
  path = "/datastore/test"
  maintenance_mode {
    message = "Down for maintenance"
    # MISSING: type = "offline" or "read-only"
  }
}
```

**Error message would be misleading** - Terraform might report it at the wrong nesting level.

### 3. ❓ Older Provider Version Bug
The user mentioned using `~> 0.1`, but v0.2.0 was just released. There might have been a bug in v0.1.x that's already fixed.

**Solution:** Upgrade to v0.2.0

### 4. ❓ Configuration Typo
User might have accidentally typed `type = ...` at root level due to autocomplete or confusion with other resources.

**Solution:** Remove the `type` attribute from root level

## Recommendations for User

### Quick Fixes to Try:

1. **Check for maintenance_mode block:**
   ```bash
   grep -n "maintenance_mode" main.tf
   ```
   If found, ensure it has `type` field:
   ```hcl
   maintenance_mode {
     type    = "offline"  # or "read-only"
     message = "..."
   }
   ```

2. **Remove stale state:**
   ```bash
   terraform state rm pbs_datastore.s3_backup
   terraform apply
   ```

3. **Upgrade provider:**
   ```hcl
   terraform {
     required_providers {
       pbs = {
         source  = "registry.terraform.io/micah/pbs"
         version = "~> 0.2.0"  # Use latest
       }
     }
   }
   ```
   Then: `terraform init -upgrade`

4. **Check for typos:**
   ```bash
   grep -n "type.*=" main.tf
   ```
   Remove any `type =` lines at the datastore resource root level.

### If Still Failing:

Please provide:
1. Full Terraform configuration (sanitized)
2. Complete error output
3. Provider version: `terraform providers`
4. Terraform version: `terraform version`
5. Output of: `terraform providers schema -json | jq '.provider_schemas["registry.terraform.io/micah/pbs"].resource_schemas.pbs_datastore.block.attributes | keys'`

## Code Verification

The current codebase (v0.2.0) correctly implements S3 datastores without requiring a `type` field. The integration tests confirm this works:

- `test/integration/s3_providers_test.go` - Uses S3 datastores successfully
- `test/integration/datastore_test.go` - Directory datastore tests pass

## Conclusion

**Cannot reproduce the bug with current code.** Either:
- Bug was in older version and is now fixed
- User has a configuration issue (maintenance_mode or typo)
- User needs to upgrade provider or clear stale state

The provider code is correct and does not require a root-level `type` attribute for any datastore type.
