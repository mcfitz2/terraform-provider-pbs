provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "create_influxdb_http_server" {
  variables {
    server_name  = "tftest-influxdb-http"
    organization = "testorg"
    bucket       = "pbs-metrics"
    token        = "test-token-123456"
    comment      = "Test InfluxDB HTTP metrics server"
  }

  assert {
    condition     = pbs_metrics_server.test.name == var.server_name
    error_message = "Server name mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.type == "influxdb-http"
    error_message = "Server type mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.url == "http://${var.influxdb_host}:${var.influxdb_port}"
    error_message = "URL mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.organization == "testorg"
    error_message = "Organization mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.bucket == "pbs-metrics"
    error_message = "Bucket mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.enable == true
    error_message = "Enable should be true"
  }

  assert {
    condition     = pbs_metrics_server.test.comment == "Test InfluxDB HTTP metrics server"
    error_message = "Comment mismatch"
  }
}

run "update_influxdb_http_server" {
  variables {
    server_name  = "tftest-influxdb-http"
    organization = "testorg"
    bucket       = "pbs-metrics"
    token        = "test-token-123456"
    comment      = "Updated InfluxDB HTTP metrics server"
  }

  assert {
    condition     = pbs_metrics_server.test.comment == "Updated InfluxDB HTTP metrics server"
    error_message = "Comment was not updated"
  }

  assert {
    condition     = pbs_metrics_server.test.name == var.server_name
    error_message = "Server name should remain unchanged"
  }

  assert {
    condition     = pbs_metrics_server.test.organization == "testorg"
    error_message = "Organization should remain unchanged"
  }
}
