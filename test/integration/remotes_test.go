package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/remotes"
)

// TestRemotesIntegration tests the complete lifecycle of PBS remote configurations
func TestRemotesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	remoteName := GenerateTestName("remote")

	// Create initial config with required fields
	config := fmt.Sprintf(`
resource "pbs_remote" "test_remote" {
  name     = "%s"
  host     = "pbs.example.com"
  auth_id  = "sync@pbs!test-token"
  password = "test-password-12345"
  comment  = "Test remote server"
}
`, remoteName)

	// Write terraform configuration
	tc.WriteMainTF(t, config)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_remote.test_remote")
	assert.Equal(t, remoteName, resource.AttributeValues["name"])
	assert.Equal(t, "pbs.example.com", resource.AttributeValues["host"])
	assert.Equal(t, "sync@pbs!test-token", resource.AttributeValues["auth_id"])
	assert.Equal(t, "Test remote server", resource.AttributeValues["comment"])
	// Password should be in state (write-only, but stored)
	assert.Equal(t, "test-password-12345", resource.AttributeValues["password"])

	// Verify digest is present (may be empty string initially, that's ok)
	digest := resource.AttributeValues["digest"]
	assert.NotNil(t, digest)
	t.Logf("Digest value from state: %v", digest)

	// Verify remote exists via direct API call
	remotesClient := remotes.NewClient(tc.APIClient)
	remote, err := remotesClient.GetRemote(context.Background(), remoteName)
	require.NoError(t, err, "Remote should exist via API")
	assert.Equal(t, remoteName, remote.Name)
	assert.Equal(t, "pbs.example.com", remote.Host)
	assert.Equal(t, "sync@pbs!test-token", remote.AuthID)
	assert.Equal(t, "Test remote server", remote.Comment)
	// Password should NOT be returned by API
	assert.Empty(t, remote.Password, "Password should not be returned by GET")
	t.Logf("SUCCESS: Remote %s found via API with digest %s", remoteName, remote.Digest)

	// Test update - change host, port, and comment
	updatedConfig := fmt.Sprintf(`
resource "pbs_remote" "test_remote" {
  name        = "%s"
  host        = "backup.example.com"
  port        = 8008
  auth_id     = "sync@pbs!test-token"
  password    = "test-password-12345"
  fingerprint = "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
  comment     = "Updated test remote server"
}
`, remoteName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify update via API
	remote, err = remotesClient.GetRemote(context.Background(), remoteName)
	require.NoError(t, err)
	assert.Equal(t, "backup.example.com", remote.Host)
	assert.NotNil(t, remote.Port)
	assert.Equal(t, 8008, *remote.Port)
	assert.Equal(t, "Updated test remote server", remote.Comment)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99", remote.Fingerprint)

	// Test removing optional fields (port, fingerprint, comment)
	clearedConfig := fmt.Sprintf(`
resource "pbs_remote" "test_remote" {
  name     = "%s"
  host     = "backup.example.com"
  auth_id  = "sync@pbs!test-token"
  password = "test-password-12345"
}
`, remoteName)

	tc.WriteMainTF(t, clearedConfig)
	tc.ApplyTerraform(t)

	// Verify optional fields were cleared
	remote, err = remotesClient.GetRemote(context.Background(), remoteName)
	require.NoError(t, err)
	assert.Nil(t, remote.Port, "Port should be cleared")
	assert.Empty(t, remote.Fingerprint, "Fingerprint should be cleared")
	assert.Empty(t, remote.Comment, "Comment should be cleared")
}

// TestRemotePasswordUpdate tests that password updates work correctly
func TestRemotePasswordUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	remoteName := GenerateTestName("pass-remote")

	config := fmt.Sprintf(`
resource "pbs_remote" "password_test" {
  name     = "%s"
  host     = "pbs.example.com"
  auth_id  = "admin@pam"
  password = "initial-password"
}
`, remoteName)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify remote created
	resource := tc.GetResourceFromState(t, "pbs_remote.password_test")
	assert.Equal(t, "initial-password", resource.AttributeValues["password"])

	// Update password
	updatedConfig := fmt.Sprintf(`
resource "pbs_remote" "password_test" {
  name     = "%s"
  host     = "pbs.example.com"
  auth_id  = "admin@pam"
  password = "new-password-54321"
}
`, remoteName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify password in state was updated
	resource = tc.GetResourceFromState(t, "pbs_remote.password_test")
	assert.Equal(t, "new-password-54321", resource.AttributeValues["password"])

	// Password is write-only, so we can't verify via API
	// But we can verify the remote still exists and wasn't replaced
	remotesClient := remotes.NewClient(tc.APIClient)
	remote, err := remotesClient.GetRemote(context.Background(), remoteName)
	require.NoError(t, err)
	assert.Equal(t, remoteName, remote.Name)
}

