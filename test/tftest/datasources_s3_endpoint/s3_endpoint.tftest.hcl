run "read_s3_endpoint_datasource" {
  variables {
  }

  assert {
    condition     = data.pbs_s3_endpoint.test.id == pbs_s3_endpoint.test.id
    error_message = "Data source id doesn't match resource"
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
    condition     = output.datasource_id == "tftest-s3-ds"
    error_message = "Output id mismatch"
  }

  assert {
    condition     = output.datasource_endpoint == "https://s3.amazonaws.com"
    error_message = "Output endpoint mismatch"
  }
}
