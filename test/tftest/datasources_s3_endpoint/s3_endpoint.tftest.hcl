run "read_s3_endpoint_datasource" {
  variables {
  }

  assert {
    condition     = data.pbs_s3_endpoint.test.name == pbs_s3_endpoint.test.name
    error_message = "Data source name doesn't match resource"
  }

  assert {
    condition     = data.pbs_s3_endpoint.test.provider == pbs_s3_endpoint.test.provider
    error_message = "Data source provider doesn't match resource"
  }

  assert {
    condition     = data.pbs_s3_endpoint.test.endpoint == pbs_s3_endpoint.test.endpoint
    error_message = "Data source endpoint doesn't match resource"
  }

  assert {
    condition     = data.pbs_s3_endpoint.test.region == pbs_s3_endpoint.test.region
    error_message = "Data source region doesn't match resource"
  }

  assert {
    condition     = data.pbs_s3_endpoint.test.comment == pbs_s3_endpoint.test.comment
    error_message = "Data source comment doesn't match resource"
  }

  assert {
    condition     = output.datasource_name == "tftest-s3-ds"
    error_message = "Output name mismatch"
  }

  assert {
    condition     = output.datasource_provider == "aws"
    error_message = "Output provider mismatch"
  }
}
