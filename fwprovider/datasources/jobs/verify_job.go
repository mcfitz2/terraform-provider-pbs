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
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

var (
	_ datasource.DataSource              = &verifyJobDataSource{}
	_ datasource.DataSourceWithConfigure = &verifyJobDataSource{}
)

// NewVerifyJobDataSource is a helper function to simplify the provider implementation.
func NewVerifyJobDataSource() datasource.DataSource {
	return &verifyJobDataSource{}
}

// verifyJobDataSource is the data source implementation.
type verifyJobDataSource struct {
	client *pbs.Client
}

// verifyJobDataSourceModel maps the data source schema data.
type verifyJobDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Store              types.String `tfsdk:"store"`
	Schedule           types.String `tfsdk:"schedule"`
	IgnoreVerified     types.Bool   `tfsdk:"ignore_verified"`
	OutdatedAfter      types.Int64  `tfsdk:"outdated_after"`
	Namespace          types.String `tfsdk:"namespace"`
	MaxDepth           types.Int64  `tfsdk:"max_depth"`
	Comment            types.String `tfsdk:"comment"`
	Digest             types.String `tfsdk:"digest"`
}

// Metadata returns the data source type name.
func (d *verifyJobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verify_job"
}

// Schema defines the schema for the data source.
func (d *verifyJobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads information about an existing PBS verify job.",
		MarkdownDescription: "Reads information about an existing PBS verify job configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier for the verify job.",
				MarkdownDescription: "The unique identifier for the verify job.",
				Required:            true,
			},
			"store": schema.StringAttribute{
				Description:         "The datastore name where verification is performed.",
				MarkdownDescription: "The datastore name where verification is performed.",
				Computed:            true,
			},
			"schedule": schema.StringAttribute{
				Description:         "When the verify job runs.",
				MarkdownDescription: "When the verify job runs.",
				Computed:            true,
			},
			"ignore_verified": schema.BoolAttribute{
				Description:         "Skip snapshots verified after outdated_after.",
				MarkdownDescription: "Skip snapshots verified after outdated_after.",
				Computed:            true,
			},
			"outdated_after": schema.Int64Attribute{
				Description:         "Days after which to re-verify snapshots.",
				MarkdownDescription: "Days after which to re-verify snapshots.",
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				Description:         "Namespace filter (regex).",
				MarkdownDescription: "Namespace filter (regex).",
				Computed:            true,
			},
			"max_depth": schema.Int64Attribute{
				Description:         "Maximum depth for namespace traversal.",
				MarkdownDescription: "Maximum depth for namespace traversal.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this verify job.",
				MarkdownDescription: "A comment describing this verify job.",
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
func (d *verifyJobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *verifyJobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state verifyJobDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get verify job from API
	job, err := d.client.Jobs.GetVerifyJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Verify Job",
			fmt.Sprintf("Could not read verify job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Map API response to state
	verifyJobToState(job, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// verifyJobToState converts a verify job struct to Terraform state
func verifyJobToState(job *jobs.VerifyJob, state *verifyJobDataSourceModel) {
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)
	state.IgnoreVerified = boolPtrToValue(job.IgnoreVerified)
	state.OutdatedAfter = intPtrToValue(job.OutdatedAfter)
	state.Namespace = stringToValue(job.Namespace)
	state.MaxDepth = intPtrToValue(job.MaxDepth)
	state.Comment = stringToValue(job.Comment)
	state.Digest = types.StringValue(job.Digest)
}
