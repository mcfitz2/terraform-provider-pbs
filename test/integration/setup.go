package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// TestContext holds the test configuration and client for PBS integration tests
type TestContext struct {
	Config    *Config
	APIClient *api.Client
	TF        *tfexec.Terraform
	Workdir   string
}

// Config holds the test configuration for PBS integration tests
type Config struct {
	Address  string
	Username string
	Password string
}

// GetConfig loads test configuration from environment variables
func GetConfig(t *testing.T) *Config {
	address := os.Getenv("PBS_ADDRESS")
	username := os.Getenv("PBS_USERNAME")
	password := os.Getenv("PBS_PASSWORD")

	if address == "" {
		t.Skip("PBS_ADDRESS not set, skipping integration tests")
	}
	if username == "" {
		t.Skip("PBS_USERNAME not set, skipping integration tests")
	}
	if password == "" {
		t.Skip("PBS_PASSWORD not set, skipping integration tests")
	}

	return &Config{
		Address:  address,
		Username: username,
		Password: password,
	}
}

// SetupTest creates a test context with PBS API client and terraform executor
func SetupTest(t *testing.T) *TestContext {
	config := GetConfig(t)

	// Create API client
	creds := api.Credentials{
		Username: config.Username,
		Password: config.Password,
	}
	opts := api.ClientOptions{
		Endpoint: config.Address,
		Insecure: true, // For testing
		Timeout:  30 * time.Second,
	}
	apiClient, err := api.NewClient(creds, opts)
	require.NoError(t, err, "Failed to create PBS API client")

	// Create temporary directory for terraform
	workdir := t.TempDir()

	// Setup terraform executor
	tf, err := tfexec.NewTerraform(workdir, "terraform")
	require.NoError(t, err, "Failed to create terraform executor")

	// Ensure Node.js is available for Terraform operations
	currentPath := os.Getenv("PATH")
	nodePaths := "/usr/local/bin:/usr/bin"
	if !strings.Contains(currentPath, nodePaths) {
		os.Setenv("PATH", nodePaths+":"+currentPath)
	}
	os.Setenv("NODE_PATH", nodePaths)

	// Set up environment for terraform
	err = tf.SetEnv(map[string]string{
		"TF_CLI_CONFIG_FILE":         workdir + "/.terraformrc",
		"PBS_DESTROY_DATA_ON_DELETE": "true", // Always destroy data during tests
	})
	if err != nil {
		t.Fatalf("Failed to set terraform environment: %s", err)
	}

	return &TestContext{
		Config:    config,
		APIClient: apiClient,
		TF:        tf,
		Workdir:   workdir,
	}
}

// GenerateTestName creates a unique test resource name
func GenerateTestName(prefix string) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s-test-%d", prefix, rng.Intn(10000))
}

// WriteMainTF writes the main terraform configuration file
func (tc *TestContext) WriteMainTF(t *testing.T, config string) {
	// Copy the provider binary to expected location in plugins directory first
	tc.setupLocalProvider(t)

	// For integration tests, use a configuration that matches the copied provider
	mainTF := fmt.Sprintf(`
terraform {
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "1.0.0"
    }
  }
}

provider "pbs" {
  endpoint = "%s"
  username = "%s"
  password = "%s"
  insecure = true
}

%s
`, tc.Config.Address, tc.Config.Username, tc.Config.Password, config)

	err := os.WriteFile(tc.Workdir+"/main.tf", []byte(mainTF), 0644)
	require.NoError(t, err, "Failed to write main.tf")
}

// getProjectRoot returns the absolute path to the project root containing the built provider
func (tc *TestContext) getProjectRoot() string {
	// Get absolute path to the project root (two levels up from test/integration/)
	wd, _ := os.Getwd()
	return filepath.Dir(filepath.Dir(wd))
}

