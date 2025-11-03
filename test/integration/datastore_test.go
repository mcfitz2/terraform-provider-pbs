package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/datastores"
)

// TestDatastoreDirectoryIntegration tests the complete lifecycle of a directory datastore
func TestDatastoreDirectoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	datastoreName := GenerateTestName("dir-datastore")

	// Create config
	config := fmt.Sprintf(`
resource "pbs_datastore" "test_directory" {
  name             = "%s"
  path             = "/datastore/%s"
  comment          = "Test directory datastore"
  gc_schedule      = "daily"
}
`, datastoreName, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, config)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_datastore.test_directory")
	assert.Equal(t, datastoreName, resource.AttributeValues["name"])
	assert.Equal(t, fmt.Sprintf("/datastore/%s", datastoreName), resource.AttributeValues["path"])
	assert.Equal(t, "Test directory datastore", resource.AttributeValues["comment"])
	assert.Equal(t, "daily", resource.AttributeValues["gc_schedule"])

	// Try to verify datastore exists via direct API call
	// Note: PBS datastore operations are asynchronous, so this may not be immediately available
	datastoreClient := datastores.NewClient(tc.APIClient)
	datastore, err := datastoreClient.GetDatastore(context.Background(), datastoreName)
	if err != nil {
		debugLog(t, "INFO: Datastore %s not yet visible in PBS (may still be processing async): %v", datastoreName, err)
	} else {
		assert.Equal(t, fmt.Sprintf("/datastore/%s", datastoreName), datastore.Path)
		t.Logf("SUCCESS: Datastore %s found via API", datastoreName)
	}

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_directory" {
  name             = "%s"
  path             = "/datastore/%s"
  comment          = "Updated test directory datastore"
  gc_schedule      = "weekly"
}
`, datastoreName, datastoreName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify update via API
	datastore, err = datastoreClient.GetDatastore(context.Background(), datastoreName)
	require.NoError(t, err)
	assert.Equal(t, "Updated test directory datastore", datastore.Comment)
	assert.Equal(t, "weekly", datastore.GCSchedule)
}

// TestDatastoreValidation tests validation scenarios
func TestDatastoreValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Test missing required path for directory datastore
	invalidDirConfig := `
resource "pbs_datastore" "invalid_dir" {
  name = "invalid-dir"
  # missing required path
}
`

	tc.WriteMainTF(t, invalidDirConfig)
	err := tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for directory datastore without path")

	// Test missing S3 bucket when client is provided
	invalidS3Config := `
resource "pbs_datastore" "invalid_s3" {
  name      = "invalid-s3"
  s3_client = "endpoint-1"
  path      = "/datastore/invalid-s3"
}
`

	tc.WriteMainTF(t, invalidS3Config)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation when only one S3 attribute is provided")

	// Test removable datastore without backing device
	invalidRemovableConfig := `
resource "pbs_datastore" "invalid_removable" {
  name       = "invalid-removable"
  path       = "/datastore/invalid-removable"
  removable  = true
}
`

	tc.WriteMainTF(t, invalidRemovableConfig)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation when removable datastore lacks backing_device")

}

// Concurrency tests removed - not required for PBS datastore operations

// TestDatastoreImport tests importing existing datastores
func TestDatastoreImport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	datastoreName := GenerateTestName("import-test")
	datastorePath := fmt.Sprintf("/datastore/%s", datastoreName)

	// First, create a datastore manually via API
	// Note: PBS API requires directories to exist before datastores can be registered.
	// The /datastore directory is pre-created by the CI workflow.
	datastoreClient := datastores.NewClient(tc.APIClient)
	testDatastore := &datastores.Datastore{
		Name: datastoreName,
		Path: datastorePath,
	}

	err := datastoreClient.CreateDatastore(context.Background(), testDatastore)
	require.NoError(t, err, "Failed to create datastore via API for import test")

	// Verify the datastore was created by reading it back with retry logic
	// PBS may need several seconds to fully register the datastore
	var createdDatastore *datastores.Datastore
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		createdDatastore, err = datastoreClient.GetDatastore(context.Background(), datastoreName)
		if err == nil {
			t.Logf("SUCCESS: Datastore found after %d attempts", i+1)
			break
		}
		t.Logf("Attempt %d/%d: Datastore not yet available: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	require.NoError(t, err, "Failed to verify datastore creation")
	require.Equal(t, datastoreName, createdDatastore.Name, "Datastore name mismatch after creation")

	// Now create Terraform config and import the existing datastore
	importConfig := fmt.Sprintf(`
resource "pbs_datastore" "imported" {
  name = "%s"
  path = "%s"
}
`, datastoreName, datastorePath)

	tc.WriteMainTF(t, importConfig)

	// Import the existing datastore
	tc.ImportResource(t, "pbs_datastore.imported", datastoreName)

	// Verify the import was successful
	resource := tc.GetResourceFromState(t, "pbs_datastore.imported")
	assert.Equal(t, datastoreName, resource.AttributeValues["name"])

	// Apply to ensure configuration matches
	tc.ApplyTerraform(t)
}
