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
	_ datasource.DataSource              = &verifyJobsDataSource{}
	_ datasource.DataSourceWithConfigure = &verifyJobsDataSource{}
)

// NewVerifyJobsDataSource is a helper function to simplify the provider implementation.
func NewVerifyJobsDataSource() datasource.DataSource {
	return &verifyJobsDataSource{}
}

// verifyJobsDataSource is the data source implementation.
type verifyJobsDataSource struct {
	client *pbs.Client
}

// verifyJobsDataSourceModel maps the data source schema data.
type verifyJobsDataSourceModel struct {
	Store types.String               `tfsdk:"store"`
	Jobs  []verifyJobDataSourceModel `tfsdk:"jobs"`
}

// Metadata returns the data source type name.
func (d *verifyJobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verify_jobs"
}

// Schema defines the schema for the data source.
func (d *verifyJobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all PBS verify jobs, optionally filtered by datastore.",
		MarkdownDescription: "Lists all PBS verify jobs, optionally filtered by datastore.",
		Attributes: map[string]schema.Attribute{
			"store": schema.StringAttribute{
				Description:         "Filter verify jobs by datastore name (optional).",
				MarkdownDescription: "Filter verify jobs by datastore name (optional).",
				Optional:            true,
			},
			"jobs": schema.ListNestedAttribute{
				Description:         "List of verify jobs.",
				MarkdownDescription: "List of verify jobs.",
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
						"ignore_verified": schema.BoolAttribute{
							Computed: true,
						},
						"outdated_after": schema.Int64Attribute{
							Computed: true,
						},
						"namespace": schema.StringAttribute{
							Computed: true,
						},
						"max_depth": schema.Int64Attribute{
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
func (d *verifyJobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *verifyJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state verifyJobsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all verify jobs from API
	jobs, err := d.client.Jobs.ListVerifyJobs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Verify Jobs",
			fmt.Sprintf("Could not list verify jobs: %s", err.Error()),
		)
		return
	}

	// Filter by store if specified
	storeFilter := ""
	if !state.Store.IsNull() && !state.Store.IsUnknown() {
		storeFilter = state.Store.ValueString()
	}

	// Map API response to state
	state.Jobs = make([]verifyJobDataSourceModel, 0)
	for _, job := range jobs {
		if storeFilter != "" && job.Store != storeFilter {
			continue
		}
		var jobModel verifyJobDataSourceModel
		verifyJobToState(&job, &jobModel)
		state.Jobs = append(state.Jobs, jobModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