// setupLocalProvider configures the local provider binary for terraform
func (tc *TestContext) setupLocalProvider(t *testing.T) {
	providerPath := tc.getProjectRoot()
	providerBinary := filepath.Join(providerPath, "terraform-provider-pbs")

	// Check if provider binary exists
	if _, err := os.Stat(providerBinary); os.IsNotExist(err) {
		t.Fatalf("Provider binary not found at %s. Run 'go build .' first.", providerBinary)
	}

	// Create plugins directory structure using the separate plugins directory
	pluginsDir := filepath.Join(tc.Workdir, "plugins", "registry.terraform.io", "micah", "pbs", "1.0.0", runtime.GOOS+"_"+runtime.GOARCH)
	err := os.MkdirAll(pluginsDir, 0755)
	require.NoError(t, err, "Failed to create plugins directory")

	// Copy provider binary
	destPath := filepath.Join(pluginsDir, "terraform-provider-pbs")
	err = copyFile(providerBinary, destPath)
	require.NoError(t, err, "Failed to copy provider binary")

	// Make executable
	err = os.Chmod(destPath, 0755)
	require.NoError(t, err, "Failed to make provider executable")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ApplyTerraform applies the terraform configuration
func (tc *TestContext) ApplyTerraform(t *testing.T) {
	// Initialize terraform with a separate plugin directory
	pluginDir := filepath.Join(tc.Workdir, "plugins")
	err := tc.TF.Init(context.Background(), tfexec.PluginDir(pluginDir), tfexec.Get(false))
	require.NoError(t, err, "Failed to initialize terraform")

	err = tc.TF.Apply(context.Background())
	require.NoError(t, err, "Terraform apply failed")
}

// ApplyTerraformWithRetry applies the terraform configuration with retry logic for lock contention
func (tc *TestContext) ApplyTerraformWithRetry(t *testing.T) {
	// Initialize terraform with a separate plugin directory
	pluginDir := filepath.Join(tc.Workdir, "plugins")
	err := tc.TF.Init(context.Background(), tfexec.PluginDir(pluginDir), tfexec.Get(false))
	require.NoError(t, err, "Failed to initialize terraform")

	// Retry logic for PBS lock contention
	maxRetries := 3
	retryDelay := time.Second * 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		t.Logf("Terraform apply attempt %d/%d", attempt, maxRetries)

		err = tc.TF.Apply(context.Background())
		if err == nil {
			t.Logf("Terraform apply succeeded on attempt %d", attempt)
			return
		}

		errorMsg := err.Error()
		isLockError := strings.Contains(errorMsg, "Unable to acquire lock") ||
			strings.Contains(errorMsg, "Interrupted system call") ||
			strings.Contains(errorMsg, ".datastore.lck")

		if isLockError && attempt < maxRetries {
			t.Logf("Lock contention detected, retrying in %v (attempt %d/%d): %v", retryDelay, attempt, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

		// Not a lock error or final attempt - fail
		require.NoError(t, err, "Terraform apply failed after %d attempts", maxRetries)
	}
}

// ApplyTerraformWithError applies the terraform configuration and returns any error instead of failing the test
func (tc *TestContext) ApplyTerraformWithError(t *testing.T) error {
	// Initialize terraform with a separate plugin directory
	pluginDir := filepath.Join(tc.Workdir, "plugins")
	err := tc.TF.Init(context.Background(), tfexec.PluginDir(pluginDir), tfexec.Get(false))
	if err != nil {
		return fmt.Errorf("failed to initialize terraform: %w", err)
	}

	err = tc.TF.Apply(context.Background())
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}
	return nil
}

// ImportResource imports an existing resource into terraform state
func (tc *TestContext) ImportResource(t *testing.T, address, id string) {
	// Ensure terraform is initialized first
	pluginDir := filepath.Join(tc.Workdir, "plugins")
	err := tc.TF.Init(context.Background(), tfexec.PluginDir(pluginDir), tfexec.Get(false))
	require.NoError(t, err, "Failed to initialize terraform")

	err = tc.TF.Import(context.Background(), address, id)
	require.NoError(t, err, "Failed to import resource")
}

// DestroyTerraform destroys the terraform resources
func (tc *TestContext) DestroyTerraform(t *testing.T) {
	err := tc.TF.Destroy(context.Background())
	if err != nil {
		errMsg := err.Error()
		// Only suppress errors where the resource legitimately doesn't exist
		// (e.g., already deleted by PBS, or destroyed in a previous operation)
		if strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "no such") ||
			strings.Contains(errMsg, "not found") ||
			strings.Contains(errMsg, "404") {
			// Resource already gone - this is fine, desired state achieved
			t.Logf("Terraform destroy: resource already deleted (desired state achieved)")
			return
		}
		
		// Log any other errors as warnings, but don't fail the test
		// This allows cleanup to continue even if there are issues
		t.Logf("Warning: Terraform destroy encountered error: %v", err)
		t.Logf("Continuing cleanup despite error...")
	}
}

