package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// TestSMTPNotificationIntegration tests SMTP notification target lifecycle
func TestSMTPNotificationIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/smtp.tftest.hcl")

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
	mailto       = ["admin@example.com", "backup@example.com"]
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
	// Port is stored as json.Number in Terraform state
	portVal := resource.AttributeValues["port"]
	if portStr, ok := portVal.(string); ok {
		assert.Equal(t, "587", portStr)
	} else {
		// Handle json.Number type
		assert.Equal(t, "587", fmt.Sprint(portVal))
	}
	assert.Equal(t, "test@example.com", resource.AttributeValues["username"])
	mailtoState, ok := resource.AttributeValues["mailto"].([]interface{})
	require.True(t, ok, "expected mailto to be a list in state")
	assert.ElementsMatch(t, []interface{}{"admin@example.com", "backup@example.com"}, mailtoState)

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetSMTPTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "smtp.example.com", target.Server)
	assert.Equal(t, 587, *target.Port)
	assert.ElementsMatch(t, []string{"admin@example.com", "backup@example.com"}, target.To)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_smtp_notification" "test_smtp" {
  name         = "%s"
  server       = "smtp.newserver.com"
  port         = 465
  username     = "updated@example.com"
  password     = "newsecret456"
	mailto       = ["newadmin@example.com"]
  from_address = "pbs-updated@example.com"
  author       = "Updated PBS Admin"
  comment      = "Updated SMTP notification"
  disable      = false
}
`, targetName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	resource = tc.GetResourceFromState(t, "pbs_smtp_notification.test_smtp")
	mailtoState, ok = resource.AttributeValues["mailto"].([]interface{})
	require.True(t, ok, "expected mailto to be a list in state after update")
	assert.ElementsMatch(t, []interface{}{"newadmin@example.com"}, mailtoState)

	target, err = notifClient.GetSMTPTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, "smtp.newserver.com", target.Server)
	assert.Equal(t, 465, *target.Port)
	assert.Equal(t, "Updated SMTP notification", target.Comment)
	assert.ElementsMatch(t, []string{"newadmin@example.com"}, target.To)
}

// TestGotifyNotificationIntegration tests Gotify notification target lifecycle
func TestGotifyNotificationIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl")

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
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	targetName := GenerateTestName("sendmail-target")

	testConfig := fmt.Sprintf(`
resource "pbs_sendmail_notification" "test_sendmail" {
	name         = "%s"
	mailto       = ["admin@example.com"]
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
	mailtoState, ok := resource.AttributeValues["mailto"].([]interface{})
	require.True(t, ok, "expected mailto to be a list in state")
	assert.ElementsMatch(t, []interface{}{"admin@example.com"}, mailtoState)

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetSendmailTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "pbs@example.com", target.From)
	assert.ElementsMatch(t, []string{"admin@example.com"}, target.Mailto)
}

// TestWebhookNotificationIntegration tests Webhook notification target lifecycle
func TestWebhookNotificationIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl")

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
	method  = "post"
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
	assert.Equal(t, "post", method)

	// Verify via API
	notifClient := notifications.NewClient(tc.APIClient)
	target, err := notifClient.GetWebhookTarget(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, targetName, target.Name)
	assert.Equal(t, "https://webhook.example.com/notify", target.URL)
	// API may return uppercase or lowercase
	assert.True(t, strings.EqualFold(target.Method, "post"), "method should be 'post' (any case), got: %s", target.Method)
}

// TestNotificationMatcherIntegration tests notification matcher lifecycle
func TestNotificationMatcherIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl")

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
	mailto       = ["admin@example.com"]
	from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test_matcher" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["error", "warning"]
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
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl (merged)")

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
	mailto       = ["admin@example.com"]
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
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl (merged)")

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
	mailto       = ["admin@example.com"]
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
	// MatchCalendar is an array in the API
	assert.Equal(t, []string{"Mon..Fri 08:00-17:00"}, matcher.MatchCalendar)
}

