# Scaleway Object Storage S3-Compatible Provider Test
#
# This test validates the complete lifecycle of S3-backed datastores using Scaleway:
# 1. Creates a Scaleway Object Storage bucket via S3 API
# 2. Creates a PBS S3 endpoint configured for Scaleway
# 3. Creates a PBS datastore using the Scaleway backend
# 4. Verifies all resources work with Scaleway's S3 compatibility
# 5. Cleans up all resources
#
# Scaleway Requirements:
# - path_style = true (recommended for PBS)
# - Endpoint format: s3.{region}.scw.cloud

variables {
  s3_provider_name = "Scaleway"
  s3_endpoint      = "s3.${var.pbs_s3_region}.scw.cloud"
  s3_region        = "us-east-1"  # Generic region for AWS provider
  pbs_s3_region    = "fr-par"  # Actual Scaleway region for PBS endpoint
  s3_bucket_name   = "pbs-test-scw-${var.test_id}"
  s3_endpoint_id   = "pbs-scw-${var.test_id}"
  datastore_name   = "scw-ds-${var.test_id}"
  s3_provider_quirks = []
}

run "setup_scaleway" {
  command = plan
  
  assert {
    condition     = var.s3_provider_name == "Scaleway"
    error_message = "Provider name should be Scaleway"
  }
}

run "create_scaleway_s3_infrastructure" {
  command = apply
  
  assert {
    condition     = local.bucket.bucket == "pbs-test-scw-${var.test_id}"
    error_message = "Scaleway bucket name should match expected pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.id == "pbs-scw-${var.test_id}"
    error_message = "S3 endpoint ID should match expected pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.endpoint == "s3.${var.s3_region}.scw.cloud"
    error_message = "S3 endpoint should match Scaleway pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.region == var.s3_region
    error_message = "S3 endpoint region should match configured region"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.path_style == true
    error_message = "Path style should be enabled for PBS compatibility"
  }
  
  assert {
    condition     = pbs_datastore.test.name == "scw-ds-${var.test_id}"
    error_message = "Datastore name should match expected pattern"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_client == pbs_s3_endpoint.test.id
    error_message = "Datastore should reference the S3 endpoint"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_bucket == local.bucket.bucket
    error_message = "Datastore should reference the Scaleway bucket"
  }
}

run "verify_scaleway_no_drift" {
  command = plan
  
  assert {
    condition     = pbs_datastore.test.name == "scw-ds-${var.test_id}"
    error_message = "Datastore should not have drifted"
  }
}

run "update_scaleway_datastore_mutable" {
  command = apply
  
  # Verify immutable fields remain unchanged
  assert {
    condition     = pbs_datastore.test.s3_bucket == local.bucket.bucket
    error_message = "S3 bucket should remain unchanged"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_client == pbs_s3_endpoint.test.id
    error_message = "S3 client should remain unchanged"
  }
}