// TestRemoteDataSources tests the remote data sources for stores, namespaces, and groups
func TestRemoteDataSources(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	remoteName := GenerateTestName("ds-remote")

	// Note: These data sources require a real remote PBS server to scan
	// For testing purposes, we create the remote configuration but don't expect
	// the scan to succeed (would require actual remote server)
	config := fmt.Sprintf(`
resource "pbs_remote" "data_source_test" {
  name     = "%s"
  host     = "pbs.example.com"
  auth_id  = "sync@pbs!test-token"
  password = "test-password"
}

# These data sources would work with a real remote server
# In testing, they may fail if the remote is not accessible
# data "pbs_remote_stores" "test_stores" {
#   remote_name = pbs_remote.data_source_test.name
# }
#
# data "pbs_remote_namespaces" "test_namespaces" {
#   remote_name = pbs_remote.data_source_test.name
#   store       = "datastore1"
# }
#
# data "pbs_remote_groups" "test_groups" {
#   remote_name = pbs_remote.data_source_test.name
#   store       = "datastore1"
#   namespace   = "production"
# }

output "remote_name" {
  value = pbs_remote.data_source_test.name
}
`, remoteName)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify remote was created
	resource := tc.GetResourceFromState(t, "pbs_remote.data_source_test")
	assert.Equal(t, remoteName, resource.AttributeValues["name"])

	t.Log("INFO: Data source tests skipped - require real remote PBS server")
}

// TestRemoteValidation tests validation scenarios
func TestRemoteValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Test invalid name (too short)
	invalidNameConfig := `
resource "pbs_remote" "invalid_name" {
  name     = "ab"
  host     = "pbs.example.com"
  auth_id  = "admin@pam"
  password = "test"
}
`

	tc.WriteMainTF(t, invalidNameConfig)
	err := tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for name too short")

	// Test invalid auth_id format
	invalidAuthIDConfig := `
resource "pbs_remote" "invalid_auth" {
  name     = "test-remote"
  host     = "pbs.example.com"
  auth_id  = "invalid-format"
  password = "test"
}
`

	tc.WriteMainTF(t, invalidAuthIDConfig)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for invalid auth_id format")

	// Test invalid fingerprint format
	invalidFingerprintConfig := `
resource "pbs_remote" "invalid_fingerprint" {
  name        = "test-remote"
  host        = "pbs.example.com"
  auth_id     = "admin@pam"
  password    = "test"
  fingerprint = "invalid-fingerprint"
}
`

	tc.WriteMainTF(t, invalidFingerprintConfig)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for invalid fingerprint format")

	// Test invalid port
	invalidPortConfig := `
resource "pbs_remote" "invalid_port" {
  name     = "test-remote"
  host     = "pbs.example.com"
  port     = 99999
  auth_id  = "admin@pam"
  password = "test"
}
`

	tc.WriteMainTF(t, invalidPortConfig)
	err = tc.ApplyTerraformWithError(t)
	assert.Error(t, err, "Should fail validation for port out of range")
}

// TestRemoteImport tests importing existing remotes
func TestRemoteImport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	remoteName := GenerateTestName("import-remote")

	// First, create a remote manually via API
	remotesClient := remotes.NewClient(tc.APIClient)
	testRemote := &remotes.Remote{
		Name:    remoteName,
		Host:    "import.example.com",
		AuthID:  "sync@pbs!import-token",
		Password: "import-password",
		Comment: "Test remote for import",
	}

	err := remotesClient.CreateRemote(context.Background(), testRemote)
	require.NoError(t, err, "Failed to create remote via API for import test")

	// Create Terraform config for import
	importConfig := fmt.Sprintf(`
resource "pbs_remote" "imported" {
  name     = "%s"
  host     = "import.example.com"
  auth_id  = "sync@pbs!import-token"
  password = "import-password"
  comment  = "Test remote for import"
}
`, remoteName)

	tc.WriteMainTF(t, importConfig)

	// Import the remote
	tc.ImportResource(t, "pbs_remote.imported", remoteName)

	// Verify imported state matches
	resource := tc.GetResourceFromState(t, "pbs_remote.imported")
	assert.Equal(t, remoteName, resource.AttributeValues["name"])
	assert.Equal(t, "import.example.com", resource.AttributeValues["host"])
	assert.Equal(t, "sync@pbs!import-token", resource.AttributeValues["auth_id"])
	// Password will be null after import (per ImportState), and populated from config on first apply/refresh
	assert.Equal(t, "Test remote for import", resource.AttributeValues["comment"])

	t.Log("SUCCESS: Remote imported successfully")
}
