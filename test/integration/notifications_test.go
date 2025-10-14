package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// TestSMTPNotificationIntegration tests SMTP notification target lifecycle
func TestSMTPNotificationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	targetName := GenerateTestName("smtp-target")

	testConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "test_smtp" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  mode         = "insecure"
  username     = "test@example.com"
  password     = "secret123"
  mailto       = "admin@example.com,backup@example.com"
  from_address = "pbs@example.com"
  author       = "PBS Admin"
  comment      = "Test SMTP notification"
  disable      = false
}
`, targetName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_smtp_notification.test_smtp")
	assert.Equal(t, targetName, resource.AttributeValues["name"])
	assert.Equal(t, "smtp.example.com", resource.AttributeValues["server"])
	assert.Equal(t, float64(587), resource.AttributeValues["port"])
	assert.Equal(t, "test@example.com", resource.AttributeValues["username"])

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetSMTPTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "smtp.example.com", target.Server)
	assert.Equal(t, 587, target.Port)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "test_smtp" {
  name         = "%s"
  server       = "smtp.newserver.com"
  port         = 465
  username     = "updated@example.com"
  password     = "newsecret456"
  mailto       = "newadmin@example.com"
  from_address = "pbs-updated@example.com"
  author       = "Updated PBS Admin"
  comment      = "Updated SMTP notification"
  disable      = false
}
`, targetName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	target, err = notifClient.GetSMTPTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, "smtp.newserver.com", target.Server)
	assert.Equal(t, 465, target.Port)
	assert.Equal(t, "Updated SMTP notification", target.Comment)
}

// TestGotifyNotificationIntegration tests Gotify notification target lifecycle
func TestGotifyNotificationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	targetName := GenerateTestName("gotify-target")

	testConfig := fmt.Sprintf(`
resource "pbs_gotify_notification" "test_gotify" {
  name    = "%s"
  server  = "https://gotify.example.com"
  token   = "Aabcd1234567890" # gitleaks:allow
  comment = "Test Gotify notification"
  disable = false
}
`, targetName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_gotify_notification.test_gotify")
	assert.Equal(t, targetName, resource.AttributeValues["name"])
	assert.Equal(t, "https://gotify.example.com", resource.AttributeValues["server"])

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetGotifyTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "https://gotify.example.com", target.Server)
}

// TestSendmailNotificationIntegration tests Sendmail notification target lifecycle
func TestSendmailNotificationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	targetName := GenerateTestName("sendmail-target")

	testConfig := fmt.Sprintf(`
resource "pbs_sendmail_notification" "test_sendmail" {
  name         = "%s"
  mailto       = "admin@example.com"
  from_address = "pbs@example.com"
  author       = "PBS System"
  comment      = "Test Sendmail notification"
  disable      = false
}
`, targetName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_sendmail_notification.test_sendmail")
	assert.Equal(t, targetName, resource.AttributeValues["name"])
	assert.Equal(t, "pbs@example.com", resource.AttributeValues["from_address"])

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetSendmailTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "pbs@example.com", target.From)
}

// TestWebhookNotificationIntegration tests Webhook notification target lifecycle
func TestWebhookNotificationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	targetName := GenerateTestName("webhook-target")

	testConfig := fmt.Sprintf(`
resource "pbs_webhook_notification" "test_webhook" {
  name    = "%s"
  url     = "https://webhook.example.com/notify"
  method  = "POST"
  comment = "Test Webhook notification"
  disable = false
}
`, targetName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_webhook_notification.test_webhook")
	assert.Equal(t, targetName, resource.AttributeValues["name"])
	assert.Equal(t, "https://webhook.example.com/notify", resource.AttributeValues["url"])
	// Method is normalized to lowercase in our provider
	method := resource.AttributeValues["method"].(string)
	assert.True(t, method == "post" || method == "POST", "method should be 'post' or 'POST', got: %s", method)

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetWebhookTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "https://webhook.example.com/notify", target.URL)
	// API may return uppercase or lowercase
	assert.True(t, target.Method == "post" || target.Method == "POST", "method should be 'post' or 'POST', got: %s", target.Method)
}

