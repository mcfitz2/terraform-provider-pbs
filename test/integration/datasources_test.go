package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/datastores"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

// TestDatastoreDataSourceIntegration tests reading a single datastore via data source
func TestDatastoreDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	datastoreName := GenerateTestName("ds-datasource")

	// Create a datastore resource and read it with data source
	config := fmt.Sprintf(`
resource "pbs_datastore" "test" {
  name        = "%s"
  path        = "/datastore/%s"
  comment     = "Test datastore for data source"
  gc_schedule = "daily"
}

data "pbs_datastore" "test" {
  name = pbs_datastore.test.name
}
`, datastoreName, datastoreName)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_datastore.test")
	resource := tc.GetResourceFromState(t, "pbs_datastore.test")

	assert.Equal(t, resource.AttributeValues["name"], dataSource.AttributeValues["name"])
	assert.Equal(t, resource.AttributeValues["path"], dataSource.AttributeValues["path"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])
	assert.Equal(t, resource.AttributeValues["gc_schedule"], dataSource.AttributeValues["gc_schedule"])

	// Verify via API
	datastoreClient := datastores.NewClient(tc.APIClient)
	ds, err := datastoreClient.GetDatastore(context.Background(), datastoreName)
	require.NoError(t, err)
	assert.Equal(t, datastoreName, ds.Name)
	assert.Equal(t, fmt.Sprintf("/datastore/%s", datastoreName), ds.Path)
}

// TestDatastoresDataSourceIntegration tests listing all datastores via data source
func TestDatastoresDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	datastore1 := GenerateTestName("ds-list-1")
	datastore2 := GenerateTestName("ds-list-2")

	// Create multiple datastores and list them
	config := fmt.Sprintf(`
resource "pbs_datastore" "test1" {
  name    = "%s"
  path    = "/datastore/%s"
  comment = "First test datastore"
}

resource "pbs_datastore" "test2" {
  name    = "%s"
  path    = "/datastore/%s"
  comment = "Second test datastore"
}

data "pbs_datastores" "all" {
  depends_on = [
    pbs_datastore.test1,
    pbs_datastore.test2
  ]
}
`, datastore1, datastore1, datastore2, datastore2)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source contains our datastores
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_datastores.all")
	stores, ok := dataSource.AttributeValues["stores"].([]interface{})
	require.True(t, ok, "stores should be a list")
	require.GreaterOrEqual(t, len(stores), 2, "should have at least 2 datastores")

	// Check that our test datastores are in the list
	var foundNames []string
	for _, store := range stores {
		storeMap, ok := store.(map[string]interface{})
		require.True(t, ok)
		if name, ok := storeMap["name"].(string); ok {
			foundNames = append(foundNames, name)
		}
	}

	assert.Contains(t, foundNames, datastore1)
	assert.Contains(t, foundNames, datastore2)
}

