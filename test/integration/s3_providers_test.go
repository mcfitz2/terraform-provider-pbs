package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/datastores"
	"github.com/micah/terraform-provider-pbs/pbs/endpoints"
)

// TestS3EndpointMultiProvider tests S3 endpoint functionality with multiple real S3 providers
func TestS3EndpointMultiProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-provider integration test in short mode")
	}

	providers := GetS3ProviderConfigs(t)
	if len(providers) == 0 {
		t.Skip("No S3 providers configured for testing")
	}

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			// No parallel execution - PBS server operations must be sequential

			testS3ProviderEndpoint(t, provider)
		})
	}
}

// testS3ProviderEndpoint tests a single S3 provider end-to-end
func testS3ProviderEndpoint(t *testing.T, provider *S3ProviderConfig) {
	// Setup PBS test context
	tc := SetupTest(t)
	defer func() {
		// Always try to clean up the bucket
		provider.DeleteTestBucket(t)
		tc.DestroyTerraform(t)
	}()

	t.Logf("Testing S3 endpoint with %s provider", provider.Name)

	// Step 1: Setup S3 client and create test bucket
	provider.SetupS3Client(t)
	provider.CreateTestBucket(t)

	// Step 2: Test S3 connectivity directly
	provider.TestS3Connectivity(t)

	// Step 3: Get PBS endpoint configuration
	pbsConfig := provider.GetPBSEndpointConfig()

	// Step 4: Create Terraform configuration for PBS S3 endpoint
	terraformConfig := fmt.Sprintf(`
resource "pbs_s3_endpoint" "test_%s" {
  id         = "%s"
  endpoint   = "%s"
  region     = "%s"
  access_key = "%s"
  secret_key = "%s"
`,
		provider.Name,
		pbsConfig["id"],
		pbsConfig["endpoint"],
		pbsConfig["region"],
		pbsConfig["access_key"],
		pbsConfig["secret_key"])

	// Add optional parameters based on provider
	if pathStyle, exists := pbsConfig["path_style"]; exists {
		terraformConfig += fmt.Sprintf(`  path_style = %s
`, pathStyle)
	}

	if providerQuirks, exists := pbsConfig["provider_quirks"]; exists {
		terraformConfig += fmt.Sprintf(`  provider_quirks = %s
`, providerQuirks)
	}

	terraformConfig += "}\n"

	// Step 5: Apply Terraform configuration
	tc.WriteMainTF(t, terraformConfig)
	tc.ApplyTerraform(t)

	// Step 6: Verify PBS S3 endpoint was created via Terraform state
	resource := tc.GetResourceFromState(t, fmt.Sprintf("pbs_s3_endpoint.test_%s", provider.Name))
	assert.Equal(t, pbsConfig["id"], resource.AttributeValues["id"])
	assert.Equal(t, pbsConfig["endpoint"], resource.AttributeValues["endpoint"])
	assert.Equal(t, pbsConfig["region"], resource.AttributeValues["region"])
	assert.Equal(t, pbsConfig["access_key"], resource.AttributeValues["access_key"])

	// Step 7: Verify PBS S3 endpoint via direct API call
	endpointsClient := endpoints.NewClient(tc.APIClient)
	pbsEndpoint, err := endpointsClient.GetS3Endpoint(context.Background(), pbsConfig["id"])
	require.NoError(t, err, "Failed to get S3 endpoint via PBS API")
	assert.Equal(t, pbsConfig["id"], pbsEndpoint.ID)
	assert.Equal(t, pbsConfig["endpoint"], pbsEndpoint.Endpoint)
	assert.Equal(t, pbsConfig["region"], pbsEndpoint.Region)
	assert.Equal(t, pbsConfig["access_key"], pbsEndpoint.AccessKey)

	// Step 8: Test PBS S3 endpoint functionality (optional - would require more PBS setup)
	t.Logf("Successfully created and verified PBS S3 endpoint for %s provider", provider.Name)

	// Terraform destroy will be handled by defer
}

// TestS3EndpointProviderSpecificFeatures tests provider-specific S3 configurations
func TestS3EndpointProviderSpecificFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping provider-specific feature tests in short mode")
	}

	providers := GetS3ProviderConfigs(t)
	if len(providers) == 0 {
		t.Skip("No S3 providers configured for testing")
	}

	for _, provider := range providers {
		provider := provider // capture loop variable
		t.Run(fmt.Sprintf("%s_Features", provider.Name), func(t *testing.T) {
			testProviderSpecificFeatures(t, provider)
		})
	}
}

