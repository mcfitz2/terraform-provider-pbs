provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "create_smtp_notification" {
  variables {
    name          = "tftest-smtp"
    server        = "smtp.example.com"
    port          = 587
    mode          = "insecure"
    username      = "test@example.com"
    password_smtp = "secret123"
    mailto        = ["admin@example.com", "backup@example.com"]
    from_address  = "pbs@example.com"
    author        = "PBS Admin"
    comment       = "Test SMTP notification"
  }

  assert {
    condition     = pbs_smtp_notification.test.name == var.name
    error_message = "Name mismatch"
  }

  assert {
    condition     = pbs_smtp_notification.test.server == "smtp.example.com"
    error_message = "Server mismatch"
  }

  assert {
    condition     = pbs_smtp_notification.test.port == 587
    error_message = "Port mismatch"
  }

  assert {
    condition     = pbs_smtp_notification.test.username == "test@example.com"
    error_message = "Username mismatch"
  }

  assert {
    condition     = length(pbs_smtp_notification.test.mailto) == 2
    error_message = "Mailto list length mismatch"
  }

  assert {
    condition     = contains(pbs_smtp_notification.test.mailto, "admin@example.com")
    error_message = "Missing admin email in mailto"
  }

  assert {
    condition     = contains(pbs_smtp_notification.test.mailto, "backup@example.com")
    error_message = "Missing backup email in mailto"
  }
}

run "update_smtp_notification" {
  variables {
    name          = "tftest-smtp"
    server        = "smtp.newserver.com"
    port          = 465
    username      = "updated@example.com"
    password_smtp = "newsecret456"
    mailto        = ["newadmin@example.com"]
    from_address  = "pbs-updated@example.com"
    author        = "Updated PBS Admin"
    comment       = "Updated SMTP notification"
  }

  assert {
    condition     = pbs_smtp_notification.test.server == "smtp.newserver.com"
    error_message = "Server was not updated"
  }

  assert {
    condition     = pbs_smtp_notification.test.port == 465
    error_message = "Port was not updated"
  }

  assert {
    condition     = length(pbs_smtp_notification.test.mailto) == 1
    error_message = "Mailto list should have 1 entry"
  }

  assert {
    condition     = pbs_smtp_notification.test.comment == "Updated SMTP notification"
    error_message = "Comment was not updated"
  }
}
