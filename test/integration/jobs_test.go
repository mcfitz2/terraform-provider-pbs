package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

// TestPruneJobIntegration tests the complete lifecycle of a prune job
func TestPruneJobIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	jobID := GenerateTestName("prune-job")
	datastoreName := "datastore1" // Assuming a datastore exists

	// Test configuration for prune job
	testConfig := fmt.Sprintf(`
resource "pbs_prune_job" "test_prune" {
  id             = "%s"
  store          = "%s"
  schedule       = "daily"
  keep_last      = 7
  keep_daily     = 14
  keep_weekly    = 8
  keep_monthly   = 12
  keep_yearly    = 3
  max_depth      = 3
  comment        = "Test prune job"
}
`, jobID, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_prune_job.test_prune")
	assert.Equal(t, jobID, resource.AttributeValues["id"])
	assert.Equal(t, datastoreName, resource.AttributeValues["store"])
	assert.Equal(t, "daily", resource.AttributeValues["schedule"])
	assert.Equal(t, json.Number("7"), resource.AttributeValues["keep_last"])
	assert.Equal(t, json.Number("14"), resource.AttributeValues["keep_daily"])
	assert.Equal(t, "Test prune job", resource.AttributeValues["comment"])

	// Verify job exists via direct API call
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetPruneJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
	assert.Equal(t, "daily", job.Schedule)
	assert.NotNil(t, job.KeepLast)
	assert.Equal(t, 7, *job.KeepLast)
	assert.NotNil(t, job.KeepDaily)
	assert.Equal(t, 14, *job.KeepDaily)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_prune_job" "test_prune" {
  id             = "%s"
  store          = "%s"
  schedule       = "weekly"
  keep_last      = 10
  keep_daily     = 21
  keep_weekly    = 12
  keep_monthly   = 18
  keep_yearly    = 5
  max_depth      = 5
  comment        = "Updated test prune job"
}
`, jobID, datastoreName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify update via API
	job, err = jobsClient.GetPruneJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, "weekly", job.Schedule)
	assert.Equal(t, "Updated test prune job", job.Comment)
	assert.NotNil(t, job.KeepLast)
	assert.Equal(t, 10, *job.KeepLast)
}

// TestSyncJobIntegration tests the complete lifecycle of a sync job
func TestSyncJobIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	jobID := GenerateTestName("sync-job")
	datastoreName := "datastore1"
	remoteName := "remote1" // Assuming a remote exists

	// Test configuration for sync job
	testConfig := fmt.Sprintf(`
resource "pbs_sync_job" "test_sync" {
  id             = "%s"
  store          = "%s"
  remote         = "%s"
  remote_store   = "backup"
  schedule       = "hourly"
  remove_vanished = true
  rate_limit_in  = "10M"
  comment        = "Test sync job"
}
`, jobID, datastoreName, remoteName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_sync_job.test_sync")
	assert.Equal(t, jobID, resource.AttributeValues["id"])
	assert.Equal(t, datastoreName, resource.AttributeValues["store"])
	assert.Equal(t, remoteName, resource.AttributeValues["remote"])
	assert.Equal(t, "backup", resource.AttributeValues["remote_store"])
	assert.Equal(t, "hourly", resource.AttributeValues["schedule"])
	assert.Equal(t, true, resource.AttributeValues["remove_vanished"])

	// Verify job exists via direct API call
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetSyncJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
	assert.Equal(t, remoteName, job.Remote)
	assert.Equal(t, "backup", job.RemoteStore)
	assert.Equal(t, "hourly", job.Schedule)
	assert.NotNil(t, job.RemoveVanished)
	assert.True(t, *job.RemoveVanished)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_sync_job" "test_sync" {
  id             = "%s"
  store          = "%s"
  remote         = "%s"
  remote_store   = "backup"
  schedule       = "daily"
  remove_vanished = false
  rate_limit_in  = "20M"
  comment        = "Updated test sync job"
}
`, jobID, datastoreName, remoteName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify update via API
	job, err = jobsClient.GetSyncJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, "daily", job.Schedule)
	assert.Equal(t, "Updated test sync job", job.Comment)
	assert.NotNil(t, job.RemoveVanished)
	assert.False(t, *job.RemoveVanished)
}