// TestNotificationMatcherIntegration tests notification matcher lifecycle
func TestNotificationMatcherIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	matcherName := GenerateTestName("matcher")
	smtpTarget := GenerateTestName("smtp")

	testConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "target" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = "admin@example.com"
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test_matcher" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["error", "warning"]
  match_field    = ["type=prune", "datastore=backup"]
  mode           = "all"
  invert_match   = false
  comment        = "Test notification matcher"
  disable        = false
}
`, smtpTarget, matcherName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_notification_matcher.test_matcher")
	assert.Equal(t, matcherName, resource.AttributeValues["name"])
	assert.Equal(t, "all", resource.AttributeValues["mode"])
	assert.NotNil(t, resource.AttributeValues["match_severity"])

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	matcher, err := notifClient.GetNotificationMatcher(context.Background(), matcherName)
	require.NoError(t, err)
	assert.Equal(t, matcherName, matcher.Name)
	assert.Len(t, matcher.Targets, 1)
	assert.Contains(t, matcher.Targets, smtpTarget)
	assert.Len(t, matcher.MatchSeverity, 2)
	assert.Contains(t, matcher.MatchSeverity, "error")
	assert.Contains(t, matcher.MatchSeverity, "warning")
	assert.Equal(t, "all", matcher.Mode)
}

// TestNotificationMatcherModes tests different matcher modes
func TestNotificationMatcherModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	matcherName := GenerateTestName("matcher-any")
	smtpTarget := GenerateTestName("smtp")

	// Test "any" mode
	testConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "target" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = "admin@example.com"
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test_any" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["info", "notice"]
  mode           = "any"
  comment        = "Matcher with any mode"
}
`, smtpTarget, matcherName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_notification_matcher.test_any")
	assert.Equal(t, "any", resource.AttributeValues["mode"])

	notifClient := notifications.NewClient(tc.APIClient)
	matcher, err := notifClient.GetNotificationMatcher(context.Background(), matcherName)
	require.NoError(t, err)
	assert.Equal(t, "any", matcher.Mode)
}

// TestNotificationMatcherWithCalendar tests matcher with calendar filter
func TestNotificationMatcherWithCalendar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	matcherName := GenerateTestName("matcher-calendar")
	smtpTarget := GenerateTestName("smtp")

	testConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "target" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = "admin@example.com"
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test_calendar" {
  name            = "%s"
  targets         = [pbs_smtp_notification.target.name]
  match_severity  = ["error"]
  match_calendar  = ["Mon..Fri 08:00-17:00"]
  mode            = "all"
  comment         = "Matcher with calendar - business hours only"
}
`, smtpTarget, matcherName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_notification_matcher.test_calendar")
	assert.Equal(t, []interface{}{"Mon..Fri 08:00-17:00"}, resource.AttributeValues["match_calendar"])

	notifClient := notifications.NewClient(tc.APIClient)
	matcher, err := notifClient.GetNotificationMatcher(context.Background(), matcherName)
	require.NoError(t, err)
	assert.Equal(t, "Mon..Fri 08:00-17:00", matcher.MatchCalendar)
}

// TestNotificationMatcherInvertMatch tests matcher with inverted matching
func TestNotificationMatcherInvertMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	matcherName := GenerateTestName("matcher-invert")
	smtpTarget := GenerateTestName("smtp")

	testConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "target" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = "admin@example.com"
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test_invert" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["info"]
  mode           = "all"
  invert_match   = true
  comment        = "Matcher with inverted match - notify on anything except info"
}
`, smtpTarget, matcherName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_notification_matcher.test_invert")
	assert.Equal(t, true, resource.AttributeValues["invert_match"])

	notifClient := notifications.NewClient(tc.APIClient)
	matcher, err := notifClient.GetNotificationMatcher(context.Background(), matcherName)
	require.NoError(t, err)
	assert.NotNil(t, matcher.InvertMatch)
	assert.True(t, *matcher.InvertMatch)
}
