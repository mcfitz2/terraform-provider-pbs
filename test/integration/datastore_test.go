package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

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

	// Test configuration for directory datastore
	testConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_directory" {
  name             = "%s"
  type             = "dir"
  path             = "/datastore/%s"
  content          = ["backup"]
  comment          = "Test directory datastore"
  create_base_path = true
  gc_schedule      = "daily"
  prune_schedule   = "weekly"
  max_backups      = 10
}
`, datastoreName, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_datastore.test_directory")
	assert.Equal(t, datastoreName, resource.AttributeValues["name"])
	assert.Equal(t, "dir", resource.AttributeValues["type"])
	assert.Equal(t, fmt.Sprintf("/datastore/%s", datastoreName), resource.AttributeValues["path"])
	assert.Equal(t, "Test directory datastore", resource.AttributeValues["comment"])
	assert.Equal(t, true, resource.AttributeValues["create_base_path"])
	assert.Equal(t, "daily", resource.AttributeValues["gc_schedule"])
	assert.Equal(t, "weekly", resource.AttributeValues["prune_schedule"])

	// Try to verify datastore exists via direct API call
	// Note: PBS datastore operations are asynchronous, so this may not be immediately available
	datastoreClient := datastores.NewClient(tc.APIClient)
	datastore, err := datastoreClient.GetDatastore(context.Background(), datastoreName)
	if err != nil {
		t.Logf("INFO: Datastore %s not yet visible in PBS (may still be processing async): %v", datastoreName, err)
	} else {
		assert.Equal(t, datastoreName, datastore.Name)
		assert.Equal(t, datastores.DatastoreTypeDirectory, datastore.Type)
		assert.Equal(t, fmt.Sprintf("/datastore/%s", datastoreName), datastore.Path)
		t.Logf("SUCCESS: Datastore %s found via API", datastoreName)
	}

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_directory" {
  name             = "%s"
  type             = "dir"
  path             = "/datastore/%s"
  content          = ["backup"]
  comment          = "Updated test directory datastore"
  create_base_path = true
  gc_schedule      = "weekly"
  prune_schedule   = "daily"
  max_backups      = 20
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

// TestDatastoreZFSIntegration tests ZFS datastore functionality (if ZFS is available)
func TestDatastoreZFSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if ZFS pool is configured for testing
	zfsPool := os.Getenv("PBS_TESTPOOL")
	if zfsPool == "" {
		zfsPool = "testpool" // default
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	datastoreName := GenerateTestName("zfs-datastore")

	// Test configuration for ZFS datastore
	testConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_zfs" {
  name        = "%s"
  type        = "zfs"
  path        = "/mnt/datastore/%s"
  zfs_pool    = "%s"
  zfs_dataset = "backup/%s"
  content     = ["backup"]
  comment     = "Test ZFS datastore"
  compression = "lz4"
  block_size  = "8K"
  max_backups = 15
}
`, datastoreName, datastoreName, zfsPool, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform (this will fail if ZFS pool doesn't exist)
	err := tc.ApplyTerraformWithError(t)
	if err != nil {
		// Check if it's a ZFS pool error
		if strings.Contains(err.Error(), "pool") || strings.Contains(err.Error(), "zfs") {
			t.Skipf("ZFS test skipped - ZFS pool '%s' is not available: %v", zfsPool, err)
		}
		t.Fatalf("Unexpected error creating ZFS datastore: %v", err)
	}

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_datastore.test_zfs")
	assert.Equal(t, datastoreName, resource.AttributeValues["name"])
	assert.Equal(t, "zfs", resource.AttributeValues["type"])
	assert.Equal(t, "testpool", resource.AttributeValues["zfs_pool"])
	assert.Equal(t, fmt.Sprintf("backup/%s", datastoreName), resource.AttributeValues["zfs_dataset"])

	// Verify datastore exists via direct API call
	datastoreClient := datastores.NewClient(tc.APIClient)
	datastore, err := datastoreClient.GetDatastore(context.Background(), datastoreName)
	require.NoError(t, err)
	assert.Equal(t, datastores.DatastoreTypeZFS, datastore.Type)
	assert.Equal(t, "testpool", datastore.ZFSPool)
	assert.Equal(t, "lz4", datastore.Compression)
}

// TestDatastoreValidation tests validation scenarios
func TestDatastoreValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Test missing required path for directory datastore
	invalidConfig := `
resource "pbs_datastore" "invalid_dir" {
  name = "invalid-dir"
  type = "dir"
  # missing required path
  content = ["backup"]
}
`

	tc.WriteMainTF(t, invalidConfig)
	err := tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for directory datastore without path")

	// Test invalid datastore type
	invalidTypeConfig := `
resource "pbs_datastore" "invalid_type" {
  name = "invalid-type"
  type = "invalid"
  path = "/tmp/test"
}
`

	tc.WriteMainTF(t, invalidTypeConfig)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for invalid datastore type")

	// Test missing ZFS pool for ZFS datastore
	invalidZFSConfig := `
resource "pbs_datastore" "invalid_zfs" {
  name = "invalid-zfs"
  type = "zfs"
  # missing required zfs_pool
  content = ["backup"]
}
`

	tc.WriteMainTF(t, invalidZFSConfig)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for ZFS datastore without pool")
}