// testProviderSpecificFeatures tests provider-specific configurations and features
func testProviderSpecificFeatures(t *testing.T, provider *S3ProviderConfig) {
	tc := SetupTest(t)
	defer func() {
		provider.DeleteTestBucket(t)
		tc.DestroyTerraform(t)
	}()

	provider.SetupS3Client(t)
	provider.CreateTestBucket(t)

	pbsConfig := provider.GetPBSEndpointConfig()

	var terraformConfig string

	switch provider.Name {
	case "AWS":
		// Test AWS-specific features
		terraformConfig = fmt.Sprintf(`
resource "pbs_s3_endpoint" "test_aws_features" {
  id         = "%s"
  endpoint   = "%s"
  region     = "%s" 
  access_key = "%s"
  secret_key = "%s"
  # AWS supports virtual-hosted style (default)
}
`, pbsConfig["id"], pbsConfig["endpoint"], pbsConfig["region"], pbsConfig["access_key"], pbsConfig["secret_key"])

	case "Backblaze":
		// Test Backblaze B2 specific features
		terraformConfig = fmt.Sprintf(`
resource "pbs_s3_endpoint" "test_backblaze_features" {
  id         = "%s"
  endpoint   = "%s"
  region     = "%s"
  access_key = "%s"  
  secret_key = "%s"
  path_style = true  # B2 requires path-style addressing
}
`, pbsConfig["id"], pbsConfig["endpoint"], pbsConfig["region"], pbsConfig["access_key"], pbsConfig["secret_key"])

	case "Scaleway":
		// Test Scaleway specific features
		terraformConfig = fmt.Sprintf(`
resource "pbs_s3_endpoint" "test_scaleway_features" {
  id         = "%s"
  endpoint   = "%s" 
  region     = "%s"
  access_key = "%s"
  secret_key = "%s"
  # Scaleway supports virtual-hosted style (default)
}
`, pbsConfig["id"], pbsConfig["endpoint"], pbsConfig["region"], pbsConfig["access_key"], pbsConfig["secret_key"])

	default:
		t.Skipf("No specific feature tests for provider: %s", provider.Name)
	}

	tc.WriteMainTF(t, terraformConfig)
	tc.ApplyTerraform(t)

	// Verify the endpoint was created with provider-specific configurations
	resourceName := fmt.Sprintf("pbs_s3_endpoint.test_%s_features", strings.ToLower(provider.Name))
	resource := tc.GetResourceFromState(t, resourceName)
	assert.Equal(t, pbsConfig["id"], resource.AttributeValues["id"])

	// Verify via API
	endpointsClient := endpoints.NewClient(tc.APIClient)
	pbsEndpoint, err := endpointsClient.GetS3Endpoint(context.Background(), pbsConfig["id"])
	require.NoError(t, err)

	// Provider-specific assertions
	switch provider.Name {
	case "Backblaze":
		if pbsEndpoint.PathStyle != nil {
			assert.True(t, *pbsEndpoint.PathStyle, "Backblaze should use path-style addressing")
		}
	}

	t.Logf("Successfully tested %s provider-specific features", provider.Name)
}

// TestS3EndpointConcurrentProviders tests creating endpoints for multiple providers simultaneously
func TestS3EndpointConcurrentProviders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent provider test in short mode")
	}

	providers := GetS3ProviderConfigs(t)
	if len(providers) < 2 {
		t.Skip("Need at least 2 S3 providers for concurrent testing")
	}

	tc := SetupTest(t)
	defer func() {
		for _, provider := range providers {
			provider.DeleteTestBucket(t)
		}
		tc.DestroyTerraform(t)
	}()

	// Setup all providers
	var terraformConfigs []string
	for i, provider := range providers {
		provider.SetupS3Client(t)
		provider.CreateTestBucket(t)

		pbsConfig := provider.GetPBSEndpointConfig()
		config := fmt.Sprintf(`
resource "pbs_s3_endpoint" "concurrent_%d" {
  id         = "%s"
  endpoint   = "%s"
  region     = "%s"
  access_key = "%s"
  secret_key = "%s"
}
`, i, pbsConfig["id"], pbsConfig["endpoint"], pbsConfig["region"], pbsConfig["access_key"], pbsConfig["secret_key"])

		terraformConfigs = append(terraformConfigs, config)
	}

	// Apply all configurations at once
	allConfigs := ""
	for _, config := range terraformConfigs {
		allConfigs += config
	}

	tc.WriteMainTF(t, allConfigs)
	tc.ApplyTerraform(t)

	// Verify all endpoints were created
	endpointsClient := endpoints.NewClient(tc.APIClient)
	for i, provider := range providers {
		resource := tc.GetResourceFromState(t, fmt.Sprintf("pbs_s3_endpoint.concurrent_%d", i))
		pbsConfig := provider.GetPBSEndpointConfig()
		assert.Equal(t, pbsConfig["id"], resource.AttributeValues["id"])

		// Verify via API
		pbsEndpoint, err := endpointsClient.GetS3Endpoint(context.Background(), pbsConfig["id"])
		require.NoError(t, err, "Failed to get endpoint for provider %s", provider.Name)
		assert.Equal(t, pbsConfig["id"], pbsEndpoint.ID)
	}

	t.Logf("Successfully created and verified concurrent endpoints for %d providers", len(providers))
}

