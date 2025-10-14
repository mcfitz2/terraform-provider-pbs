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
	// BackupType field removed in PBS 4.0
	BackupID    types.String `tfsdk:"backup_id"`
	Comment     types.String `tfsdk:"comment"`
	// Disable field removed in PBS 4.0
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
			// backup_type field removed in PBS 4.0
			"backup_id": schema.StringAttribute{
				Description:         "Specific backup ID filter.",
				MarkdownDescription: "Specific backup ID to prune. If set, only this backup ID will be affected.",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this prune job.",
				MarkdownDescription: "A comment describing this prune job.",
				Optional:            true,
			},
			// disable attribute removed in PBS 4.0
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
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := &jobs.PruneJob{
		ID:       plan.ID.ValueString(),
		Store:    plan.Store.ValueString(),
		Schedule: plan.Schedule.ValueString(),
	}

	if !plan.KeepLast.IsNull() {
		keepLast := int(plan.KeepLast.ValueInt64())
		job.KeepLast = &keepLast
	}
	if !plan.KeepHourly.IsNull() {
		keepHourly := int(plan.KeepHourly.ValueInt64())
		job.KeepHourly = &keepHourly
	}
	if !plan.KeepDaily.IsNull() {
		keepDaily := int(plan.KeepDaily.ValueInt64())
		job.KeepDaily = &keepDaily
	}
	if !plan.KeepWeekly.IsNull() {
		keepWeekly := int(plan.KeepWeekly.ValueInt64())
		job.KeepWeekly = &keepWeekly
	}
	if !plan.KeepMonthly.IsNull() {
		keepMonthly := int(plan.KeepMonthly.ValueInt64())
		job.KeepMonthly = &keepMonthly
	}
	if !plan.KeepYearly.IsNull() {
		keepYearly := int(plan.KeepYearly.ValueInt64())
		job.KeepYearly = &keepYearly
	}
	if !plan.MaxDepth.IsNull() {
		maxDepth := int(plan.MaxDepth.ValueInt64())
		job.MaxDepth = &maxDepth
	}
	if !plan.Namespace.IsNull() {
		job.NamespaceRE = plan.Namespace.ValueString()
	}
	// BackupType field removed in PBS 4.0
	if !plan.BackupID.IsNull() {
		job.BackupID = plan.BackupID.ValueString()
	}
	if !plan.Comment.IsNull() {
		job.Comment = plan.Comment.ValueString()
	}
	// Disable field removed in PBS 4.0

	err := r.client.Jobs.CreatePruneJob(ctx, job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating prune job",
			fmt.Sprintf("Could not create prune job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *pruneJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pruneJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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

	// Update state
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)

	if job.KeepLast != nil {
		state.KeepLast = types.Int64Value(int64(*job.KeepLast))
	} else {
		state.KeepLast = types.Int64Null()
	}
	if job.KeepHourly != nil {
		state.KeepHourly = types.Int64Value(int64(*job.KeepHourly))
	} else {
		state.KeepHourly = types.Int64Null()
	}
	if job.KeepDaily != nil {
		state.KeepDaily = types.Int64Value(int64(*job.KeepDaily))
	} else {
		state.KeepDaily = types.Int64Null()
	}
	if job.KeepWeekly != nil {
		state.KeepWeekly = types.Int64Value(int64(*job.KeepWeekly))
	} else {
		state.KeepWeekly = types.Int64Null()
	}
	if job.KeepMonthly != nil {
		state.KeepMonthly = types.Int64Value(int64(*job.KeepMonthly))
	} else {
		state.KeepMonthly = types.Int64Null()
	}
	if job.KeepYearly != nil {
		state.KeepYearly = types.Int64Value(int64(*job.KeepYearly))
	} else {
		state.KeepYearly = types.Int64Null()
	}
	if job.MaxDepth != nil {
		state.MaxDepth = types.Int64Value(int64(*job.MaxDepth))
	} else {
		state.MaxDepth = types.Int64Null()
	}

	if job.NamespaceRE != "" {
		state.Namespace = types.StringValue(job.NamespaceRE)
	} else {
		state.Namespace = types.StringNull()
	}
	// BackupType field removed in PBS 4.0
	if job.BackupID != "" {
		state.BackupID = types.StringValue(job.BackupID)
	} else {
		state.BackupID = types.StringNull()
	}
	if job.Comment != "" {
		state.Comment = types.StringValue(job.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	// Disable field removed in PBS 4.0

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *pruneJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pruneJobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := &jobs.PruneJob{
		ID:       plan.ID.ValueString(),
		Store:    plan.Store.ValueString(),
		Schedule: plan.Schedule.ValueString(),
	}

	if !plan.KeepLast.IsNull() {
		keepLast := int(plan.KeepLast.ValueInt64())
		job.KeepLast = &keepLast
	}
	if !plan.KeepHourly.IsNull() {
		keepHourly := int(plan.KeepHourly.ValueInt64())
		job.KeepHourly = &keepHourly
	}
	if !plan.KeepDaily.IsNull() {
		keepDaily := int(plan.KeepDaily.ValueInt64())
		job.KeepDaily = &keepDaily
	}
	if !plan.KeepWeekly.IsNull() {
		keepWeekly := int(plan.KeepWeekly.ValueInt64())
		job.KeepWeekly = &keepWeekly
	}
	if !plan.KeepMonthly.IsNull() {
		keepMonthly := int(plan.KeepMonthly.ValueInt64())
		job.KeepMonthly = &keepMonthly
	}
	if !plan.KeepYearly.IsNull() {
		keepYearly := int(plan.KeepYearly.ValueInt64())
		job.KeepYearly = &keepYearly
	}
	if !plan.MaxDepth.IsNull() {
		maxDepth := int(plan.MaxDepth.ValueInt64())
		job.MaxDepth = &maxDepth
	}
	if !plan.Namespace.IsNull() {
		job.NamespaceRE = plan.Namespace.ValueString()
	}
	// BackupType field removed in PBS 4.0
	if !plan.BackupID.IsNull() {
		job.BackupID = plan.BackupID.ValueString()
	}
	if !plan.Comment.IsNull() {
		job.Comment = plan.Comment.ValueString()
	}
	// Disable field removed in PBS 4.0

	err := r.client.Jobs.UpdatePruneJob(ctx, plan.ID.ValueString(), job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating prune job",
			fmt.Sprintf("Could not update prune job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *pruneJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pruneJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Jobs.DeletePruneJob(ctx, state.ID.ValueString())
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