// TestNotificationMatcherInvertMatch tests matcher with inverted matching
func TestNotificationMatcherInvertMatch(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/notifications/endpoints_and_matcher.tftest.hcl (merged)")

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
	mailto       = ["admin@example.com"]
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

// Data Source Integration Tests

// TestNotificationEndpointDataSourceIntegration tests reading a single notification endpoint via data source
func TestNotificationEndpointDataSourceIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/datasources/notification_endpoint.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Use Gotify for the test since it's simpler
	targetName := GenerateTestName("gotify-ds")

	config := fmt.Sprintf(`
resource "pbs_gotify_notification" "test" {
  name    = "%s"
  server  = "https://gotify.example.com"
  token   = "Aabcd1234567890" # gitleaks:allow
  comment = "Integration test gotify endpoint for data source"
  disable = false
}

data "pbs_notification_endpoint" "test" {
  name = pbs_gotify_notification.test.name
  type = "gotify"
}
`, targetName)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_notification_endpoint.test")
	resource := tc.GetResourceFromState(t, "pbs_gotify_notification.test")

	assert.Equal(t, resource.AttributeValues["name"], dataSource.AttributeValues["name"])
	assert.Equal(t, "gotify", dataSource.AttributeValues["type"])
	// Gotify resource uses "server" attribute, data source uses "url"
	assert.Equal(t, resource.AttributeValues["server"], dataSource.AttributeValues["url"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])
	assert.Equal(t, resource.AttributeValues["disable"], dataSource.AttributeValues["disable"])
}

// TestNotificationEndpointsDataSourceIntegration tests listing all notification endpoints via data source
func TestNotificationEndpointsDataSourceIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/datasources/notification_endpoints.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Create multiple endpoints of different types
	gotifyName := GenerateTestName("gotify-plural")
	smtpName := GenerateTestName("smtp-plural")

	config := fmt.Sprintf(`
resource "pbs_gotify_notification" "test1" {
  name    = "%s"
  server  = "https://gotify.example.com"
  token   = "Aabcd1234567890" # gitleaks:allow
  comment = "Gotify endpoint 1"
  disable = false
}

resource "pbs_smtp_notification" "test2" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = ["admin@example.com"]
  from_address = "pbs@example.com"
  comment      = "SMTP endpoint 1"
  disable      = false
}

data "pbs_notification_endpoints" "all" {
  depends_on = [
    pbs_gotify_notification.test1,
    pbs_smtp_notification.test2
  ]
}
`, gotifyName, smtpName)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify the resources were created
	res1 := tc.GetResourceFromState(t, "pbs_gotify_notification.test1")
	res2 := tc.GetResourceFromState(t, "pbs_smtp_notification.test2")
	require.Equal(t, gotifyName, res1.AttributeValues["name"])
	require.Equal(t, smtpName, res2.AttributeValues["name"])

	// Verify data source contains our endpoints
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_notification_endpoints.all")
	endpoints, ok := dataSource.AttributeValues["endpoints"].([]interface{})
	require.True(t, ok, "endpoints attribute should be a list")
	require.GreaterOrEqual(t, len(endpoints), 2, "Should have at least 2 endpoints")

	// Count endpoints by type
	typeCounts := make(map[string]int)
	for _, ep := range endpoints {
		endpointMap, ok := ep.(map[string]interface{})
		require.True(t, ok)
		if epType, ok := endpointMap["type"].(string); ok {
			typeCounts[epType]++
		}
	}

	// Verify we have at least one of each type we created
	assert.GreaterOrEqual(t, typeCounts["gotify"], 1, "Should have at least 1 gotify endpoint")
	assert.GreaterOrEqual(t, typeCounts["smtp"], 1, "Should have at least 1 smtp endpoint")

	// Verify that our specific endpoints are in the list
	res1Data := res1.AttributeValues
	res2Data := res2.AttributeValues

	var foundRes1, foundRes2 bool
	for _, ep := range endpoints {
		endpointMap, ok := ep.(map[string]interface{})
		require.True(t, ok)
		name, ok := endpointMap["name"].(string)
		if !ok {
			continue
		}

		if name == res1Data["name"] {
			foundRes1 = true
			assert.Equal(t, "gotify", endpointMap["type"])
		} else if name == res2Data["name"] {
			foundRes2 = true
			assert.Equal(t, "smtp", endpointMap["type"])
		}
	}

	require.True(t, foundRes1, "Did not find resource 1 in data source")
	require.True(t, foundRes2, "Did not find resource 2 in data source")
}

