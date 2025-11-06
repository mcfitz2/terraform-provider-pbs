provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "create_gotify_notification" {
  variables {
    gotify_name   = "tftest-gotify"
    gotify_server = "https://gotify.example.com"
    gotify_token  = "Aabcd1234567890"
    
    # Other required vars with dummy values
    sendmail_name   = "dummy-sendmail"
    sendmail_mailto = ["dummy@example.com"]
    sendmail_from   = "dummy@example.com"
    webhook_name    = "dummy-webhook"
    webhook_url     = "https://dummy.example.com"
    matcher_name    = "dummy-matcher"
  }

  assert {
    condition     = pbs_gotify_notification.test.name == "tftest-gotify"
    error_message = "Gotify name mismatch"
  }

  assert {
    condition     = pbs_gotify_notification.test.server == "https://gotify.example.com"
    error_message = "Gotify server mismatch"
  }

  assert {
    condition     = pbs_gotify_notification.test.disable == false
    error_message = "Gotify should be enabled"
  }
}

run "create_sendmail_notification" {
  variables {
    sendmail_name   = "tftest-sendmail"
    sendmail_mailto = ["admin@example.com"]
    sendmail_from   = "pbs@example.com"
    
    # Other required vars
    gotify_name   = "tftest-gotify"
    gotify_server = "https://gotify.example.com"
    gotify_token  = "Aabcd1234567890"
    webhook_name  = "dummy-webhook"
    webhook_url   = "https://dummy.example.com"
    matcher_name  = "dummy-matcher"
  }

  assert {
    condition     = pbs_sendmail_notification.test.name == "tftest-sendmail"
    error_message = "Sendmail name mismatch"
  }

  assert {
    condition     = pbs_sendmail_notification.test.from_address == "pbs@example.com"
    error_message = "Sendmail from address mismatch"
  }

  assert {
    condition     = length(pbs_sendmail_notification.test.mailto) == 1
    error_message = "Sendmail mailto list length mismatch"
  }

  assert {
    condition     = contains(pbs_sendmail_notification.test.mailto, "admin@example.com")
    error_message = "Sendmail mailto missing admin email"
  }
}

run "create_webhook_notification" {
  variables {
    webhook_name   = "tftest-webhook"
    webhook_url    = "https://webhook.example.com/notify"
    webhook_method = "post"
    
    # Other required vars
    gotify_name     = "tftest-gotify"
    gotify_server   = "https://gotify.example.com"
    gotify_token    = "Aabcd1234567890"
    sendmail_name   = "tftest-sendmail"
    sendmail_mailto = ["admin@example.com"]
    sendmail_from   = "pbs@example.com"
    matcher_name    = "dummy-matcher"
  }

  assert {
    condition     = pbs_webhook_notification.test.name == "tftest-webhook"
    error_message = "Webhook name mismatch"
  }

  assert {
    condition     = pbs_webhook_notification.test.url == "https://webhook.example.com/notify"
    error_message = "Webhook URL mismatch"
  }

  assert {
    condition     = pbs_webhook_notification.test.method == "post"
    error_message = "Webhook method mismatch"
  }
}

run "create_notification_matcher_all_mode" {
  variables {
    matcher_name    = "tftest-matcher-all"
    matcher_mode    = "all"
    match_severity  = ["error", "warning"]
    
    # Endpoints
    gotify_name     = "tftest-gotify"
    gotify_server   = "https://gotify.example.com"
    gotify_token    = "Aabcd1234567890"
    sendmail_name   = "tftest-sendmail"
    sendmail_mailto = ["admin@example.com"]
    sendmail_from   = "pbs@example.com"
    webhook_name    = "tftest-webhook"
    webhook_url     = "https://webhook.example.com/notify"
  }

  assert {
    condition     = pbs_notification_matcher.test.name == "tftest-matcher-all"
    error_message = "Matcher name mismatch"
  }

  assert {
    condition     = pbs_notification_matcher.test.mode == "all"
    error_message = "Matcher mode should be 'all'"
  }

  assert {
    condition     = length(pbs_notification_matcher.test.match_severity) == 2
    error_message = "Matcher should have 2 severity filters"
  }

  assert {
    condition     = contains(pbs_notification_matcher.test.match_severity, "error")
    error_message = "Matcher should include 'error' severity"
  }

  assert {
    condition     = length(pbs_notification_matcher.test.targets) >= 1
    error_message = "Matcher should have at least 1 target"
  }
}

run "update_matcher_with_calendar_and_any_mode" {
  variables {
    matcher_name    = "tftest-matcher-all"
    matcher_mode    = "any"
    match_severity  = ["info", "notice"]
    match_calendar  = ["Mon..Fri 08:00-17:00"]
    
    # Endpoints
    gotify_name     = "tftest-gotify"
    gotify_server   = "https://gotify.example.com"
    gotify_token    = "Aabcd1234567890"
    sendmail_name   = "tftest-sendmail"
    sendmail_mailto = ["admin@example.com"]
    sendmail_from   = "pbs@example.com"
    webhook_name    = "tftest-webhook"
    webhook_url     = "https://webhook.example.com/notify"
  }

  assert {
    condition     = pbs_notification_matcher.test.mode == "any"
    error_message = "Matcher mode should be updated to 'any'"
  }

  assert {
    condition     = length(pbs_notification_matcher.test.match_calendar) == 1
    error_message = "Matcher should have 1 calendar filter"
  }

  assert {
    condition     = contains(pbs_notification_matcher.test.match_calendar, "Mon..Fri 08:00-17:00")
    error_message = "Matcher calendar filter mismatch"
  }
}

run "test_matcher_with_invert_match" {
  variables {
    matcher_name    = "tftest-matcher-all"
    matcher_mode    = "all"
    match_severity  = ["error"]
    invert_match    = true
    
    # Endpoints
    gotify_name     = "tftest-gotify"
    gotify_server   = "https://gotify.example.com"
    gotify_token    = "Aabcd1234567890"
    sendmail_name   = "tftest-sendmail"
    sendmail_mailto = ["admin@example.com"]
    sendmail_from   = "pbs@example.com"
    webhook_name    = "tftest-webhook"
    webhook_url     = "https://webhook.example.com/notify"
  }

  assert {
    condition     = pbs_notification_matcher.test.invert_match == true
    error_message = "Matcher invert_match should be true"
  }

  assert {
    condition     = pbs_notification_matcher.test.mode == "all"
    error_message = "Matcher mode mismatch"
  }
}
