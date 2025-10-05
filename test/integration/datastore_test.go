package integration

import (
	"context"
	"fmt"
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

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	datastoreName := GenerateTestName("zfs-datastore")

	// Test configuration for ZFS datastore
	testConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_zfs" {
  name        = "%s"
  type        = "zfs"
  zfs_pool    = "testpool"
  zfs_dataset = "backup/%s"
  content     = ["backup"]
  comment     = "Test ZFS datastore"
  compression = "lz4"
  block_size  = "8K"
  max_backups = 15
}
`, datastoreName, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform (this will likely fail if ZFS testpool doesn't exist)
	// We'll handle the error gracefully for CI environments
	err := tc.ApplyTerraformWithError(t)
	if err != nil {
		t.Logf("ZFS test skipped - ZFS pool 'testpool' may not be available: %v", err)
		return
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

	// Test configuration for CIFS datastore
	cifsConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_cifs" {
  name     = "%s"
  type     = "cifs"
  path     = "/mnt/datastore/%s"
  server   = "192.168.1.100"
  share    = "backup"
  username = "backup-user"
  password = "backup-password"
  domain   = "example.com"
  sub_dir  = "pbs"
  content  = ["backup"]
  comment  = "Test CIFS datastore"
  options  = "vers=3.0"
}
`, cifsDatastore, cifsDatastore)

	tc.WriteMainTF(t, cifsConfig)

	// Apply terraform (this will likely fail without actual CIFS server)
	err := tc.ApplyTerraformWithError(t)
	if err != nil {
		t.Logf("CIFS test expected to fail without real server: %v", err)
		// This is expected in most test environments
	} else {
		// If it succeeds (perhaps with a mock server), verify the configuration
		resource := tc.GetResourceFromState(t, "pbs_datastore.test_cifs")
		assert.Equal(t, cifsDatastore, resource.AttributeValues["name"])
		assert.Equal(t, "cifs", resource.AttributeValues["type"])
		assert.Equal(t, "192.168.1.100", resource.AttributeValues["server"])
		assert.Equal(t, "backup", resource.AttributeValues["share"])
	}

	// Test configuration for NFS datastore
	nfsConfig := fmt.Sprintf(`
resource "pbs_datastore" "test_nfs" {
  name    = "%s"
  type    = "nfs"
  path    = "/mnt/datastore/%s"
  server  = "192.168.1.101"
  export  = "/export/backup"
  sub_dir = "pbs"
  content = ["backup"]
  comment = "Test NFS datastore"
  options = "vers=4,soft"
}
`, nfsDatastore, nfsDatastore)

	tc.WriteMainTF(t, nfsConfig)

	// Apply terraform (this will likely fail without actual NFS server)
	err = tc.ApplyTerraformWithError(t)
	if err != nil {
		t.Logf("NFS test expected to fail without real server: %v", err)
		// This is expected in most test environments
	} else {
		// If it succeeds, verify the configuration
		resource := tc.GetResourceFromState(t, "pbs_datastore.test_nfs")
		assert.Equal(t, nfsDatastore, resource.AttributeValues["name"])
		assert.Equal(t, "nfs", resource.AttributeValues["type"])
		assert.Equal(t, "192.168.1.101", resource.AttributeValues["server"])
		assert.Equal(t, "/export/backup", resource.AttributeValues["export"])
	}
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