// TestVerifyJobIntegration tests the complete lifecycle of a verify job
func TestVerifyJobIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	jobID := GenerateTestName("verify-job")
	datastoreName := "datastore1"

	// Test configuration for verify job
	testConfig := fmt.Sprintf(`
resource "pbs_verify_job" "test_verify" {
  id              = "%s"
  store           = "%s"
  schedule        = "weekly"
  ignore_verified = true
  outdated_after  = 30
  max_depth       = 3
  comment         = "Test verify job"
}
`, jobID, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_verify_job.test_verify")
	assert.Equal(t, jobID, resource.AttributeValues["id"])
	assert.Equal(t, datastoreName, resource.AttributeValues["store"])
	assert.Equal(t, "weekly", resource.AttributeValues["schedule"])
	assert.Equal(t, true, resource.AttributeValues["ignore_verified"])
	assert.Equal(t, json.Number("30"), resource.AttributeValues["outdated_after"])

	// Verify job exists via direct API call
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetVerifyJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
	assert.Equal(t, "weekly", job.Schedule)
	assert.NotNil(t, job.IgnoreVerified)
	assert.True(t, *job.IgnoreVerified)
	assert.NotNil(t, job.OutdatedAfter)
	assert.Equal(t, 30, *job.OutdatedAfter)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_verify_job" "test_verify" {
  id              = "%s"
  store           = "%s"
  schedule        = "monthly"
  ignore_verified = false
  outdated_after  = 60
  max_depth       = 5
  comment         = "Updated test verify job"
}
`, jobID, datastoreName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify update via API
	job, err = jobsClient.GetVerifyJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, "monthly", job.Schedule)
	assert.Equal(t, "Updated test verify job", job.Comment)
	assert.NotNil(t, job.OutdatedAfter)
	assert.Equal(t, 60, *job.OutdatedAfter)
}

// TestGCJobIntegration tests the complete lifecycle of a GC job
func TestGCJobIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	jobID := GenerateTestName("gc-job")
	datastoreName := "datastore1"

	// Test configuration for GC job
	testConfig := fmt.Sprintf(`
resource "pbs_gc_job" "test_gc" {
  id       = "%s"
  store    = "%s"
  schedule = "weekly"
  comment  = "Test GC job"
}
`, jobID, datastoreName)

	// Write terraform configuration
	tc.WriteMainTF(t, testConfig)

	// Apply terraform
	tc.ApplyTerraform(t)

	// Verify resource was created via Terraform state
	resource := tc.GetResourceFromState(t, "pbs_gc_job.test_gc")
	assert.Equal(t, jobID, resource.AttributeValues["id"])
	assert.Equal(t, datastoreName, resource.AttributeValues["store"])
	assert.Equal(t, "weekly", resource.AttributeValues["schedule"])
	assert.Equal(t, "Test GC job", resource.AttributeValues["comment"])

	// Verify job exists via direct API call
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetGCJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
	assert.Equal(t, "weekly", job.Schedule)
	assert.Equal(t, "Test GC job", job.Comment)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_gc_job" "test_gc" {
  id       = "%s"
  store    = "%s"
  schedule = "daily"
  comment  = "Updated test GC job"
}
`, jobID, datastoreName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	// Verify update via API
	job, err = jobsClient.GetGCJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, "daily", job.Schedule)
	assert.Equal(t, "Updated test GC job", job.Comment)
}

// TestPruneJobWithFilters tests prune job with backup filters
func TestPruneJobWithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	jobID := GenerateTestName("prune-filter")
	datastoreName := "datastore1"

	testConfig := fmt.Sprintf(`
resource "pbs_prune_job" "test_filter" {
  id           = "%s"
  store        = "%s"
  schedule     = "daily"
  keep_last    = 5
  namespace    = "vm"
  max_depth    = 2
  comment      = "Prune job with filters"
}
`, jobID, datastoreName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_prune_job.test_filter")
	assert.Equal(t, "vm", resource.AttributeValues["namespace"])

	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetPruneJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, "vm", job.NamespaceRE)
}

// TestSyncJobWithGroupFilter tests sync job with group filters
func TestSyncJobWithGroupFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	jobID := GenerateTestName("sync-filter")
	datastoreName := "datastore1"
	remoteName := "remote1"

	testConfig := fmt.Sprintf(`
resource "pbs_sync_job" "test_filter" {
  id           = "%s"
  store        = "%s"
  remote       = "%s"
  remote_store = "backup"
  schedule     = "daily"
  group_filter = ["type:vm", "type:ct"]
  namespace    = "production"
  comment      = "Sync job with filters"
}
`, jobID, datastoreName, remoteName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_sync_job.test_filter")
	assert.NotNil(t, resource.AttributeValues["group_filter"])
	assert.Equal(t, "production", resource.AttributeValues["namespace"])

	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetSyncJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Len(t, job.GroupFilter, 2)
	assert.Contains(t, job.GroupFilter, "type:vm")
	assert.Contains(t, job.GroupFilter, "type:ct")
	assert.Equal(t, "production", job.NamespaceRE)
}
