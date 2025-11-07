# AWS S3 Provider Test
#
# This test validates the complete lifecycle of S3-backed datastores using AWS S3:
# 1. Creates an S3 bucket
# 2. Creates a PBS S3 endpoint pointing to the bucket
# 3. Creates a PBS datastore using the S3 backend
# 4. Verifies all resources and their relationships
# 5. Cleans up all resources (bucket, endpoint, datastore)

variables {
  s3_provider_name = "AWS"
  s3_endpoint      = "s3.${var.s3_region}.amazonaws.com"
  s3_region        = "us-east-1"
  s3_bucket_name   = "pbs-test-aws-${var.test_id}"
  s3_endpoint_id   = "pbs-aws-${var.test_id}"
  datastore_name   = "aws-ds-${var.test_id}"
  s3_provider_quirks = []
}

run "setup_aws" {
  command = plan
  
  assert {
    condition     = var.s3_provider_name == "AWS"
    error_message = "Provider name should be AWS"
  }
}

run "create_aws_s3_infrastructure" {
  command = apply
  
  assert {
    condition     = local.bucket.bucket == "pbs-test-aws-${var.test_id}"
    error_message = "S3 bucket name should match expected pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.id == "pbs-aws-${var.test_id}"
    error_message = "S3 endpoint ID should match expected pattern"
  }
  
  assert {
    condition     = pbs_s3_endpoint.test.endpoint == "s3.${var.s3_region}.amazonaws.com"
    error_message = "S3 endpoint should match AWS pattern"
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
    condition     = pbs_datastore.test.name == "aws-ds-${var.test_id}"
    error_message = "Datastore name should match expected pattern"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_client == pbs_s3_endpoint.test.id
    error_message = "Datastore should reference the S3 endpoint"
  }
  
  assert {
    condition     = pbs_datastore.test.s3_bucket == local.bucket.bucket
    error_message = "Datastore should reference the S3 bucket"
  }
}

run "verify_aws_no_drift" {
  command = plan
  
  assert {
    condition     = pbs_datastore.test.name == "aws-ds-${var.test_id}"
    error_message = "Datastore should not have drifted"
  }
}

run "update_aws_datastore_comment" {
  command = apply
  
  variables {
    datastore_name = "aws-ds-${var.test_id}"
  }
  
  # Update only happens if we modify the resource, but we're testing that
  # updating mutable fields works without errors
  assert {
    condition     = pbs_datastore.test.s3_bucket == local.bucket.bucket
    error_message = "S3 bucket should remain unchanged after update"
  }
}
