provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "test_notification_endpoints_data_source" {
  variables {
    pbs_endpoint = var.pbs_endpoint
    pbs_username = var.pbs_username
    pbs_password = var.pbs_password
  }

  assert {
    condition     = length(data.pbs_notification_endpoints.all.endpoints) >= 2
    error_message = "Data source should return at least 2 notification endpoints"
  }

  assert {
    condition     = contains([for e in data.pbs_notification_endpoints.all.endpoints : e.name], pbs_gotify_notification.test1.name)
    error_message = "Data source should include gotify test endpoint"
  }

  assert {
    condition     = contains([for e in data.pbs_notification_endpoints.all.endpoints : e.name], pbs_smtp_notification.test2.name)
    error_message = "Data source should include smtp test endpoint"
  }

  assert {
    condition     = length([for e in data.pbs_notification_endpoints.all.endpoints : e if e.type == "gotify"]) >= 1
    error_message = "Data source should include at least one gotify endpoint"
  }

  assert {
    condition     = length([for e in data.pbs_notification_endpoints.all.endpoints : e if e.type == "smtp"]) >= 1
    error_message = "Data source should include at least one smtp endpoint"
  }
}