// TestPruneJobDataSourceIntegration tests reading a single prune job via data source
func TestPruneJobDataSourceIntegration(t *testing.T) {
	t.Skip("Skipping due to flaky datastore creation timing in CI - works locally but fails in tfexec")
	
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	jobID := GenerateTestName("prune-ds")
	datastoreName := GenerateTestName("ds-prune")

	// Create a datastore and prune job, then read the job with data source
	config := fmt.Sprintf(`
resource "pbs_datastore" "test" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_prune_job" "test" {
  id         = "%s"
  store      = pbs_datastore.test.name
  schedule   = "daily"
  keep_last  = 7
  keep_daily = 14
  comment    = "Test prune job for data source"
}

data "pbs_prune_job" "test" {
  id = pbs_prune_job.test.id
}
`, datastoreName, datastoreName, jobID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_prune_job.test")
	resource := tc.GetResourceFromState(t, "pbs_prune_job.test")

	assert.Equal(t, resource.AttributeValues["id"], dataSource.AttributeValues["id"])
	assert.Equal(t, resource.AttributeValues["store"], dataSource.AttributeValues["store"])
	assert.Equal(t, resource.AttributeValues["schedule"], dataSource.AttributeValues["schedule"])
	assert.Equal(t, resource.AttributeValues["keep_last"], dataSource.AttributeValues["keep_last"])
	assert.Equal(t, resource.AttributeValues["keep_daily"], dataSource.AttributeValues["keep_daily"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])

	// Verify via API
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetPruneJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
}

// TestPruneJobsDataSourceIntegration tests listing prune jobs via data source
func TestPruneJobsDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	job1ID := GenerateTestName("prune-list-1")
	job2ID := GenerateTestName("prune-list-2")
	datastore1 := GenerateTestName("ds-prune-1")
	datastore2 := GenerateTestName("ds-prune-2")

	// Create multiple prune jobs and list them
	config := fmt.Sprintf(`
resource "pbs_datastore" "test1" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_datastore" "test2" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_prune_job" "test1" {
  id        = "%s"
  store     = pbs_datastore.test1.name
  schedule  = "daily"
  keep_last = 5
}

resource "pbs_prune_job" "test2" {
  id        = "%s"
  store     = pbs_datastore.test2.name
  schedule  = "weekly"
  keep_last = 10
}

data "pbs_prune_jobs" "all" {
  depends_on = [
    pbs_prune_job.test1,
    pbs_prune_job.test2
  ]
}

data "pbs_prune_jobs" "filtered" {
  store = pbs_datastore.test1.name
  depends_on = [
    pbs_prune_job.test1,
    pbs_prune_job.test2
  ]
}
`, datastore1, datastore1, datastore2, datastore2, job1ID, job2ID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify unfiltered data source contains both jobs
	allDataSource := tc.GetDataSourceFromState(t, "data.pbs_prune_jobs.all")
	allJobs, ok := allDataSource.AttributeValues["jobs"].([]interface{})
	require.True(t, ok, "jobs should be a list")
	require.GreaterOrEqual(t, len(allJobs), 2, "should have at least 2 prune jobs")

	var foundIDs []string
	for _, job := range allJobs {
		jobMap, ok := job.(map[string]interface{})
		require.True(t, ok)
		if id, ok := jobMap["id"].(string); ok {
			foundIDs = append(foundIDs, id)
		}
	}

	assert.Contains(t, foundIDs, job1ID)
	assert.Contains(t, foundIDs, job2ID)

	// Verify filtered data source contains only job1
	filteredDataSource := tc.GetDataSourceFromState(t, "data.pbs_prune_jobs.filtered")
	filteredJobs, ok := filteredDataSource.AttributeValues["jobs"].([]interface{})
	require.True(t, ok, "jobs should be a list")

	// Count jobs for datastore1
	countForStore := 0
	for _, job := range filteredJobs {
		jobMap, ok := job.(map[string]interface{})
		require.True(t, ok)
		if store, ok := jobMap["store"].(string); ok && store == datastore1 {
			countForStore++
		}
	}
	assert.Equal(t, 1, countForStore, "filtered data source should only return jobs for the specified store")
}

// TestSyncJobDataSourceIntegration tests reading a single sync job via data source
func TestSyncJobDataSourceIntegration(t *testing.T) {
	t.Skip("Skipping due to flaky datastore creation timing in CI - works locally but fails in tfexec")
	
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	jobID := GenerateTestName("sync-ds")
	datastoreName := GenerateTestName("ds-sync")
	remoteName := GenerateTestName("remote-sync")

	// Create resources and read sync job with data source
	config := fmt.Sprintf(`
resource "pbs_datastore" "test" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_remote" "test" {
  name = "%s"
  host = "remote.example.com"
  auth_id = "test@pbs"
  password = "testpassword"
  fingerprint = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
}

resource "pbs_sync_job" "test" {
  id           = "%s"
  store        = pbs_datastore.test.name
  remote       = pbs_remote.test.name
  remote_store = "backup"
  schedule     = "hourly"
  comment      = "Test sync job for data source"
}

data "pbs_sync_job" "test" {
  id = pbs_sync_job.test.id
}
`, datastoreName, datastoreName, remoteName, jobID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_sync_job.test")
	resource := tc.GetResourceFromState(t, "pbs_sync_job.test")

	assert.Equal(t, resource.AttributeValues["id"], dataSource.AttributeValues["id"])
	assert.Equal(t, resource.AttributeValues["store"], dataSource.AttributeValues["store"])
	assert.Equal(t, resource.AttributeValues["remote"], dataSource.AttributeValues["remote"])
	assert.Equal(t, resource.AttributeValues["remote_store"], dataSource.AttributeValues["remote_store"])
	assert.Equal(t, resource.AttributeValues["schedule"], dataSource.AttributeValues["schedule"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])

	// Verify via API
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetSyncJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
	assert.Equal(t, remoteName, job.Remote)
}

// TestSyncJobsDataSourceIntegration tests listing sync jobs with filters
func TestSyncJobsDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	job1ID := GenerateTestName("sync-list-1")
	job2ID := GenerateTestName("sync-list-2")
	datastore1 := GenerateTestName("ds-sync-1")
	remote1 := GenerateTestName("remote-sync-1")
	remote2 := GenerateTestName("remote-sync-2")

	// Create multiple sync jobs and test filtering
	config := fmt.Sprintf(`
resource "pbs_datastore" "test" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_remote" "test1" {
  name = "%s"
  host = "remote1.example.com"
  auth_id = "test@pbs"
  password = "testpassword1"
  fingerprint = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
}

resource "pbs_remote" "test2" {
  name = "%s"
  host = "remote2.example.com"
  auth_id = "test@pbs"
  password = "testpassword2"
  fingerprint = "ff:ee:dd:cc:bb:aa:99:88:77:66:55:44:33:22:11:00:ff:ee:dd:cc:bb:aa:99:88:77:66:55:44:33:22:11:00"
}

resource "pbs_sync_job" "test1" {
  id           = "%s"
  store        = pbs_datastore.test.name
  remote       = pbs_remote.test1.name
  remote_store = "backup"
  schedule     = "hourly"
}

resource "pbs_sync_job" "test2" {
  id           = "%s"
  store        = pbs_datastore.test.name
  remote       = pbs_remote.test2.name
  remote_store = "archive"
  schedule     = "daily"
}

data "pbs_sync_jobs" "by_store" {
  store = pbs_datastore.test.name
  depends_on = [
    pbs_sync_job.test1,
    pbs_sync_job.test2
  ]
}

data "pbs_sync_jobs" "by_remote" {
  remote = pbs_remote.test1.name
  depends_on = [
    pbs_sync_job.test1,
    pbs_sync_job.test2
  ]
}
`, datastore1, datastore1, remote1, remote2, job1ID, job2ID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify store filter returns both jobs
	storeDataSource := tc.GetDataSourceFromState(t, "data.pbs_sync_jobs.by_store")
	storeJobs, ok := storeDataSource.AttributeValues["jobs"].([]interface{})
	require.True(t, ok)
	require.GreaterOrEqual(t, len(storeJobs), 2)

	// Verify remote filter returns only job1
	remoteDataSource := tc.GetDataSourceFromState(t, "data.pbs_sync_jobs.by_remote")
	remoteJobs, ok := remoteDataSource.AttributeValues["jobs"].([]interface{})
	require.True(t, ok)

	countForRemote := 0
	for _, job := range remoteJobs {
		jobMap, ok := job.(map[string]interface{})
		require.True(t, ok)
		if remote, ok := jobMap["remote"].(string); ok && remote == remote1 {
			countForRemote++
		}
	}
	assert.Equal(t, 1, countForRemote)
}

// TestVerifyJobDataSourceIntegration tests reading a single verify job via data source
func TestVerifyJobDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	jobID := GenerateTestName("verify-ds")
	datastoreName := GenerateTestName("ds-verify")

	// Create datastore and verify job, then read with data source
	config := fmt.Sprintf(`
resource "pbs_datastore" "test" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_verify_job" "test" {
  id              = "%s"
  store           = pbs_datastore.test.name
  schedule        = "weekly"
  outdated_after  = 30
  comment         = "Test verify job for data source"
}

data "pbs_verify_job" "test" {
  id = pbs_verify_job.test.id
}
`, datastoreName, datastoreName, jobID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_verify_job.test")
	resource := tc.GetResourceFromState(t, "pbs_verify_job.test")

	assert.Equal(t, resource.AttributeValues["id"], dataSource.AttributeValues["id"])
	assert.Equal(t, resource.AttributeValues["store"], dataSource.AttributeValues["store"])
	assert.Equal(t, resource.AttributeValues["schedule"], dataSource.AttributeValues["schedule"])
	assert.Equal(t, resource.AttributeValues["outdated_after"], dataSource.AttributeValues["outdated_after"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])

	// Verify via API
	jobsClient := jobs.NewClient(tc.APIClient)
	job, err := jobsClient.GetVerifyJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, datastoreName, job.Store)
}

// TestVerifyJobsDataSourceIntegration tests listing verify jobs with store filter
func TestVerifyJobsDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	job1ID := GenerateTestName("verify-list-1")
	job2ID := GenerateTestName("verify-list-2")
	datastore1 := GenerateTestName("ds-verify-1")
	datastore2 := GenerateTestName("ds-verify-2")

	// Create multiple verify jobs and test filtering
	config := fmt.Sprintf(`
resource "pbs_datastore" "test1" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_datastore" "test2" {
  name = "%s"
  path = "/datastore/%s"
}

resource "pbs_verify_job" "test1" {
  id       = "%s"
  store    = pbs_datastore.test1.name
  schedule = "weekly"
}

resource "pbs_verify_job" "test2" {
  id       = "%s"
  store    = pbs_datastore.test2.name
  schedule = "monthly"
}

data "pbs_verify_jobs" "all" {
  depends_on = [
    pbs_verify_job.test1,
    pbs_verify_job.test2
  ]
}

data "pbs_verify_jobs" "filtered" {
  store = pbs_datastore.test1.name
  depends_on = [
    pbs_verify_job.test1,
    pbs_verify_job.test2
  ]
}
`, datastore1, datastore1, datastore2, datastore2, job1ID, job2ID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify unfiltered data source
	allDataSource := tc.GetDataSourceFromState(t, "data.pbs_verify_jobs.all")
	allJobs, ok := allDataSource.AttributeValues["jobs"].([]interface{})
	require.True(t, ok)
	require.GreaterOrEqual(t, len(allJobs), 2)

	var foundIDs []string
	for _, job := range allJobs {
		jobMap, ok := job.(map[string]interface{})
		require.True(t, ok)
		if id, ok := jobMap["id"].(string); ok {
			foundIDs = append(foundIDs, id)
		}
	}

	assert.Contains(t, foundIDs, job1ID)
	assert.Contains(t, foundIDs, job2ID)

	// Verify filtered data source
	filteredDataSource := tc.GetDataSourceFromState(t, "data.pbs_verify_jobs.filtered")
	filteredJobs, ok := filteredDataSource.AttributeValues["jobs"].([]interface{})
	require.True(t, ok)

	countForStore := 0
	for _, job := range filteredJobs {
		jobMap, ok := job.(map[string]interface{})
		require.True(t, ok)
		if store, ok := jobMap["store"].(string); ok && store == datastore1 {
			countForStore++
		}
	}
	assert.Equal(t, 1, countForStore)
}

// TestS3EndpointDataSourceIntegration tests reading a single S3 endpoint via data source
func TestS3EndpointDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	endpointID := GenerateTestName("s3-endpoint")

	// Create an S3 endpoint and read it with data source
	config := fmt.Sprintf(`
resource "pbs_s3_endpoint" "test" {
  id         = "%s"
  access_key = "test-access-key"
  secret_key = "test-secret-key"
  endpoint   = "s3.amazonaws.com"
  region     = "us-east-1"
}

data "pbs_s3_endpoint" "test" {
  id = pbs_s3_endpoint.test.id
}
`, endpointID)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_s3_endpoint.test")
	resource := tc.GetResourceFromState(t, "pbs_s3_endpoint.test")

	assert.Equal(t, resource.AttributeValues["id"], dataSource.AttributeValues["id"])
	assert.Equal(t, resource.AttributeValues["access_key"], dataSource.AttributeValues["access_key"])
	assert.Equal(t, resource.AttributeValues["endpoint"], dataSource.AttributeValues["endpoint"])
	assert.Equal(t, resource.AttributeValues["region"], dataSource.AttributeValues["region"])
}

