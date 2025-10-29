/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package metrics

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &metricsServersDataSource{}
	_ datasource.DataSourceWithConfigure = &metricsServersDataSource{}
)

// NewMetricsServersDataSource is a helper function to simplify the provider implementation.
func NewMetricsServersDataSource() datasource.DataSource {
	return &metricsServersDataSource{}
}

// metricsServersDataSource is the data source implementation.
type metricsServersDataSource struct {
	client *pbs.Client
}

// metricsServersDataSourceModel maps the data source schema data.
type metricsServersDataSourceModel struct {
	Servers []metricsServerModel `tfsdk:"servers"`
}

// metricsServerModel represents a single metrics server in the list
type metricsServerModel struct {
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	URL          types.String `tfsdk:"url"`
	Server       types.String `tfsdk:"server"`
	Port         types.Int64  `tfsdk:"port"`
	Enable       types.Bool   `tfsdk:"enable"`
	MTU          types.Int64  `tfsdk:"mtu"`
	Organization types.String `tfsdk:"organization"`
	Bucket       types.String `tfsdk:"bucket"`
	MaxBodySize  types.Int64  `tfsdk:"max_body_size"`
	VerifyTLS    types.Bool   `tfsdk:"verify_tls"`
	Comment      types.String `tfsdk:"comment"`
}

// Metadata returns the data source type name.
func (d *metricsServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_servers"
}

// Schema defines the schema for the data source.
func (d *metricsServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all metrics server configurations from Proxmox Backup Server.",
		MarkdownDescription: "Lists all metrics server configurations from Proxmox Backup Server.",

		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Description:         "List of metrics servers.",
				MarkdownDescription: "List of metrics servers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique name identifier for the metrics server.",
							MarkdownDescription: "The unique name identifier for the metrics server.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							Description:         "The type of metrics server.",
							MarkdownDescription: "The type of metrics server (e.g., `influxdb-udp`, `influxdb-http`).",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							Description:         "Full URL for InfluxDB HTTP.",
							MarkdownDescription: "Full URL for InfluxDB HTTP.",
							Computed:            true,
						},
						"server": schema.StringAttribute{
							Description:         "The server address (hostname or IP).",
							MarkdownDescription: "The server address extracted from URL or Host field.",
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							Description:         "The server port.",
							MarkdownDescription: "The server port extracted from URL or Host field.",
							Computed:            true,
						},
						"enable": schema.BoolAttribute{
							Description:         "Whether this metrics server is enabled.",
							MarkdownDescription: "Whether metrics export to this server is enabled.",
							Computed:            true,
						},
						"mtu": schema.Int64Attribute{
							Description:         "MTU for the metrics connection.",
							MarkdownDescription: "Maximum transmission unit for the metrics connection.",
							Computed:            true,
						},
						"organization": schema.StringAttribute{
							Description:         "InfluxDB organization (InfluxDB HTTP only).",
							MarkdownDescription: "InfluxDB organization name.",
							Computed:            true,
						},
						"bucket": schema.StringAttribute{
							Description:         "InfluxDB bucket name (InfluxDB HTTP only).",
							MarkdownDescription: "InfluxDB bucket name.",
							Computed:            true,
						},
						"max_body_size": schema.Int64Attribute{
							Description:         "Maximum body size for HTTP requests in bytes (InfluxDB HTTP only).",
							MarkdownDescription: "Maximum body size for HTTP requests in bytes.",
							Computed:            true,
						},
						"verify_tls": schema.BoolAttribute{
							Description:         "Verify TLS certificate for HTTPS connections (InfluxDB HTTP only).",
							MarkdownDescription: "Whether to verify TLS certificate for HTTPS connections.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							Description:         "A comment describing this metrics server.",
							MarkdownDescription: "A comment describing this metrics server configuration.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *metricsServersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *metricsServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state metricsServersDataSourceModel

	// Get all metrics servers from API
	serversList, err := d.client.Metrics.ListMetricsServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Metrics Servers",
			fmt.Sprintf("Could not list metrics servers: %s", err.Error()),
		)
		return
	}

	// Map API response to state
	state.Servers = make([]metricsServerModel, 0, len(serversList))
	for _, srv := range serversList {
		server := metricsServerModel{
			Name:    types.StringValue(srv.Name),
			Type:    types.StringValue(string(srv.Type)),
			Comment: stringValueOrNull(srv.Comment),
			Enable:  boolValueOrNull(srv.Enable),
			URL:     stringValueOrNull(srv.URL),
			Server:  stringValueOrNull(srv.Server),
		}

		if srv.Port > 0 {
			server.Port = types.Int64Value(int64(srv.Port))
		} else {
			server.Port = types.Int64Null()
		}

		// Type-specific fields
		server.MTU = intValueOrNull(srv.MTU)
		server.Organization = stringValueOrNull(srv.Organization)
		server.Bucket = stringValueOrNull(srv.Bucket)
		server.MaxBodySize = intValueOrNull(srv.MaxBodySize)
		server.VerifyTLS = boolValueOrNull(srv.VerifyTLS)

		state.Servers = append(state.Servers, server)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
