provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "test_notification_endpoint_data_source" {
  variables {
    pbs_endpoint = var.pbs_endpoint
    pbs_username = var.pbs_username
    pbs_password = var.pbs_password
  }

  assert {
    condition     = data.pbs_notification_endpoint.test.name == pbs_gotify_notification.test.name
    error_message = "Data source name should match resource name"
  }

  assert {
    condition     = data.pbs_notification_endpoint.test.type == "gotify"
    error_message = "Data source type should be gotify"
  }

  assert {
    condition     = data.pbs_notification_endpoint.test.url == pbs_gotify_notification.test.server
    error_message = "Data source URL should match resource server"
  }

  assert {
    condition     = data.pbs_notification_endpoint.test.comment == "Integration test gotify endpoint for data source"
    error_message = "Data source comment should match resource comment"
  }

  assert {
    condition     = data.pbs_notification_endpoint.test.disable == false
    error_message = "Data source disable should match resource disable"
  }
}