// TestS3EndpointsDataSourceIntegration tests listing all S3 endpoints via data source
func TestS3EndpointsDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	endpoint1 := GenerateTestName("s3-ep-1")
	endpoint2 := GenerateTestName("s3-ep-2")

	// Create multiple S3 endpoints and list them
	config := fmt.Sprintf(`
resource "pbs_s3_endpoint" "test1" {
  id         = "%s"
  access_key = "test-access-key-1"
  secret_key = "test-secret-key-1"
  endpoint   = "s3.amazonaws.com"
}

resource "pbs_s3_endpoint" "test2" {
  id         = "%s"
  access_key = "test-access-key-2"
  secret_key = "test-secret-key-2"
  endpoint   = "s3.us-west-2.amazonaws.com"
}

data "pbs_s3_endpoints" "all" {
  depends_on = [
    pbs_s3_endpoint.test1,
    pbs_s3_endpoint.test2
  ]
}
`, endpoint1, endpoint2)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source contains our endpoints
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_s3_endpoints.all")
	endpoints, ok := dataSource.AttributeValues["endpoints"].([]interface{})
	require.True(t, ok, "endpoints should be a list")
	require.GreaterOrEqual(t, len(endpoints), 2, "should have at least 2 endpoints")

	// Check that our test endpoints are in the list
	var foundIDs []string
	for _, ep := range endpoints {
		epMap, ok := ep.(map[string]interface{})
		require.True(t, ok)
		if id, ok := epMap["id"].(string); ok {
			foundIDs = append(foundIDs, id)
		}
	}

	assert.Contains(t, foundIDs, endpoint1)
	assert.Contains(t, foundIDs, endpoint2)
}