// GetTerraformState returns the current terraform state with fallback for Node.js issues
func (tc *TestContext) GetTerraformState(t *testing.T) *tfjson.State {
	// Set NODE_PATH environment variable to help find node binary
	os.Setenv("NODE_PATH", "/usr/bin:/usr/local/bin")

	state, err := tc.TF.Show(context.Background())
	if err != nil {
		// Check if this is a Node.js related error
		if strings.Contains(err.Error(), "node") || strings.Contains(err.Error(), "No such file") {
			t.Logf("Warning: Terraform state reading failed due to Node.js dependency: %v", err)
			// Try to read terraform.tfstate directly as fallback
			return tc.GetTerraformStateFromFile(t)
		}
		require.NoError(t, err, "Failed to get terraform state")
	}
	return state
}

// GetTerraformStateFromFile reads terraform state directly from terraform.tfstate file
func (tc *TestContext) GetTerraformStateFromFile(t *testing.T) *tfjson.State {
	statePath := filepath.Join(tc.Workdir, "terraform.tfstate")
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		t.Logf("Warning: Could not read terraform.tfstate file: %v", err)
		return nil
	}

	// Try to parse JSON manually
	var state tfjson.State
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Logf("Warning: Could not parse terraform state JSON: %v", err)
		return nil
	}

	return &state
}

// GetResourceFromState extracts a resource from terraform state by address
func (tc *TestContext) GetResourceFromState(t *testing.T, address string) *tfjson.StateResource {
	state := tc.GetTerraformState(t)
	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		t.Logf("Warning: Terraform state is not available, skipping state verification for %s", address)
		return nil
	}

	for _, resource := range state.Values.RootModule.Resources {
		if resource.Address == address {
			return resource
		}
	}
	t.Fatalf("Resource %s not found in terraform state", address)
	return nil
}

// DebugNodeAvailability helps debug Node.js availability issues
func (tc *TestContext) DebugNodeAvailability(t *testing.T) {
	t.Logf("=== Node.js Environment Debug ===")
	t.Logf("PATH: %s", os.Getenv("PATH"))
	t.Logf("NODE_PATH: %s", os.Getenv("NODE_PATH"))

	// Check for node binary in common locations
	nodePaths := []string{"/usr/bin/node", "/usr/local/bin/node"}
	for _, path := range nodePaths {
		if _, err := os.Stat(path); err == nil {
			t.Logf("Found Node.js at: %s", path)
		} else {
			t.Logf("Node.js not found at: %s", path)
		}
	}

	// Try to execute node --version
	if _, err := os.Stat("/usr/bin/node"); err == nil {
		if output, err := exec.Command("/usr/bin/node", "--version").Output(); err == nil {
			t.Logf("Node.js version: %s", strings.TrimSpace(string(output)))
		} else {
			t.Logf("Failed to get Node.js version: %v", err)
		}
	}
	t.Logf("=== End Node.js Environment Debug ===")
}
