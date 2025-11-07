# Backblaze B2 S3-Compatible Provider Test
#
# This test validates the complete lifecycle of S3-backed datastores using Backblaze B2:
# 1. Creates a B2 bucket via S3 API
# 2. Creates a PBS S3 endpoint with B2-specific quirks
# 3. Creates a PBS datastore using the B2 backend
# 4. Verifies all resources work with B2's S3 compatibility layer
# 5. Cleans up all resources
#
# Backblaze B2 Requirements:
# - path_style = true (required)
# - provider_quirks = ["skip-if-none-match-header"] (prevents 501 errors)
# - Endpoint format: s3.{region}.backblazeb2.com

variables {
  s3_provider_name   = "Backblaze"
  s3_endpoint        = "s3.${var.pbs_s3_region}.backblazeb2.com"
  s3_region          = "us-east-1"  # Generic region for AWS provider
  pbs_s3_region      = "us-east-005"  # Actual B2 region for PBS endpoint
  s3_bucket_name     = "pbs-test-b2-${var.test_id}"
  s3_endpoint_id     = "pbs-b2-${var.test_id}"
  datastore_name     = "b2-ds-${var.test_id}"
  s3_provider_quirks = ["skip-if-none-match-header"]
}

run "setup_backblaze" {
  command = plan
  
  assert {
    condition     = var.s3_provider_name == "Backblaze"
    error_message = "Provider name should be Backblaze"
  }
  
  assert {
    condition     = length(var.s3_provider_quirks) > 0
    error_message = "Backblaze requires provider quirks"
  }
}

run "create_backblaze_s3_infrastructure" {
  command = apply
  
  assert {
    condition     = local.bucket.bucket == "pbs-test-b2-${var.test_id}"
    error_message = "B2 bucket name should match expected pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.id == "pbs-b2-${var.test_id}"
    error_message = "S3 endpoint ID should match expected pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.endpoint == "s3.${var.s3_region}.backblazeb2.com"
    error_message = "S3 endpoint should match Backblaze pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.path_style == true
    error_message = "Path style must be enabled for Backblaze"
  }
  
  assert {
    condition     = contains(pbs_s3_endpoint.test.provider_quirks, "skip-if-none-match-header")
    error_message = "Backblaze requires skip-if-none-match-header quirk"
  }
  
  assert {
    condition     = pbs_datastore.test.name == "b2-ds-${var.test_id}"
    error_message = "Datastore name should match expected pattern"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_client == pbs_s3_endpoint.test.id
    error_message = "Datastore should reference the S3 endpoint"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_bucket == local.bucket.bucket
    error_message = "Datastore should reference the B2 bucket"
  }
}

run "verify_backblaze_no_drift" {
  command = plan
  
  assert {
    condition     = pbs_datastore.test.name == "b2-ds-${var.test_id}"
    error_message = "Datastore should not have drifted"
  }
}

run "test_backblaze_compatibility" {
  command = apply
  
  # Verify that B2-specific settings persist
  assert {
    condition     = contains(pbs_s3_endpoint.test.provider_quirks, "skip-if-none-match-header")
    error_message = "Backblaze quirks should persist"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_bucket == "pbs-test-b2-${var.test_id}"
    error_message = "Datastore should still reference correct B2 bucket"
  }
}
