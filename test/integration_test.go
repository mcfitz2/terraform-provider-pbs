package test

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

	t.Logf("✅ Smoke test passed: PBS provider basic functionality works")
}
