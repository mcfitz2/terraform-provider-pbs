/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package jobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &gcJobResource{}
	_ resource.ResourceWithConfigure   = &gcJobResource{}
	_ resource.ResourceWithImportState = &gcJobResource{}
)

// NewGCJobResource is a helper function to simplify the provider implementation.
func NewGCJobResource() resource.Resource {
	return &gcJobResource{}
}

// gcJobResource is the resource implementation.
type gcJobResource struct {
	client *pbs.Client
}

// gcJobResourceModel maps the resource schema data.
type gcJobResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Store    types.String `tfsdk:"store"`
	Schedule types.String `tfsdk:"schedule"`
	Comment  types.String `tfsdk:"comment"`
	Disable  types.Bool   `tfsdk:"disable"`
}

// Metadata returns the resource type name.
func (r *gcJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gc_job"
}

// Schema defines the schema for the resource.
func (r *gcJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PBS garbage collection job for automated storage cleanup.",
		MarkdownDescription: `Manages a PBS garbage collection (GC) job.

Garbage collection jobs remove unreferenced data chunks from the datastore, freeing up storage 
space after backups are pruned. GC should be run after prune jobs to actually reclaim the 
disk space. It's recommended to schedule GC jobs to run during off-peak hours as they can 
be I/O intensive.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier for the GC job.",
				MarkdownDescription: "The unique identifier for the garbage collection job.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"store": schema.StringAttribute{
				Description:         "The datastore name where garbage collection will be performed.",
				MarkdownDescription: "The datastore name where garbage collection will be performed.",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				Description:         "When to run the GC job (systemd calendar event format).",
				MarkdownDescription: "When to run the GC job. Uses systemd calendar event format (e.g., `daily`, `weekly`, `Sun 02:00`).",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this GC job.",
				MarkdownDescription: "A comment describing this garbage collection job.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this GC job.",
				MarkdownDescription: "Disable this garbage collection job. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *gcJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pbs.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *pbs.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *gcJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gcJobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := &jobs.GCJob{
		ID:       plan.ID.ValueString(),
		Store:    plan.Store.ValueString(),
		Schedule: plan.Schedule.ValueString(),
	}

	if !plan.Comment.IsNull() {
		job.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		job.Disable = &disable
	}

	err := r.client.Jobs.CreateGCJob(ctx, job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating GC job",
			fmt.Sprintf("Could not create GC job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *gcJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gcJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.Jobs.GetGCJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading GC job",
			fmt.Sprintf("Could not read GC job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update state
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)

	if job.Comment != "" {
		state.Comment = types.StringValue(job.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	if job.Disable != nil {
		state.Disable = types.BoolValue(*job.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *gcJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gcJobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := &jobs.GCJob{
		ID:       plan.ID.ValueString(),
		Store:    plan.Store.ValueString(),
		Schedule: plan.Schedule.ValueString(),
	}

	if !plan.Comment.IsNull() {
		job.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		job.Disable = &disable
	}

	err := r.client.Jobs.UpdateGCJob(ctx, plan.ID.ValueString(), job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating GC job",
			fmt.Sprintf("Could not update GC job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *gcJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gcJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Jobs.DeleteGCJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting GC job",
			fmt.Sprintf("Could not delete GC job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *gcJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
