/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package endpoints provides Terraform data sources for PBS endpoints
package endpoints

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/endpoints"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &s3EndpointDataSource{}
	_ datasource.DataSourceWithConfigure = &s3EndpointDataSource{}
)

// NewS3EndpointDataSource is a helper function to simplify the provider implementation.
func NewS3EndpointDataSource() datasource.DataSource {
	return &s3EndpointDataSource{}
}

// s3EndpointDataSource is the data source implementation.
type s3EndpointDataSource struct {
	client *pbs.Client
}

// s3EndpointDataSourceModel maps the data source schema data.
type s3EndpointDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	AccessKey      types.String `tfsdk:"access_key"`
	Endpoint       types.String `tfsdk:"endpoint"`
	Region         types.String `tfsdk:"region"`
	Fingerprint    types.String `tfsdk:"fingerprint"`
	Port           types.Int64  `tfsdk:"port"`
	PathStyle      types.Bool   `tfsdk:"path_style"`
	ProviderQuirks types.Set    `tfsdk:"provider_quirks"`
}

// Metadata returns the data source type name.
func (d *s3EndpointDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_endpoint"
}

// Schema defines the schema for the data source.
func (d *s3EndpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads a specific S3 endpoint configuration from Proxmox Backup Server.",
		MarkdownDescription: "Reads a specific S3 endpoint configuration from Proxmox Backup Server.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Unique identifier for the S3 endpoint.",
				MarkdownDescription: "Unique identifier for the S3 endpoint.",
				Required:            true,
			},
			"access_key": schema.StringAttribute{
				Description:         "Access key for S3 object store.",
				MarkdownDescription: "Access key for S3 object store.",
				Computed:            true,
				Sensitive:           true,
			},
			"endpoint": schema.StringAttribute{
				Description:         "Endpoint to access S3 object store.",
				MarkdownDescription: "Endpoint to access S3 object store.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				Description:         "Region to access S3 object store.",
				MarkdownDescription: "Region to access S3 object store.",
				Computed:            true,
			},
			"fingerprint": schema.StringAttribute{
				Description:         "X509 certificate fingerprint (sha256).",
				MarkdownDescription: "X509 certificate fingerprint (sha256).",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				Description:         "Port to access S3 object store.",
				MarkdownDescription: "Port to access S3 object store.",
				Computed:            true,
			},
			"path_style": schema.BoolAttribute{
				Description:         "Use path style bucket addressing over vhost style.",
				MarkdownDescription: "Use path style bucket addressing over vhost style.",
				Computed:            true,
			},
			"provider_quirks": schema.SetAttribute{
				Description:         "S3 provider-specific quirks.",
				MarkdownDescription: "S3 provider-specific quirks. Example: `['skip-if-none-match-header']` for Backblaze B2 compatibility.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *s3EndpointDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.DataSource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.DataSource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = cfg.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *s3EndpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state s3EndpointDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get S3 endpoint from API
	endpoint, err := d.client.Endpoints.GetS3Endpoint(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading S3 Endpoint",
			fmt.Sprintf("Could not read S3 endpoint %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Map API response to state
	if err := s3EndpointToState(ctx, endpoint, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting S3 Endpoint",
			fmt.Sprintf("Could not convert S3 endpoint to state: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// s3EndpointToState converts an S3 endpoint struct to Terraform state
func s3EndpointToState(ctx context.Context, endpoint *endpoints.S3Endpoint, state *s3EndpointDataSourceModel) error {
	state.ID = types.StringValue(endpoint.ID)
	state.AccessKey = types.StringValue(endpoint.AccessKey)
	state.Endpoint = types.StringValue(endpoint.Endpoint)
	state.Region = stringValueOrNull(endpoint.Region)
	state.Fingerprint = stringValueOrNull(endpoint.Fingerprint)
	state.Port = intValueOrNull(endpoint.Port)
	state.PathStyle = boolValueOrNull(endpoint.PathStyle)

	// Convert provider quirks slice to set
	if len(endpoint.ProviderQuirks) > 0 {
		quirks := make([]attr.Value, len(endpoint.ProviderQuirks))
		for i, q := range endpoint.ProviderQuirks {
			quirks[i] = types.StringValue(q)
		}
		quirkSet, diags := types.SetValue(types.StringType, quirks)
		if diags.HasError() {
			return fmt.Errorf("failed to convert provider quirks to set")
		}
		state.ProviderQuirks = quirkSet
	} else {
		state.ProviderQuirks = types.SetNull(types.StringType)
	}

	return nil
}

// Helper functions

func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func intValueOrNull(ptr *int) types.Int64 {
	if ptr == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*ptr))
}

func boolValueOrNull(ptr *bool) types.Bool {
	if ptr == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*ptr)
}
