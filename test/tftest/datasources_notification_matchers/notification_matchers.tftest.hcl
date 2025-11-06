provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "test_notification_matchers_data_source" {
  variables {
    pbs_endpoint = var.pbs_endpoint
    pbs_username = var.pbs_username
    pbs_password = var.pbs_password
  }

  assert {
    condition     = length(data.pbs_notification_matchers.all.matchers) >= 2
    error_message = "Data source should return at least 2 notification matchers"
  }

  assert {
    condition     = contains([for m in data.pbs_notification_matchers.all.matchers : m.name], pbs_notification_matcher.test1.name)
    error_message = "Data source should include matcher 1"
  }

  assert {
    condition     = contains([for m in data.pbs_notification_matchers.all.matchers : m.name], pbs_notification_matcher.test2.name)
    error_message = "Data source should include matcher 2"
  }

  assert {
    condition     = length([for m in data.pbs_notification_matchers.all.matchers : m if m.mode == "all"]) >= 1
    error_message = "Data source should include at least one matcher with mode 'all'"
  }

  assert {
    condition     = length([for m in data.pbs_notification_matchers.all.matchers : m if m.mode == "any"]) >= 1
    error_message = "Data source should include at least one matcher with mode 'any'"
  }
}
