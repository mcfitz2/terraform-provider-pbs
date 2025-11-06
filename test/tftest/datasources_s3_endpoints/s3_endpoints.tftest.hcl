run "list_s3_endpoints" {
  variables {
  }

  assert {
    condition     = length(data.pbs_s3_endpoints.all.endpoints) >= 2
    error_message = "Should have at least 2 S3 endpoints"
  }

  assert {
    condition     = contains([for e in data.pbs_s3_endpoints.all.endpoints : e.id], "tftest-s3-list-1")
    error_message = "Should contain test S3 endpoint 1"
  }

  assert {
    condition     = contains([for e in data.pbs_s3_endpoints.all.endpoints : e.id], "tftest-s3-list-2")
    error_message = "Should contain test S3 endpoint 2"
  }

  assert {
    condition     = output.endpoint_count >= 2
    error_message = "Endpoint count output mismatch"
  }
}
