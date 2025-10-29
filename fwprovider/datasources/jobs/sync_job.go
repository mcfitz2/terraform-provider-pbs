/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package jobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
)

var (
	_ datasource.DataSource              = &syncJobDataSource{}
	_ datasource.DataSourceWithConfigure = &syncJobDataSource{}
)

// NewSyncJobDataSource is a helper function to simplify the provider implementation.
func NewSyncJobDataSource() datasource.DataSource {
	return &syncJobDataSource{}
}

// syncJobDataSource is the data source implementation.
type syncJobDataSource struct {
	client *pbs.Client
}

// syncJobDataSourceModel maps the data source schema data.
type syncJobDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Store           types.String `tfsdk:"store"`
	Schedule        types.String `tfsdk:"schedule"`
	Remote          types.String `tfsdk:"remote"`
	RemoteStore     types.String `tfsdk:"remote_store"`
	RemoteNamespace types.String `tfsdk:"remote_namespace"`
	Namespace       types.String `tfsdk:"namespace"`
	MaxDepth        types.Int64  `tfsdk:"max_depth"`
	GroupFilter     types.List   `tfsdk:"group_filter"`
	RemoveVanished  types.Bool   `tfsdk:"remove_vanished"`
	Comment         types.String `tfsdk:"comment"`
	Digest          types.String `tfsdk:"digest"`
}

// Metadata returns the data source type name.
func (d *syncJobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_job"
}

// Schema defines the schema for the data source.
func (d *syncJobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads information about an existing PBS sync job.",
		MarkdownDescription: "Reads information about an existing PBS sync job configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier for the sync job.",
				MarkdownDescription: "The unique identifier for the sync job.",
				Required:            true,
			},
			"store": schema.StringAttribute{
				Description:         "The target datastore name.",
				MarkdownDescription: "The target datastore name.",
				Computed:            true,
			},
			"schedule": schema.StringAttribute{
				Description:         "When the sync job runs.",
				MarkdownDescription: "When the sync job runs.",
				Computed:            true,
			},
			"remote": schema.StringAttribute{
				Description:         "The remote server name.",
				MarkdownDescription: "The remote server name.",
				Computed:            true,
			},
			"remote_store": schema.StringAttribute{
				Description:         "The remote datastore name.",
				MarkdownDescription: "The remote datastore name.",
				Computed:            true,
			},
			"remote_namespace": schema.StringAttribute{
				Description:         "The remote namespace.",
				MarkdownDescription: "The remote namespace.",
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				Description:         "Local namespace for synced backups.",
				MarkdownDescription: "Local namespace for synced backups.",
				Computed:            true,
			},
			"max_depth": schema.Int64Attribute{
				Description:         "Maximum depth for namespace traversal.",
				MarkdownDescription: "Maximum depth for namespace traversal.",
				Computed:            true,
			},
			"group_filter": schema.ListAttribute{
				Description:         "Filter backup groups to sync.",
				MarkdownDescription: "Filter backup groups to sync.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"remove_vanished": schema.BoolAttribute{
				Description:         "Remove backups vanished from remote.",
				MarkdownDescription: "Remove backups vanished from remote.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this sync job.",
				MarkdownDescription: "A comment describing this sync job.",
				Computed:            true,
			},
			"digest": schema.StringAttribute{
				Description:         "Opaque digest returned by PBS.",
				MarkdownDescription: "Opaque digest returned by PBS.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *syncJobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *syncJobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state syncJobDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get sync job from API
	job, err := d.client.Jobs.GetSyncJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Sync Job",
			fmt.Sprintf("Could not read sync job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Map API response to state
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)
	state.Remote = types.StringValue(job.Remote)
	state.RemoteStore = types.StringValue(job.RemoteStore)
	state.RemoteNamespace = stringToValue(job.RemoteNamespace)
	state.Namespace = stringToValue(job.Namespace)
	state.MaxDepth = intPtrToValue(job.MaxDepth)
	state.RemoveVanished = boolPtrToValue(job.RemoveVanished)
	state.Comment = stringToValue(job.Comment)
	state.Digest = types.StringValue(job.Digest)

	// Convert group filter
	if len(job.GroupFilter) > 0 {
		groupFilterValues := make([]types.String, 0, len(job.GroupFilter))
		for _, filter := range job.GroupFilter {
			groupFilterValues = append(groupFilterValues, types.StringValue(filter))
		}
		listValue, diag := types.ListValueFrom(ctx, types.StringType, groupFilterValues)
		resp.Diagnostics.Append(diag...)
		state.GroupFilter = listValue
	} else {
		state.GroupFilter = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
