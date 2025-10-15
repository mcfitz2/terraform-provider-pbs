/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package fwprovider implements the Terraform provider for Proxmox Backup Server
package fwprovider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/fwprovider/resources/datastores"
	"github.com/micah/terraform-provider-pbs/fwprovider/resources/endpoints"
	"github.com/micah/terraform-provider-pbs/fwprovider/resources/jobs"
	"github.com/micah/terraform-provider-pbs/fwprovider/resources/metrics"
	"github.com/micah/terraform-provider-pbs/fwprovider/resources/notifications"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &pbsProvider{}
)

// pbsProvider defines the provider implementation.
type pbsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// pbsProviderModel maps provider schema data.
type pbsProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Insecure types.Bool   `tfsdk:"insecure"`
	APIToken types.String `tfsdk:"api_token"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

// New creates a new provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pbsProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *pbsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pbs"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *pbsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Proxmox Backup Server provider is used to interact with the resources " +
			"supported by Proxmox Backup Server. The provider needs to be configured with the " +
			"proper credentials before it can be used.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The endpoint URL for the Proxmox Backup Server API (e.g., https://pbs.example.com:8007)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"insecure": schema.BoolAttribute{
				Description: "Whether to skip the TLS verification step. Defaults to false.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "The API token for authentication (format: user@realm:token_name=token_value)",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "The username for authentication (alternative to API token)",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for authentication (used with username)",
				Optional:    true,
				Sensitive:   true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout for API requests in seconds. Defaults to 30.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a PBS API client for data sources and resources.
func (p *pbsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring PBS provider")

	// Retrieve provider data from configuration
	var cfg pbsProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	endpoint := os.Getenv("PBS_ENDPOINT")
	insecure := os.Getenv("PBS_INSECURE") == "true"
	apiToken := os.Getenv("PBS_API_TOKEN")
	username := os.Getenv("PBS_USERNAME")
	password := os.Getenv("PBS_PASSWORD")
	timeout := int64(30)

	if !cfg.Endpoint.IsNull() {
		endpoint = cfg.Endpoint.ValueString()
	}

	if !cfg.Insecure.IsNull() {
		insecure = cfg.Insecure.ValueBool()
	}

	if !cfg.APIToken.IsNull() {
		apiToken = cfg.APIToken.ValueString()
	}

	if !cfg.Username.IsNull() {
		username = cfg.Username.ValueString()
	}

	if !cfg.Password.IsNull() {
		password = cfg.Password.ValueString()
	}

	if !cfg.Timeout.IsNull() {
		timeout = cfg.Timeout.ValueInt64()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing PBS API Endpoint",
			"The provider cannot create the PBS API client as there is a missing or empty value for the PBS API endpoint. "+
				"Set the endpoint value in the configuration or use the PBS_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiToken == "" && (username == "" || password == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Authentication",
			"The provider cannot create the PBS API client as there are missing authentication credentials. "+
				"Either set the api_token value in the configuration or use the PBS_API_TOKEN environment variable, "+
				"or provide both username and password (with PBS_USERNAME and PBS_PASSWORD environment variables).",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate endpoint URL
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Invalid Endpoint URL",
			fmt.Sprintf("The endpoint URL must start with http:// or https://. Got: %s", endpoint),
		)
		return
	}

	// Create API credentials
	creds := api.Credentials{
		Username: username,
		Password: password,
		APIToken: apiToken,
	}

	// Create API client options
	opts := api.ClientOptions{
		Endpoint: endpoint,
		Insecure: insecure,
		Timeout:  time.Duration(timeout) * time.Second,
	}

	// Create PBS client
	client, err := pbs.NewClient(creds, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create PBS API Client",
			"An unexpected error occurred when creating the PBS API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"PBS Client Error: "+err.Error(),
		)
		return
	}

	// Make the PBS client available during DataSource and Resource
	// type Configure methods.
	resourceConfig := &config.Resource{Client: client}
	datasourceConfig := &config.DataSource{Client: client}

	resp.DataSourceData = datasourceConfig
	resp.ResourceData = resourceConfig

	tflog.Info(ctx, "Configured PBS provider", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *pbsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Add data sources here as they are implemented
	}
}

// Resources defines the resources implemented in the provider.
func (p *pbsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Endpoints
		endpoints.NewS3EndpointResource,
		// Datastores
		datastores.NewDatastoreResource,
		// Metrics
		metrics.NewMetricsServerResource,
		// Notifications - Targets
		notifications.NewSMTPNotificationResource,
		notifications.NewGotifyNotificationResource,
		notifications.NewSendmailNotificationResource,
		notifications.NewWebhookNotificationResource,
		// Notifications - Routing
		notifications.NewNotificationMatcherResource,
		// Jobs
		jobs.NewPruneJobResource,
		jobs.NewSyncJobResource,
		jobs.NewVerifyJobResource,
	}
}
