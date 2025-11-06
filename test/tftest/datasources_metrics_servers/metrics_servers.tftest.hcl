provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "test_metrics_servers_data_source" {
  variables {
    pbs_endpoint      = var.pbs_endpoint
    pbs_username      = var.pbs_username
    pbs_password      = var.pbs_password
    influxdb_host     = var.influxdb_host
    influxdb_port     = var.influxdb_port
    influxdb_udp_host = var.influxdb_udp_host
    influxdb_udp_port = var.influxdb_udp_port
  }

  assert {
    condition     = length(data.pbs_metrics_servers.all.servers) >= 2
    error_message = "Data source should return at least 2 metrics servers"
  }

  assert {
    condition     = contains([for s in data.pbs_metrics_servers.all.servers : s.name], pbs_metrics_server.test1.name)
    error_message = "Data source should include test server 1"
  }

  assert {
    condition     = contains([for s in data.pbs_metrics_servers.all.servers : s.name], pbs_metrics_server.test2.name)
    error_message = "Data source should include test server 2"
  }

  assert {
    condition     = length([for s in data.pbs_metrics_servers.all.servers : s if s.type == "influxdb-http"]) >= 1
    error_message = "Data source should include at least one influxdb-http server"
  }

  assert {
    condition     = length([for s in data.pbs_metrics_servers.all.servers : s if s.type == "influxdb-udp"]) >= 1
    error_message = "Data source should include at least one influxdb-udp server"
  }
}