// TestNotificationMatcherDataSourceIntegration tests reading a single notification matcher via data source
func TestNotificationMatcherDataSourceIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/datasources/notification_matcher.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	smtpTarget := GenerateTestName("smtp-matcher-ds")
	matcherName := GenerateTestName("matcher-ds")

	config := fmt.Sprintf(`
resource "pbs_smtp_notification" "target" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = ["admin@example.com"]
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["error", "warning"]
  mode           = "all"
  comment        = "Integration test matcher for data source"
}

data "pbs_notification_matcher" "test" {
  name = pbs_notification_matcher.test.name
}
`, smtpTarget, matcherName)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_notification_matcher.test")
	resource := tc.GetResourceFromState(t, "pbs_notification_matcher.test")

	assert.Equal(t, resource.AttributeValues["name"], dataSource.AttributeValues["name"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])
	assert.Equal(t, resource.AttributeValues["mode"], dataSource.AttributeValues["mode"])

	// Verify targets list
	dsTargets, ok := dataSource.AttributeValues["targets"].([]interface{})
	require.True(t, ok, "targets should be a list")
	resTargets, ok := resource.AttributeValues["targets"].([]interface{})
	require.True(t, ok, "targets should be a list in resource")
	assert.ElementsMatch(t, resTargets, dsTargets)

	// Verify match_severity list
	dsSeverity, ok := dataSource.AttributeValues["match_severity"].([]interface{})
	require.True(t, ok, "match_severity should be a list")
	resSeverity, ok := resource.AttributeValues["match_severity"].([]interface{})
	require.True(t, ok, "match_severity should be a list in resource")
	assert.ElementsMatch(t, resSeverity, dsSeverity)
}

// TestNotificationMatchersDataSourceIntegration tests listing all notification matchers via data source
func TestNotificationMatchersDataSourceIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/datasources/notification_matchers.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	smtpTarget := GenerateTestName("smtp-matchers")
	matcher1 := GenerateTestName("matcher1")
	matcher2 := GenerateTestName("matcher2")

	config := fmt.Sprintf(`
resource "pbs_smtp_notification" "target" {
  name         = "%s"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = ["admin@example.com"]
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test1" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["error"]
  mode           = "all"
  comment        = "Matcher 1"
}

resource "pbs_notification_matcher" "test2" {
  name           = "%s"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["warning"]
  mode           = "any"
  comment        = "Matcher 2"
}

data "pbs_notification_matchers" "all" {
  depends_on = [
    pbs_notification_matcher.test1,
    pbs_notification_matcher.test2
  ]
}
`, smtpTarget, matcher1, matcher2)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify the resources were created
	res1 := tc.GetResourceFromState(t, "pbs_notification_matcher.test1")
	res2 := tc.GetResourceFromState(t, "pbs_notification_matcher.test2")
	require.Equal(t, matcher1, res1.AttributeValues["name"])
	require.Equal(t, matcher2, res2.AttributeValues["name"])

	// Verify data source contains our matchers
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_notification_matchers.all")
	matchers, ok := dataSource.AttributeValues["matchers"].([]interface{})
	require.True(t, ok, "matchers attribute should be a list")
	require.GreaterOrEqual(t, len(matchers), 2, "Should have at least 2 matchers")

	// Verify that our specific matchers are in the list
	res1Data := res1.AttributeValues
	res2Data := res2.AttributeValues

	var foundRes1, foundRes2 bool
	for _, m := range matchers {
		matcherMap, ok := m.(map[string]interface{})
		require.True(t, ok)
		name, ok := matcherMap["name"].(string)
		if !ok {
			continue
		}

		if name == res1Data["name"] {
			foundRes1 = true
			assert.Equal(t, "Matcher 1", matcherMap["comment"])
			assert.Equal(t, "all", matcherMap["mode"])
		} else if name == res2Data["name"] {
			foundRes2 = true
			assert.Equal(t, "Matcher 2", matcherMap["comment"])
			assert.Equal(t, "any", matcherMap["mode"])
		}
	}

	require.True(t, foundRes1, "Did not find matcher 1 in data source")
	require.True(t, foundRes2, "Did not find matcher 2 in data source")
}
