/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

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
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &s3EndpointsDataSource{}
	_ datasource.DataSourceWithConfigure = &s3EndpointsDataSource{}
)

// NewS3EndpointsDataSource is a helper function to simplify the provider implementation.
func NewS3EndpointsDataSource() datasource.DataSource {
	return &s3EndpointsDataSource{}
}

// s3EndpointsDataSource is the data source implementation.
type s3EndpointsDataSource struct {
	client *pbs.Client
}

// s3EndpointsDataSourceModel maps the data source schema data.
type s3EndpointsDataSourceModel struct {
	Endpoints []s3EndpointModel `tfsdk:"endpoints"`
}

// s3EndpointModel represents a single S3 endpoint in the list
type s3EndpointModel struct {
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
func (d *s3EndpointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_endpoints"
}

// Schema defines the schema for the data source.
func (d *s3EndpointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all S3 endpoint configurations from Proxmox Backup Server.",
		MarkdownDescription: "Lists all S3 endpoint configurations from Proxmox Backup Server.",

		Attributes: map[string]schema.Attribute{
			"endpoints": schema.ListNestedAttribute{
				Description:         "List of S3 endpoints.",
				MarkdownDescription: "List of S3 endpoints.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Unique identifier for the S3 endpoint.",
							MarkdownDescription: "Unique identifier for the S3 endpoint.",
							Computed:            true,
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
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *s3EndpointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *s3EndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state s3EndpointsDataSourceModel

	// Get all S3 endpoints from API
	endpointsList, err := d.client.Endpoints.ListS3Endpoints(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading S3 Endpoints",
			fmt.Sprintf("Could not list S3 endpoints: %s", err.Error()),
		)
		return
	}

	// Map API response to state
	state.Endpoints = make([]s3EndpointModel, 0, len(endpointsList))
	for _, ep := range endpointsList {
		endpoint := s3EndpointModel{
			ID:          types.StringValue(ep.ID),
			AccessKey:   types.StringValue(ep.AccessKey),
			Endpoint:    types.StringValue(ep.Endpoint),
			Region:      stringValueOrNull(ep.Region),
			Fingerprint: stringValueOrNull(ep.Fingerprint),
			Port:        intValueOrNull(ep.Port),
			PathStyle:   boolValueOrNull(ep.PathStyle),
		}

		// Convert provider quirks slice to set
		if len(ep.ProviderQuirks) > 0 {
			quirks := make([]attr.Value, len(ep.ProviderQuirks))
			for i, q := range ep.ProviderQuirks {
				quirks[i] = types.StringValue(q)
			}
			quirkSet, diags := types.SetValue(types.StringType, quirks)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			endpoint.ProviderQuirks = quirkSet
		} else {
			endpoint.ProviderQuirks = types.SetNull(types.StringType)
		}

		state.Endpoints = append(state.Endpoints, endpoint)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
