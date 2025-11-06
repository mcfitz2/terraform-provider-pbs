provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "test_metrics_server_data_source" {
  variables {
    pbs_endpoint = var.pbs_endpoint
    pbs_username = var.pbs_username
    pbs_password = var.pbs_password
    influxdb_host = var.influxdb_host
    influxdb_port = var.influxdb_port
  }

  assert {
    condition     = data.pbs_metrics_server.test.name == pbs_metrics_server.test.name
    error_message = "Data source name should match resource name"
  }

  assert {
    condition     = data.pbs_metrics_server.test.type == "influxdb-http"
    error_message = "Data source type should be influxdb-http"
  }

  assert {
    condition     = data.pbs_metrics_server.test.url == pbs_metrics_server.test.url
    error_message = "Data source URL should match resource URL"
  }

  assert {
    condition     = data.pbs_metrics_server.test.organization == "test-org"
    error_message = "Data source organization should match resource organization"
  }

  assert {
    condition     = data.pbs_metrics_server.test.bucket == "test-bucket"
    error_message = "Data source bucket should match resource bucket"
  }

  assert {
    condition     = data.pbs_metrics_server.test.comment == "Integration test metrics server"
    error_message = "Data source comment should match resource comment"
  }

  assert {
    condition     = data.pbs_metrics_server.test.enable == true
    error_message = "Data source enable should match resource enable"
  }
}
