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
	_ datasource.DataSource              = &pruneJobsDataSource{}
	_ datasource.DataSourceWithConfigure = &pruneJobsDataSource{}
)

// NewPruneJobsDataSource is a helper function to simplify the provider implementation.
func NewPruneJobsDataSource() datasource.DataSource {
	return &pruneJobsDataSource{}
}

// pruneJobsDataSource is the data source implementation.
type pruneJobsDataSource struct {
	client *pbs.Client
}

// pruneJobsDataSourceModel maps the data source schema data.
type pruneJobsDataSourceModel struct {
	Store types.String                  `tfsdk:"store"`
	Jobs  []pruneJobDataSourceModel `tfsdk:"jobs"`
}

// Metadata returns the data source type name.
func (d *pruneJobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prune_jobs"
}

// Schema defines the schema for the data source.
func (d *pruneJobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all PBS prune jobs, optionally filtered by datastore.",
		MarkdownDescription: "Lists all PBS prune jobs, optionally filtered by datastore.",
		Attributes: map[string]schema.Attribute{
			"store": schema.StringAttribute{
				Description:         "Filter prune jobs by datastore name (optional).",
				MarkdownDescription: "Filter prune jobs by datastore name (optional).",
				Optional:            true,
			},
			"jobs": schema.ListNestedAttribute{
				Description:         "List of prune jobs.",
				MarkdownDescription: "List of prune jobs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "The unique identifier for the prune job.",
							MarkdownDescription: "The unique identifier for the prune job.",
							Computed:            true,
						},
						"store": schema.StringAttribute{
							Description:         "The datastore name where pruning is performed.",
							MarkdownDescription: "The datastore name where pruning is performed.",
							Computed:            true,
						},
						"schedule": schema.StringAttribute{
							Description:         "When the prune job runs.",
							MarkdownDescription: "When the prune job runs.",
							Computed:            true,
						},
						"keep_last": schema.Int64Attribute{
							Description:         "Keep the last N backup snapshots.",
							MarkdownDescription: "Keep the last N backup snapshots.",
							Computed:            true,
						},
						"keep_hourly": schema.Int64Attribute{
							Description:         "Keep hourly backups for the last N hours.",
							MarkdownDescription: "Keep hourly backups for the last N hours.",
							Computed:            true,
						},
						"keep_daily": schema.Int64Attribute{
							Description:         "Keep daily backups for the last N days.",
							MarkdownDescription: "Keep daily backups for the last N days.",
							Computed:            true,
						},
						"keep_weekly": schema.Int64Attribute{
							Description:         "Keep weekly backups for the last N weeks.",
							MarkdownDescription: "Keep weekly backups for the last N weeks.",
							Computed:            true,
						},
						"keep_monthly": schema.Int64Attribute{
							Description:         "Keep monthly backups for the last N months.",
							MarkdownDescription: "Keep monthly backups for the last N months.",
							Computed:            true,
						},
						"keep_yearly": schema.Int64Attribute{
							Description:         "Keep yearly backups for the last N years.",
							MarkdownDescription: "Keep yearly backups for the last N years.",
							Computed:            true,
						},
						"max_depth": schema.Int64Attribute{
							Description:         "Maximum depth for namespace traversal.",
							MarkdownDescription: "Maximum depth for namespace traversal.",
							Computed:            true,
						},
						"namespace": schema.StringAttribute{
							Description:         "Namespace filter (regex).",
							MarkdownDescription: "Namespace filter (regex).",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							Description:         "A comment describing this prune job.",
							MarkdownDescription: "A comment describing this prune job.",
							Computed:            true,
						},
						"disable": schema.BoolAttribute{
							Description:         "Whether this prune job is disabled.",
							MarkdownDescription: "Whether this prune job is disabled.",
							Computed:            true,
						},
						"digest": schema.StringAttribute{
							Description:         "Opaque digest returned by PBS.",
							MarkdownDescription: "Opaque digest returned by PBS.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *pruneJobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *pruneJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pruneJobsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all prune jobs from API
	jobs, err := d.client.Jobs.ListPruneJobs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Prune Jobs",
			fmt.Sprintf("Could not list prune jobs: %s", err.Error()),
		)
		return
	}

	// Filter by store if specified
	storeFilter := ""
	if !state.Store.IsNull() && !state.Store.IsUnknown() {
		storeFilter = state.Store.ValueString()
	}

	// Map API response to state
	state.Jobs = make([]pruneJobDataSourceModel, 0)
	for _, job := range jobs {
		if storeFilter != "" && job.Store != storeFilter {
			continue
		}
		var jobModel pruneJobDataSourceModel
		pruneJobToState(&job, &jobModel)
		state.Jobs = append(state.Jobs, jobModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