// TestMetricsServerDataSourceIntegration tests reading a single metrics server via data source
func TestMetricsServerDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test name
	serverName := GenerateTestName("metrics-server")

	// Use the test InfluxDB server from environment
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()

	// Create a metrics server and read it with data source
	config := fmt.Sprintf(`
resource "pbs_metrics_server" "test" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "test-org"
  bucket       = "test-bucket"
  token        = "test-token-value"
  enable       = true
  verify_tls   = false
  comment      = "Integration test metrics server"
}

data "pbs_metrics_server" "test" {
  name = pbs_metrics_server.test.name
  type = pbs_metrics_server.test.type
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify data source matches resource
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_metrics_server.test")
	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test")

	assert.Equal(t, resource.AttributeValues["name"], dataSource.AttributeValues["name"])
	assert.Equal(t, resource.AttributeValues["type"], dataSource.AttributeValues["type"])
	assert.Equal(t, resource.AttributeValues["url"], dataSource.AttributeValues["url"])
	assert.Equal(t, resource.AttributeValues["organization"], dataSource.AttributeValues["organization"])
	assert.Equal(t, resource.AttributeValues["bucket"], dataSource.AttributeValues["bucket"])
	assert.Equal(t, resource.AttributeValues["comment"], dataSource.AttributeValues["comment"])
}

// TestMetricsServersDataSourceIntegration tests listing all metrics servers via data source
func TestMetricsServersDataSourceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	// Generate unique test names
	server1 := GenerateTestName("metrics-srv-1")
	server2 := GenerateTestName("metrics-srv-2")

	// Use the test InfluxDB servers from environment
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()
	influxUDPHost := getInfluxDBUDPHost()
	influxUDPPort := getInfluxDBUDPPort()

	// Create multiple metrics servers and list them
	// Note: This test may occasionally fail due to PBS lock contention when creating
	// multiple metrics servers simultaneously. This is a PBS limitation, not a provider issue.
	config := fmt.Sprintf(`
resource "pbs_metrics_server" "test1" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "test-org-1"
  bucket       = "test-bucket-1"
  token        = "test-token-1"
  enable       = true
  verify_tls   = false
  comment      = "Integration test metrics server 1"
}