// TestS3DatastoreMultiProvider tests S3 datastore functionality with multiple real S3 providers
// Creates bucket, endpoint, datastore, verifies all, then cleans up for each provider
func TestS3DatastoreMultiProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-provider S3 datastore integration test in short mode")
	}

	providers := GetS3ProviderConfigs(t)
	if len(providers) == 0 {
		t.Skip("No S3 providers configured for testing")
	}

	for _, provider := range providers {
		provider := provider // capture loop variable
		t.Run(provider.Name, func(t *testing.T) {
			// No parallel execution - PBS server operations must be sequential

			testS3ProviderDatastore(t, provider)
		})
	}
}

// testS3ProviderDatastore tests a single S3 provider's complete datastore lifecycle
func testS3ProviderDatastore(t *testing.T, provider *S3ProviderConfig) {
	// Setup PBS test context
	tc := SetupTest(t)
	defer func() {
		// Terraform destroy will handle PBS resources (datastore -> endpoint)
		// in the correct order thanks to depends_on relationships
		tc.DestroyTerraform(t)

		// Clean up the S3 bucket (external to PBS)
		provider.DeleteTestBucket(t)
	}()

	t.Logf("Testing S3 datastore with %s provider", provider.Name)

	// Step 1: Setup S3 client and create test bucket
	provider.SetupS3Client(t)
	provider.CreateTestBucket(t)

	// Step 2: Test S3 connectivity directly
	provider.TestS3Connectivity(t)

	// Step 3: Get PBS endpoint configuration
	pbsConfig := provider.GetPBSEndpointConfig()

	// Step 4: Generate unique datastore name
	datastoreName := GenerateTestName(fmt.Sprintf("s3-%s-datastore", strings.ToLower(provider.Name)))

	// Step 5: Create Terraform configuration for PBS S3 endpoint and datastore
	terraformConfig := fmt.Sprintf(`
resource "pbs_s3_endpoint" "test_%s" {
  id         = "%s"
  endpoint   = "%s"
  region     = "%s"
  access_key = "%s"
  secret_key = "%s"
`,
		provider.Name,
		pbsConfig["id"],
		pbsConfig["endpoint"],
		pbsConfig["region"],
		pbsConfig["access_key"],
		pbsConfig["secret_key"])

	// Add provider-specific path_style parameter
	if pathStyle, exists := pbsConfig["path_style"]; exists {
		terraformConfig += fmt.Sprintf(`  path_style = %s
`, pathStyle)
	}

	// Add provider_quirks if present (required for B2)
	if providerQuirks, exists := pbsConfig["provider_quirks"]; exists {
		terraformConfig += fmt.Sprintf(`  provider_quirks = %s
`, providerQuirks)
	}

	terraformConfig += "}\n"

	// Add S3 datastore configuration
	// NOTE: We use the .id attribute reference which creates an implicit dependency
	// ensuring the datastore can only be created after the endpoint exists.
	// During destroy, Terraform reverses this dependency, destroying the datastore first.
	endpointID := pbsConfig["id"]
	terraformConfig += fmt.Sprintf(`
resource "pbs_datastore" "test_%s" {
  name       = "%s"
  path       = "/datastore/%s-cache"
  s3_client  = "%s"
  s3_bucket  = "%s"
  comment    = "Test S3 datastore for %s provider"
  
  # Explicit dependency ensures proper destroy order: datastore -> endpoint
  depends_on = [pbs_s3_endpoint.test_%s]
}
`,
		provider.Name,
		datastoreName,
		datastoreName,
		endpointID,
		provider.BucketName,
		provider.Name,
		provider.Name)

	// Step 6: Apply Terraform configuration
	tc.WriteMainTF(t, terraformConfig)
	tc.ApplyTerraform(t)

	// Step 7: Verify S3 endpoint was created via Terraform state (with fallback to API verification)
	endpointResource := tc.GetResourceFromState(t, fmt.Sprintf("pbs_s3_endpoint.test_%s", provider.Name))
	if endpointResource != nil {
		assert.Equal(t, pbsConfig["id"], endpointResource.AttributeValues["id"])
		assert.Equal(t, pbsConfig["endpoint"], endpointResource.AttributeValues["endpoint"])
		assert.Equal(t, pbsConfig["region"], endpointResource.AttributeValues["region"])
		assert.Equal(t, pbsConfig["access_key"], endpointResource.AttributeValues["access_key"])
	} else {
		t.Logf("Warning: Could not verify S3 endpoint via Terraform state, will verify via PBS API only")
	}

	// Step 8: Verify S3 datastore was created via Terraform state (with fallback to API verification)
	datastoreResource := tc.GetResourceFromState(t, fmt.Sprintf("pbs_datastore.test_%s", provider.Name))
	if datastoreResource != nil {
		assert.Equal(t, datastoreName, datastoreResource.AttributeValues["name"])
		assert.Equal(t, fmt.Sprintf("/datastore/%s-cache", datastoreName), datastoreResource.AttributeValues["path"])
		assert.Equal(t, pbsConfig["id"], datastoreResource.AttributeValues["s3_client"])
		assert.Equal(t, provider.BucketName, datastoreResource.AttributeValues["s3_bucket"])
	} else {
		t.Logf("Warning: Could not verify S3 datastore via Terraform state, will verify via PBS API only")
	}

	// Step 9: Verify S3 endpoint via direct PBS API call
	endpointsClient := endpoints.NewClient(tc.APIClient)
	pbsEndpoint, err := endpointsClient.GetS3Endpoint(context.Background(), pbsConfig["id"])
	require.NoError(t, err, "Failed to get S3 endpoint via PBS API")
	assert.Equal(t, pbsConfig["id"], pbsEndpoint.ID)
	assert.Equal(t, pbsConfig["endpoint"], pbsEndpoint.Endpoint)
	assert.Equal(t, pbsConfig["region"], pbsEndpoint.Region)
	assert.Equal(t, pbsConfig["access_key"], pbsEndpoint.AccessKey)

	// Step 10: Verify S3 datastore via direct PBS API call (with retry for eventual consistency)
	datastoresClient := datastores.NewClient(tc.APIClient)
	var pbsDatastore *datastores.Datastore
	maxRetries := 10 // Increased retries for AWS eventual consistency
	for i := 0; i < maxRetries; i++ {
		pbsDatastore, err = datastoresClient.GetDatastore(context.Background(), datastoreName)
		if err == nil {
			t.Logf("Datastore successfully retrieved on attempt %d/%d", i+1, maxRetries)
			break
		}
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 2 * time.Second // Exponential backoff: 2s, 4s, 6s, 8s, 10s, ...
			t.Logf("Datastore not yet available, waiting %v... (attempt %d/%d) - error: %v", waitTime, i+1, maxRetries, err)
			time.Sleep(waitTime)
		}
	}
	require.NoError(t, err, "Failed to get S3 datastore via PBS API after %d retries", maxRetries)
	assert.Equal(t, datastoreName, pbsDatastore.Name)
	assert.Equal(t, fmt.Sprintf("/datastore/%s-cache", datastoreName), pbsDatastore.Path)
	assert.Equal(t, pbsConfig["id"], pbsDatastore.S3Client)
	assert.Equal(t, provider.BucketName, pbsDatastore.S3Bucket)
	assert.Contains(t, pbsDatastore.Backend, "type=s3")
	assert.Contains(t, pbsDatastore.Backend, fmt.Sprintf("client=%s", pbsConfig["id"]))
	assert.Contains(t, pbsDatastore.Backend, fmt.Sprintf("bucket=%s", provider.BucketName))

	// Step 11: Test that S3 bucket has PBS datastore structure created
	// PBS creates a directory structure in the S3 bucket with the datastore name as prefix
	listResp, err := provider.S3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(provider.BucketName),
		Prefix: aws.String(fmt.Sprintf("%s/", datastoreName)),
	})
	require.NoError(t, err, "Failed to list objects in S3 bucket")

	// PBS should have created some initial structure in the bucket
	// We don't expect specific files immediately, but the API call should succeed
	t.Logf("S3 bucket contains %d objects with datastore prefix %s/", len(listResp.Contents), datastoreName)

	t.Logf("Successfully created and verified S3 datastore for %s provider", provider.Name)
	t.Logf("  - Endpoint ID: %s", pbsConfig["id"])
	t.Logf("  - Datastore Name: %s", datastoreName)
	t.Logf("  - S3 Bucket: %s", provider.BucketName)
	t.Logf("  - Local Cache Path: /datastore/%s-cache", datastoreName)

	// Terraform destroy will be handled by defer
}
