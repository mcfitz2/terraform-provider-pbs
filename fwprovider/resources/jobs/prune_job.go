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

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &pruneJobResource{}
	_ resource.ResourceWithConfigure   = &pruneJobResource{}
	_ resource.ResourceWithImportState = &pruneJobResource{}
)

// NewPruneJobResource is a helper function to simplify the provider implementation.
func NewPruneJobResource() resource.Resource {
	return &pruneJobResource{}
}

// pruneJobResource is the resource implementation.
type pruneJobResource struct {
	client *pbs.Client
}

// pruneJobResourceModel maps the resource schema data.
type pruneJobResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Store       types.String `tfsdk:"store"`
	Schedule    types.String `tfsdk:"schedule"`
	KeepLast    types.Int64  `tfsdk:"keep_last"`
	KeepHourly  types.Int64  `tfsdk:"keep_hourly"`
	KeepDaily   types.Int64  `tfsdk:"keep_daily"`
	KeepWeekly  types.Int64  `tfsdk:"keep_weekly"`
	KeepMonthly types.Int64  `tfsdk:"keep_monthly"`
	KeepYearly  types.Int64  `tfsdk:"keep_yearly"`
	MaxDepth    types.Int64  `tfsdk:"max_depth"`
	Namespace   types.String `tfsdk:"namespace"`
	Comment     types.String `tfsdk:"comment"`
	Disable     types.Bool   `tfsdk:"disable"`
	Digest      types.String `tfsdk:"digest"`
}

// Metadata returns the resource type name.
func (r *pruneJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prune_job"
}

