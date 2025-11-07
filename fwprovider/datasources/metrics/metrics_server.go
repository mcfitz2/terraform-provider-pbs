/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package metrics provides Terraform data sources for PBS metrics server configurations
package metrics

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/metrics"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &metricsServerDataSource{}
	_ datasource.DataSourceWithConfigure = &metricsServerDataSource{}
)

// NewMetricsServerDataSource is a helper function to simplify the provider implementation.
func NewMetricsServerDataSource() datasource.DataSource {
	return &metricsServerDataSource{}
}

// metricsServerDataSource is the data source implementation.
type metricsServerDataSource struct {
	client *pbs.Client
}

// metricsServerDataSourceModel maps the data source schema data.
type metricsServerDataSourceModel struct {
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
func (d *metricsServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_server"
}

// Schema defines the schema for the data source.
func (d *metricsServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads a specific metrics server configuration from Proxmox Backup Server.",
		MarkdownDescription: "Reads a specific metrics server configuration from Proxmox Backup Server.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique name identifier for the metrics server.",
				MarkdownDescription: "The unique name identifier for the metrics server.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				Description:         "The type of metrics server.",
				MarkdownDescription: "The type of metrics server. Valid values: `influxdb-udp`, `influxdb-http`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("influxdb-udp", "influxdb-http"),
				},
			},
			"url": schema.StringAttribute{
				Description:         "Full URL for InfluxDB HTTP.",
				MarkdownDescription: "Full URL for InfluxDB HTTP (e.g., `http://influxdb:8086`). Only applicable for `influxdb-http` type.",
				Computed:            true,
			},
			"server": schema.StringAttribute{
				Description:         "The server address (hostname or IP).",
				MarkdownDescription: "The server address (hostname or IP) extracted from URL or Host field.",
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
				MarkdownDescription: "InfluxDB organization name. Only applicable for `influxdb-http` type.",
				Computed:            true,
			},
			"bucket": schema.StringAttribute{
				Description:         "InfluxDB bucket name (InfluxDB HTTP only).",
				MarkdownDescription: "InfluxDB bucket name. Only applicable for `influxdb-http` type.",
				Computed:            true,
			},
			"max_body_size": schema.Int64Attribute{
				Description:         "Maximum body size for HTTP requests in bytes (InfluxDB HTTP only).",
				MarkdownDescription: "Maximum body size for HTTP requests in bytes. Only applicable for `influxdb-http` type.",
				Computed:            true,
			},
			"verify_tls": schema.BoolAttribute{
				Description:         "Verify TLS certificate for HTTPS connections (InfluxDB HTTP only).",
				MarkdownDescription: "Whether to verify TLS certificate for HTTPS connections. Only applicable for `influxdb-http` type.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this metrics server.",
				MarkdownDescription: "A comment describing this metrics server configuration.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *metricsServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *metricsServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state metricsServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get metrics server from API
	serverType := metrics.MetricsServerType(state.Type.ValueString())
	server, err := d.client.Metrics.GetMetricsServer(ctx, serverType, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Metrics Server",
			fmt.Sprintf("Could not read metrics server %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	// Map API response to state
	if err := metricsServerToState(server, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Metrics Server",
			fmt.Sprintf("Could not convert metrics server to state: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// metricsServerToState converts a metrics server struct to Terraform state
func metricsServerToState(server *metrics.MetricsServer, state *metricsServerDataSourceModel) error {
	state.Name = types.StringValue(server.Name)
	state.Type = types.StringValue(string(server.Type))
	state.Comment = stringValueOrNull(server.Comment)
	// PBS defaults enable to true when not specified
	state.Enable = boolValueWithDefault(server.Enable, true)

	// URL and parsed server/port fields
	state.URL = stringValueOrNull(server.URL)
	state.Server = stringValueOrNull(server.Server)
	if server.Port > 0 {
		state.Port = types.Int64Value(int64(server.Port))
	} else {
		state.Port = types.Int64Null()
	}

	// Type-specific fields
	if server.Type == metrics.MetricsServerTypeInfluxDBUDP {
		state.MTU = intValueOrNull(server.MTU)
	}

	if server.Type == metrics.MetricsServerTypeInfluxDBHTTP {
		state.Organization = stringValueOrNull(server.Organization)
		state.Bucket = stringValueOrNull(server.Bucket)
		state.MaxBodySize = intValueOrNull(server.MaxBodySize)
		state.VerifyTLS = boolValueOrNull(server.VerifyTLS)
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

func boolValueWithDefault(ptr *bool, defaultValue bool) types.Bool {
	if ptr == nil {
		return types.BoolValue(defaultValue)
	}
	return types.BoolValue(*ptr)
}