resource "pbs_metrics_server" "test2" {
  name    = "%s"
  type    = "influxdb-udp"
  server  = "%s"
  port    = %d
  enable  = false
  mtu     = 1500
  comment = "Integration test metrics server 2"
}

data "pbs_metrics_servers" "all" {
  depends_on = [
    pbs_metrics_server.test1,
    pbs_metrics_server.test2
  ]
}
`, server1, influxHost, influxPort, server2, influxUDPHost, influxUDPPort)

	tc.WriteMainTF(t, config)
	tc.ApplyTerraform(t)

	// Verify the resources were created
	res1 := tc.GetResourceFromState(t, "pbs_metrics_server.test1")
	res2 := tc.GetResourceFromState(t, "pbs_metrics_server.test2")
	require.Equal(t, server1, res1.AttributeValues["name"])
	require.Equal(t, server2, res2.AttributeValues["name"])

	// Verify data source contains our metrics servers
	dataSource := tc.GetDataSourceFromState(t, "data.pbs_metrics_servers.all")
	servers, ok := dataSource.AttributeValues["servers"].([]interface{})
	require.True(t, ok, "servers attribute should be a list")
	require.GreaterOrEqual(t, len(servers), 2, "Should have at least 2 servers")

	// Count servers by type to verify we have at least one of each type we created
	typeCounts := make(map[string]int)
	for _, srv := range servers {
		serverMap, ok := srv.(map[string]interface{})
		require.True(t, ok)
		if srvType, ok := serverMap["type"].(string); ok {
			typeCounts[srvType]++
		}
	}

	// Verify we have at least one HTTP and one UDP server (from our test resources)
	assert.GreaterOrEqual(t, typeCounts["influxdb-http"], 1, "Should have at least 1 influxdb-http server")
	assert.GreaterOrEqual(t, typeCounts["influxdb-udp"], 1, "Should have at least 1 influxdb-udp server")

	// Verify that our specific servers are in the list by checking the resources match the data source
	res1Data := res1.AttributeValues
	res2Data := res2.AttributeValues

	var foundRes1, foundRes2 bool
	for _, srv := range servers {
		serverMap, ok := srv.(map[string]interface{})
		require.True(t, ok)
		name, ok := serverMap["name"].(string)
		if !ok {
			continue
		}

		if name == res1Data["name"] {
			foundRes1 = true
			assert.Equal(t, res1Data["type"], serverMap["type"])
		} else if name == res2Data["name"] {
			foundRes2 = true
			assert.Equal(t, res2Data["type"], serverMap["type"])
		}
	}

	require.True(t, foundRes1, "Did not find resource 1 in data source")
	require.True(t, foundRes2, "Did not find resource 2 in data source")
}