// Schema defines the schema for the resource.
func (r *pruneJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PBS prune job for automated backup retention management.",
		MarkdownDescription: `Manages a PBS prune job.

Prune jobs automatically remove old backup snapshots based on retention policies. This helps 
maintain storage efficiency while ensuring important backups are retained according to your 
requirements. Configure retention using keep-last, keep-hourly, keep-daily, keep-weekly, 
keep-monthly, and keep-yearly parameters.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier for the prune job.",
				MarkdownDescription: "The unique identifier for the prune job.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"store": schema.StringAttribute{
				Description:         "The datastore name where pruning will be performed.",
				MarkdownDescription: "The datastore name where pruning will be performed.",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				Description:         "When to run the prune job (systemd calendar event format).",
				MarkdownDescription: "When to run the prune job. Uses systemd calendar event format (e.g., `daily`, `weekly`, `Mon..Fri *-*-* 02:00:00`).",
				Required:            true,
			},
			"keep_last": schema.Int64Attribute{
				Description:         "Keep the last N backup snapshots.",
				MarkdownDescription: "Keep the last N backup snapshots, regardless of time.",
				Optional:            true,
			},
			"keep_hourly": schema.Int64Attribute{
				Description:         "Keep hourly backups for the last N hours.",
				MarkdownDescription: "Keep hourly backups for the last N hours.",
				Optional:            true,
			},
			"keep_daily": schema.Int64Attribute{
				Description:         "Keep daily backups for the last N days.",
				MarkdownDescription: "Keep daily backups for the last N days.",
				Optional:            true,
			},
			"keep_weekly": schema.Int64Attribute{
				Description:         "Keep weekly backups for the last N weeks.",
				MarkdownDescription: "Keep weekly backups for the last N weeks.",
				Optional:            true,
			},
			"keep_monthly": schema.Int64Attribute{
				Description:         "Keep monthly backups for the last N months.",
				MarkdownDescription: "Keep monthly backups for the last N months.",
				Optional:            true,
			},
			"keep_yearly": schema.Int64Attribute{
				Description:         "Keep yearly backups for the last N years.",
				MarkdownDescription: "Keep yearly backups for the last N years.",
				Optional:            true,
			},
			"max_depth": schema.Int64Attribute{
				Description:         "Maximum depth for namespace traversal.",
				MarkdownDescription: "Maximum depth for namespace traversal when pruning.",
				Optional:            true,
			},
			"namespace": schema.StringAttribute{
				Description:         "Namespace filter (regex).",
				MarkdownDescription: "Namespace filter as a regular expression. Only backups in matching namespaces will be pruned.",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this prune job.",
				MarkdownDescription: "A comment describing this prune job.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this prune job without deleting it.",
				MarkdownDescription: "Disable this prune job without deleting it.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"digest": schema.StringAttribute{
				Description:         "Opaque digest returned by PBS for optimistic locking.",
				MarkdownDescription: "Opaque digest returned by PBS for optimistic locking.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *pruneJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.Resource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = cfg.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *pruneJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pruneJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := buildPruneJobFromPlan(&plan)

	if err := r.client.Jobs.CreatePruneJob(ctx, job); err != nil {
		resp.Diagnostics.AddError(
			"Error creating prune job",
			fmt.Sprintf("Could not create prune job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	createdJob, err := r.client.Jobs.GetPruneJob(ctx, job.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading prune job",
			fmt.Sprintf("Could not read prune job %s after creation: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	var state pruneJobResourceModel
	setPruneStateFromAPI(createdJob, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *pruneJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pruneJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.Jobs.GetPruneJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading prune job",
			fmt.Sprintf("Could not read prune job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	setPruneStateFromAPI(job, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *pruneJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pruneJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state pruneJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (plan.Digest.IsNull() || plan.Digest.IsUnknown()) && !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		plan.Digest = state.Digest
	}

	job := buildPruneJobFromPlan(&plan)
	job.Delete = computePruneDeletes(&plan, &state)

	if err := r.client.Jobs.UpdatePruneJob(ctx, plan.ID.ValueString(), job); err != nil {
		resp.Diagnostics.AddError(
			"Error updating prune job",
			fmt.Sprintf("Could not update prune job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	updatedJob, err := r.client.Jobs.GetPruneJob(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading prune job",
			fmt.Sprintf("Could not read prune job %s after update: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	setPruneStateFromAPI(updatedJob, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *pruneJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pruneJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	digest := ""
	if !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		digest = state.Digest.ValueString()
	}

	err := r.client.Jobs.DeletePruneJob(ctx, state.ID.ValueString(), digest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting prune job",
			fmt.Sprintf("Could not delete prune job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *pruneJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildPruneJobFromPlan(plan *pruneJobResourceModel) *jobs.PruneJob {
	job := &jobs.PruneJob{
		ID:       plan.ID.ValueString(),
		Store:    plan.Store.ValueString(),
		Schedule: plan.Schedule.ValueString(),
	}

	job.KeepLast = intPointerFromAttr(plan.KeepLast)
	job.KeepHourly = intPointerFromAttr(plan.KeepHourly)
	job.KeepDaily = intPointerFromAttr(plan.KeepDaily)
	job.KeepWeekly = intPointerFromAttr(plan.KeepWeekly)
	job.KeepMonthly = intPointerFromAttr(plan.KeepMonthly)
	job.KeepYearly = intPointerFromAttr(plan.KeepYearly)
	job.MaxDepth = intPointerFromAttr(plan.MaxDepth)

	if !plan.Namespace.IsNull() && !plan.Namespace.IsUnknown() {
		job.Namespace = plan.Namespace.ValueString()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		job.Comment = plan.Comment.ValueString()
	}
	if ptr := boolPointerFromAttr(plan.Disable); ptr != nil {
		job.Disable = ptr
	}
	if !plan.Digest.IsNull() && !plan.Digest.IsUnknown() {
		job.Digest = plan.Digest.ValueString()
	}

	return job
}

func computePruneDeletes(plan, state *pruneJobResourceModel) []string {
	if state == nil {
		return nil
	}

	var deletes []string

	if shouldDeleteIntAttr(plan.KeepLast, state.KeepLast) {
		deletes = append(deletes, "keep-last")
	}
	if shouldDeleteIntAttr(plan.KeepHourly, state.KeepHourly) {
		deletes = append(deletes, "keep-hourly")
	}
	if shouldDeleteIntAttr(plan.KeepDaily, state.KeepDaily) {
		deletes = append(deletes, "keep-daily")
	}
	if shouldDeleteIntAttr(plan.KeepWeekly, state.KeepWeekly) {
		deletes = append(deletes, "keep-weekly")
	}
	if shouldDeleteIntAttr(plan.KeepMonthly, state.KeepMonthly) {
		deletes = append(deletes, "keep-monthly")
	}
	if shouldDeleteIntAttr(plan.KeepYearly, state.KeepYearly) {
		deletes = append(deletes, "keep-yearly")
	}
	if shouldDeleteIntAttr(plan.MaxDepth, state.MaxDepth) {
		deletes = append(deletes, "max-depth")
	}
	if shouldDeleteStringAttr(plan.Namespace, state.Namespace) {
		deletes = append(deletes, "ns")
	}
	if shouldDeleteStringAttr(plan.Comment, state.Comment) {
		deletes = append(deletes, "comment")
	}

	return deletes
}

func setPruneStateFromAPI(job *jobs.PruneJob, state *pruneJobResourceModel) {
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)
	state.KeepLast = int64ValueOrNull(job.KeepLast)
	state.KeepHourly = int64ValueOrNull(job.KeepHourly)
	state.KeepDaily = int64ValueOrNull(job.KeepDaily)
	state.KeepWeekly = int64ValueOrNull(job.KeepWeekly)
	state.KeepMonthly = int64ValueOrNull(job.KeepMonthly)
	state.KeepYearly = int64ValueOrNull(job.KeepYearly)
	state.MaxDepth = int64ValueOrNull(job.MaxDepth)
	state.Namespace = stringValueOrNull(job.Namespace)
	state.Comment = stringValueOrNull(job.Comment)
	if job.Disable != nil {
		state.Disable = types.BoolValue(*job.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}
	state.Digest = stringValueOrNull(job.Digest)
}
