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
	_ datasource.DataSource              = &syncJobsDataSource{}
	_ datasource.DataSourceWithConfigure = &syncJobsDataSource{}
)

// NewSyncJobsDataSource is a helper function to simplify the provider implementation.
func NewSyncJobsDataSource() datasource.DataSource {
	return &syncJobsDataSource{}
}

// syncJobsDataSource is the data source implementation.
type syncJobsDataSource struct {
	client *pbs.Client
}

// syncJobsDataSourceModel maps the data source schema data.
type syncJobsDataSourceModel struct {
	Store  types.String                 `tfsdk:"store"`
	Remote types.String                 `tfsdk:"remote"`
	Jobs   []syncJobDataSourceModel `tfsdk:"jobs"`
}

// Metadata returns the data source type name.
func (d *syncJobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_jobs"
}

// Schema defines the schema for the data source.
func (d *syncJobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all PBS sync jobs, optionally filtered by datastore or remote.",
		MarkdownDescription: "Lists all PBS sync jobs, optionally filtered by datastore or remote.",
		Attributes: map[string]schema.Attribute{
			"store": schema.StringAttribute{
				Description:         "Filter sync jobs by target datastore name (optional).",
				MarkdownDescription: "Filter sync jobs by target datastore name (optional).",
				Optional:            true,
			},
			"remote": schema.StringAttribute{
				Description:         "Filter sync jobs by remote name (optional).",
				MarkdownDescription: "Filter sync jobs by remote name (optional).",
				Optional:            true,
			},
			"jobs": schema.ListNestedAttribute{
				Description:         "List of sync jobs.",
				MarkdownDescription: "List of sync jobs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"store": schema.StringAttribute{
							Computed: true,
						},
						"schedule": schema.StringAttribute{
							Computed: true,
						},
						"remote": schema.StringAttribute{
							Computed: true,
						},
						"remote_store": schema.StringAttribute{
							Computed: true,
						},
						"remote_namespace": schema.StringAttribute{
							Computed: true,
						},
						"namespace": schema.StringAttribute{
							Computed: true,
						},
						"max_depth": schema.Int64Attribute{
							Computed: true,
						},
						"group_filter": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"remove_vanished": schema.BoolAttribute{
							Computed: true,
						},
						"comment": schema.StringAttribute{
							Computed: true,
						},
						"digest": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *syncJobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *syncJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state syncJobsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all sync jobs from API
	jobs, err := d.client.Jobs.ListSyncJobs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Sync Jobs",
			fmt.Sprintf("Could not list sync jobs: %s", err.Error()),
		)
		return
	}

	// Apply filters
	storeFilter := ""
	if !state.Store.IsNull() && !state.Store.IsUnknown() {
		storeFilter = state.Store.ValueString()
	}
	remoteFilter := ""
	if !state.Remote.IsNull() && !state.Remote.IsUnknown() {
		remoteFilter = state.Remote.ValueString()
	}

	// Map API response to state
	state.Jobs = make([]syncJobDataSourceModel, 0)
	for _, job := range jobs {
		if storeFilter != "" && job.Store != storeFilter {
			continue
		}
		if remoteFilter != "" && job.Remote != remoteFilter {
			continue
		}
		
		var jobModel syncJobDataSourceModel
		jobModel.ID = types.StringValue(job.ID)
		jobModel.Store = types.StringValue(job.Store)
		jobModel.Schedule = types.StringValue(job.Schedule)
		jobModel.Remote = types.StringValue(job.Remote)
		jobModel.RemoteStore = types.StringValue(job.RemoteStore)
		jobModel.RemoteNamespace = stringToValue(job.RemoteNamespace)
		jobModel.Namespace = stringToValue(job.Namespace)
		jobModel.MaxDepth = intPtrToValue(job.MaxDepth)
		jobModel.RemoveVanished = boolPtrToValue(job.RemoveVanished)
		jobModel.Comment = stringToValue(job.Comment)
		jobModel.Digest = types.StringValue(job.Digest)

		// Convert group filter
		if len(job.GroupFilter) > 0 {
			groupFilterValues := make([]types.String, 0, len(job.GroupFilter))
			for _, filter := range job.GroupFilter {
				groupFilterValues = append(groupFilterValues, types.StringValue(filter))
			}
			listValue, diag := types.ListValueFrom(ctx, types.StringType, groupFilterValues)
			resp.Diagnostics.Append(diag...)
			jobModel.GroupFilter = listValue
		} else {
			jobModel.GroupFilter = types.ListNull(types.StringType)
		}

		state.Jobs = append(state.Jobs, jobModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