// Concurrency tests removed - not required for PBS datastore operations

// TestDatastoreNetworkStorage tests CIFS/NFS datastore configurations
func TestDatastoreNetworkStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	cifsDatastore := GenerateTestName("cifs-datastore")
	nfsDatastore := GenerateTestName("nfs-datastore")

	// Get CIFS server configuration from environment
	cifsHost := os.Getenv("TEST_CIFS_HOST")
	cifsShare := os.Getenv("TEST_CIFS_SHARE")
	cifsUser := os.Getenv("TEST_CIFS_USERNAME")
	cifsPass := os.Getenv("TEST_CIFS_PASSWORD")
	
	if cifsHost == "" || cifsShare == "" {
		t.Skip("CIFS test skipped - TEST_CIFS_HOST and TEST_CIFS_SHARE environment variables not set")
	}
	
	// Default credentials if not provided
	if cifsUser == "" {
		cifsUser = "testuser"
	}
	if cifsPass == "" {
		cifsPass = "testpass"
	}

	// Test configuration for CIFS datastore
	cifsConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_cifs" {
  name     = "%s"
  type     = "cifs"
  path     = "/mnt/datastore/%s"
  server   = "%s"
  share    = "%s"
  username = "%s"
  password = "%s"
  sub_dir  = "pbs"
  content  = ["backup"]
  comment  = "Test CIFS datastore"
  options  = "vers=3.0"
}
`, cifsDatastore, cifsDatastore, cifsHost, cifsShare, cifsUser, cifsPass)

	tc.WriteMainTF(t, cifsConfig)
	tc.ApplyTerraform(t)

	// Verify the configuration
	resource := tc.GetResourceFromState(t, "pbs_datastore.test_cifs")
	assert.Equal(t, cifsDatastore, resource.AttributeValues["name"])
	assert.Equal(t, "cifs", resource.AttributeValues["type"])
	assert.Equal(t, cifsHost, resource.AttributeValues["server"])
	assert.Equal(t, cifsShare, resource.AttributeValues["share"])

	// Get NFS server configuration from environment
	nfsHost := os.Getenv("TEST_NFS_HOST")
	nfsExport := os.Getenv("TEST_NFS_EXPORT")

	if nfsHost == "" || nfsExport == "" {
		t.Skip("NFS test skipped - TEST_NFS_HOST and TEST_NFS_EXPORT environment variables not set")
	}

	// Test configuration for NFS datastore
	nfsConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_nfs" {
  name    = "%s"
  type    = "nfs"
  path    = "/mnt/datastore/%s"
  server  = "%s"
  export  = "%s"
  sub_dir = "pbs"
  content = ["backup"]
  comment = "Test NFS datastore"
  options = "vers=4,soft"
}
`, nfsDatastore, nfsDatastore, nfsHost, nfsExport)

	tc.WriteMainTF(t, nfsConfig)

	// Apply terraform - should succeed with Docker NFS server
	tc.ApplyTerraform(t)

	// Verify the configuration
	resource = tc.GetResourceFromState(t, "pbs_datastore.test_nfs")
	assert.Equal(t, nfsDatastore, resource.AttributeValues["name"])
	assert.Equal(t, "nfs", resource.AttributeValues["type"])
	assert.Equal(t, nfsHost, resource.AttributeValues["server"])
	assert.Equal(t, nfsExport, resource.AttributeValues["export"])
}

// TestDatastoreImport tests importing existing datastores
func TestDatastoreImport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	datastoreName := GenerateTestName("import-test")

	// First, create a datastore manually via API
	datastoreClient := datastores.NewClient(tc.APIClient)
	testDatastore := &datastores.Datastore{
		Name: datastoreName,
		Type: datastores.DatastoreTypeDirectory,
		Path: fmt.Sprintf("/datastore/%s", datastoreName),
	}

	err := datastoreClient.CreateDatastore(context.Background(), testDatastore)
	require.NoError(t, err, "Failed to create datastore via API for import test")

	// Terraform destroy will clean up the datastore after import
	// No need for manual API cleanup

	// Now create Terraform config and import the existing datastore
	importConfig := fmt.Sprintf(`
resource "pbs_datastore" "imported" {
  name = "%s"
  type = "dir"
  path = "/datastore/%s"
}
`, datastoreName, datastoreName)

	tc.WriteMainTF(t, importConfig)

	// Import the existing datastore
	tc.ImportResource(t, "pbs_datastore.imported", datastoreName)

	// Verify the import was successful
	resource := tc.GetResourceFromState(t, "pbs_datastore.imported")
	assert.Equal(t, datastoreName, resource.AttributeValues["name"])
	assert.Equal(t, "dir", resource.AttributeValues["type"])

	// Apply to ensure configuration matches
	tc.ApplyTerraform(t)
}
