provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "test_notification_matcher_data_source" {
  variables {
    pbs_endpoint = var.pbs_endpoint
    pbs_username = var.pbs_username
    pbs_password = var.pbs_password
  }

  assert {
    condition     = data.pbs_notification_matcher.test.name == pbs_notification_matcher.test.name
    error_message = "Data source name should match resource name"
  }

  assert {
    condition     = data.pbs_notification_matcher.test.comment == "Integration test matcher for data source"
    error_message = "Data source comment should match resource comment"
  }

  assert {
    condition     = data.pbs_notification_matcher.test.mode == "all"
    error_message = "Data source mode should match resource mode"
  }

  assert {
    condition     = length(data.pbs_notification_matcher.test.targets) == 1
    error_message = "Data source should have one target"
  }

  assert {
    condition     = contains(data.pbs_notification_matcher.test.targets, pbs_smtp_notification.target.name)
    error_message = "Data source targets should include the SMTP notification"
  }

  assert {
    condition     = length(data.pbs_notification_matcher.test.match_severity) == 2
    error_message = "Data source should have two severity levels"
  }

  assert {
    condition     = contains(data.pbs_notification_matcher.test.match_severity, "error")
    error_message = "Data source match_severity should include 'error'"
  }

  assert {
    condition     = contains(data.pbs_notification_matcher.test.match_severity, "warning")
    error_message = "Data source match_severity should include 'warning'"
  }
}
