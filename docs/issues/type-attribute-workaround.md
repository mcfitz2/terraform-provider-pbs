# Workaround for "Missing required argument 'type'" Error

## TL;DR - The Fix

**You don't actually need a `type` attribute at the root level.** The error is misleading. Here's what's likely happening and how to fix it:

## Most Common Cause: maintenance_mode Block

If you have a `maintenance_mode` block in your configuration, it **requires** a `type` field inside it. The error message is confusing because Terraform reports it at the wrong nesting level.

### ❌ Wrong (causes the error):
```hcl
resource "pbs_datastore" "s3_backup" {
  name       = "my-datastore"
  path       = "/datastore/cache"
  s3_client  = "my-s3-endpoint"
  s3_bucket  = "my-bucket"
  
  maintenance_mode {
    message = "Down for maintenance"
    # ERROR: Missing required 'type' field!
  }
}
```

### ✅ Correct:
```hcl
resource "pbs_datastore" "s3_backup" {
  name       = "my-datastore"
  path       = "/datastore/cache"
  s3_client  = "my-s3-endpoint"
  s3_bucket  = "my-bucket"
  
  maintenance_mode {
    type    = "offline"  # or "read-only"
    message = "Down for maintenance"
  }
}
```

### ✅ Or just remove it if you don't need maintenance mode:
```hcl
resource "pbs_datastore" "s3_backup" {
  name       = "my-datastore"
  path       = "/datastore/cache"
  s3_client  = "my-s3-endpoint"
  s3_bucket  = "my-bucket"
  comment    = "My S3 datastore"
}
```

## Other Possible Causes & Fixes

### 1. Check Your Configuration

Search for any `type =` in your datastore configuration:
```bash
grep -n "type" your_terraform_file.tf
```

If you see `type =` at the root level of your `pbs_datastore` resource, **remove it**. It's not a valid attribute.

### 2. Clear Stale State

If you upgraded from an older provider version, Terraform state might be confused:
```bash
# List your datastores
terraform state list | grep pbs_datastore

# Remove the problematic one
terraform state rm pbs_datastore.s3_backup

# Re-import or re-create it
terraform apply
```

### 3. Upgrade Provider

Make sure you're using the latest version (v0.2.0):
```hcl
terraform {
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "~> 0.2.0"
    }
  }
}
```

Then run:
```bash
terraform init -upgrade
```

## Why This Happens

The `pbs_datastore` resource has **no root-level `type` attribute**. The backend type is automatically inferred from the attributes you provide:

- **Directory datastore**: Just provide `name` and `path`
- **S3 datastore**: Provide `name`, `path`, `s3_client`, and `s3_bucket`
- **Removable datastore**: Provide `name`, `path`, `removable = true`, and `backing_device`

The only `type` field in the schema is **inside** the optional `maintenance_mode` block, where it's required if you use that block.

## Working S3 Datastore Example

Here's a complete, working example:

```hcl
# Create S3 endpoint first
resource "pbs_s3_endpoint" "aws" {
  id          = "my-aws-s3"
  endpoint    = "https://s3.amazonaws.com"
  region      = "us-east-1"
  access_key  = var.aws_access_key
  secret_key  = var.aws_secret_key
}

# Create S3-backed datastore
resource "pbs_datastore" "s3_backup" {
  name       = "s3-backup"
  path       = "/datastore/s3-cache"    # Local cache directory
  s3_client  = pbs_s3_endpoint.aws.id   # Reference the S3 endpoint
  s3_bucket  = "my-backup-bucket"
  comment    = "S3-backed datastore for backups"
  
  # Optional: garbage collection schedule
  gc_schedule = "daily"
  
  # Optional: pruning configuration
  keep_last   = 7
  keep_daily  = 14
  keep_weekly = 8
  
  depends_on = [pbs_s3_endpoint.aws]
}
```

## Still Not Working?

If the error persists after trying these fixes:

1. **Share your full configuration** (remove sensitive data)
2. **Run this diagnostic:**
   ```bash
   terraform version
   terraform providers
   terraform validate 2>&1 | tee error.log
   ```
3. **Check the provider schema:**
   ```bash
   terraform providers schema -json | \
     jq '.provider_schemas["registry.terraform.io/micah/pbs"].resource_schemas.pbs_datastore.block.attributes | keys'
   ```
   
   This should NOT include "type" in the list. If it does, something is very wrong.

## Technical Background

The confusion arises because:
1. Terraform's error messages for nested attributes can be misleading
2. The `maintenance_mode.type` required field might get reported at the wrong nesting level
3. Some people expect a `type` field because other systems (like Proxmox VE) use it to distinguish backend types

In the PBS provider, backend type is **inferred** from which attributes you set, not explicitly declared. This is intentional and matches how the PBS API works.

## Summary

**Solution**: Check for a `maintenance_mode` block in your config. If present, add `type = "offline"` or `type = "read-only"` inside it. If not using maintenance mode, remove the block entirely. There is no root-level `type` attribute needed or supported.
