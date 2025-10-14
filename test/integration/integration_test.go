package integration

import (
	"testing"
)

// TestIntegration is the main integration test entry point that runs all provider functionality tests
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("Datastore", func(t *testing.T) {
		// Test directory datastore functionality
		t.Run("Directory", func(t *testing.T) {
			TestDatastoreDirectoryIntegration(t)
		})

		// Test S3 datastore functionality
		t.Run("S3", func(t *testing.T) {
			TestS3DatastoreMultiProvider(t)
		})
	})

	t.Run("S3Endpoints", func(t *testing.T) {
		// Test S3 endpoint creation and management
		TestS3EndpointMultiProvider(t)
	})

	t.Run("Jobs", func(t *testing.T) {
		t.Run("PruneJob", func(t *testing.T) {
			TestPruneJobIntegration(t)
		})

		t.Run("PruneJobWithFilters", func(t *testing.T) {
			TestPruneJobWithFilters(t)
		})

		t.Run("SyncJob", func(t *testing.T) {
			TestSyncJobIntegration(t)
		})

		t.Run("SyncJobWithGroupFilter", func(t *testing.T) {
			TestSyncJobWithGroupFilter(t)
		})

		t.Run("VerifyJob", func(t *testing.T) {
			TestVerifyJobIntegration(t)
		})

		t.Run("GCJob", func(t *testing.T) {
			TestGCJobIntegration(t)
		})
	})

	t.Run("Notifications", func(t *testing.T) {
		t.Run("SMTPTarget", func(t *testing.T) {
			TestSMTPNotificationIntegration(t)
		})

		t.Run("GotifyTarget", func(t *testing.T) {
			TestGotifyNotificationIntegration(t)
		})

		t.Run("SendmailTarget", func(t *testing.T) {
			TestSendmailNotificationIntegration(t)
		})

		t.Run("WebhookTarget", func(t *testing.T) {
			TestWebhookNotificationIntegration(t)
		})

		t.Run("NotificationEndpoint", func(t *testing.T) {
			TestNotificationEndpointIntegration(t)
		})

		t.Run("NotificationMatcher", func(t *testing.T) {
			TestNotificationMatcherIntegration(t)
		})

		t.Run("NotificationMatcherModes", func(t *testing.T) {
			TestNotificationMatcherModes(t)
		})

		t.Run("NotificationMatcherWithCalendar", func(t *testing.T) {
			TestNotificationMatcherWithCalendar(t)
		})

		t.Run("NotificationMatcherInvertMatch", func(t *testing.T) {
			TestNotificationMatcherInvertMatch(t)
		})
	})

	t.Run("Metrics", func(t *testing.T) {
		t.Run("InfluxDBHTTP", func(t *testing.T) {
			TestMetricsServerInfluxDBHTTPIntegration(t)
		})

		t.Run("InfluxDBUDP", func(t *testing.T) {
			TestMetricsServerInfluxDBUDPIntegration(t)
		})

		t.Run("MetricsServerMTU", func(t *testing.T) {
			TestMetricsServerMTU(t)
		})

		t.Run("MetricsServerVerifyCertificate", func(t *testing.T) {
			TestMetricsServerVerifyCertificate(t)
		})

		t.Run("MetricsServerDisabled", func(t *testing.T) {
			TestMetricsServerDisabled(t)
		})

		t.Run("MetricsServerMaxBodySize", func(t *testing.T) {
			TestMetricsServerMaxBodySize(t)
		})

		// TestMetricsServerTimeout - skipped, PBS 4.0 removed timeout parameter
	})
}

// TestQuickSmoke runs basic smoke tests that should always pass
func TestQuickSmoke(t *testing.T) {
	// Quick connectivity and basic function tests
	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Test PBS server connectivity
	datastoreName := GenerateTestName("smoke-test")

	// Simple directory datastore creation as smoke test
	testConfig := `
resource "pbs_datastore" "smoke" {
  name = "` + datastoreName + `"
  type = "dir"
  path = "/datastore/` + datastoreName + `"
}
`

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	// Verify resource exists in state
	resource := tc.GetResourceFromState(t, "pbs_datastore.smoke")
	if resource.AttributeValues["name"] != datastoreName {
		t.Errorf("Expected datastore name %s, got %v", datastoreName, resource.AttributeValues["name"])
	}

	t.Logf("Smoke test passed: PBS provider basic functionality works")
}
