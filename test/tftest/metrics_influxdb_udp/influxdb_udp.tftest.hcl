provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "create_influxdb_udp_server" {
  variables {
    server_name = "tftest-influxdb-udp"
    comment     = "Test InfluxDB UDP metrics server"
  }

  assert {
    condition     = pbs_metrics_server.test.name == var.server_name
    error_message = "Server name mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.type == "influxdb-udp"
    error_message = "Server type mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.server == var.influxdb_udp_host
    error_message = "Server host mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.port == var.influxdb_udp_port
    error_message = "Port mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.protocol == "udp"
    error_message = "Protocol mismatch"
  }

  assert {
    condition     = pbs_metrics_server.test.enable == true
    error_message = "Enable should be true"
  }
}

run "update_with_mtu" {
  variables {
    server_name = "tftest-influxdb-udp"
    mtu         = 1400
    comment     = "Updated with MTU setting"
  }

  assert {
    condition     = pbs_metrics_server.test.mtu == 1400
    error_message = "MTU was not updated"
  }

  assert {
    condition     = pbs_metrics_server.test.comment == "Updated with MTU setting"
    error_message = "Comment was not updated"
  }
}

run "disable_server" {
  variables {
    server_name = "tftest-influxdb-udp"
    mtu         = 1400
    enable      = false
    comment     = "Server disabled"
  }

  assert {
    condition     = pbs_metrics_server.test.enable == false
    error_message = "Server should be disabled"
  }

  assert {
    condition     = pbs_metrics_server.test.comment == "Server disabled"
    error_message = "Comment mismatch"
  }
}
