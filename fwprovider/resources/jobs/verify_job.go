/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package jobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &verifyJobResource{}
	_ resource.ResourceWithConfigure   = &verifyJobResource{}
	_ resource.ResourceWithImportState = &verifyJobResource{}
)

// NewVerifyJobResource is a helper function to simplify the provider implementation.
func NewVerifyJobResource() resource.Resource {
	return &verifyJobResource{}
}

// verifyJobResource is the resource implementation.
type verifyJobResource struct {
	client *pbs.Client
}

// verifyJobResourceModel maps the resource schema data.
type verifyJobResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Store          types.String `tfsdk:"store"`
	Schedule       types.String `tfsdk:"schedule"`
	IgnoreVerified types.Bool   `tfsdk:"ignore_verified"`
	OutdatedAfter  types.Int64  `tfsdk:"outdated_after"`
	Namespace      types.String `tfsdk:"namespace"`
	MaxDepth       types.Int64  `tfsdk:"max_depth"`
	Comment        types.String `tfsdk:"comment"`
	Disable        types.Bool   `tfsdk:"disable"`
	Digest         types.String `tfsdk:"digest"`
}

// Metadata returns the resource type name.
func (r *verifyJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verify_job"
}

// Schema defines the schema for the resource.
func (r *verifyJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PBS backup verification job for automated integrity checks.",
		MarkdownDescription: `Manages a PBS backup verification job.

Verification jobs check backup integrity by validating checksums and ensuring all data chunks 
are readable. This helps detect corruption or storage issues before you need to restore. 
The ` + "`ignore_verified`" + ` option skips recently verified backups, and ` + "`outdated_after`" + ` 
determines how many days until a backup is considered due for re-verification.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier for the verify job.",
				MarkdownDescription: "The unique identifier for the verification job.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"store": schema.StringAttribute{
				Description:         "The datastore name where backups will be verified.",
				MarkdownDescription: "The datastore name where backups will be verified.",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				Description:         "When to run the verify job (systemd calendar event format).",
				MarkdownDescription: "When to run the verification job. Uses systemd calendar event format (e.g., `weekly`, `Mon 03:00`).",
				Required:            true,
			},
			"ignore_verified": schema.BoolAttribute{
				Description:         "Skip backups that have been recently verified.",
				MarkdownDescription: "Skip backups that have been recently verified. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"outdated_after": schema.Int64Attribute{
				Description:         "Number of days after which a backup is considered outdated and needs re-verification.",
				MarkdownDescription: "Number of days after which a backup is considered outdated and needs re-verification. Only applies when `ignore_verified` is `true`.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"namespace": schema.StringAttribute{
				Description:         "Namespace to verify (optional, supports pattern matching).",
				MarkdownDescription: "Namespace to verify. Optional, supports pattern matching (e.g., `ns1`, `ns1/sub`).",
				Optional:            true,
			},
			"max_depth": schema.Int64Attribute{
				Description:         "Maximum recursion depth for namespaces.",
				MarkdownDescription: "Maximum recursion depth when traversing namespace hierarchy.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this verify job.",
				MarkdownDescription: "A comment describing this verification job.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this verify job without deleting it.",
				MarkdownDescription: "Disable this verify job without deleting it.",
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
func (r *verifyJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *verifyJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan verifyJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := buildVerifyJobFromPlan(&plan)

	if err := r.client.Jobs.CreateVerifyJob(ctx, job); err != nil {
		resp.Diagnostics.AddError(
			"Error creating verify job",
			fmt.Sprintf("Could not create verify job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	createdJob, err := r.client.Jobs.GetVerifyJob(ctx, job.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading verify job",
			fmt.Sprintf("Could not read verify job %s after creation: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	var state verifyJobResourceModel
	setVerifyStateFromAPI(createdJob, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *verifyJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state verifyJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.Jobs.GetVerifyJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading verify job",
			fmt.Sprintf("Could not read verify job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	setVerifyStateFromAPI(job, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *verifyJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan verifyJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state verifyJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (plan.Digest.IsNull() || plan.Digest.IsUnknown()) && !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		plan.Digest = state.Digest
	}

	job := buildVerifyJobFromPlan(&plan)
	job.Delete = computeVerifyDeletes(&plan, &state)

	if err := r.client.Jobs.UpdateVerifyJob(ctx, plan.ID.ValueString(), job); err != nil {
		resp.Diagnostics.AddError(
			"Error updating verify job",
			fmt.Sprintf("Could not update verify job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	updatedJob, err := r.client.Jobs.GetVerifyJob(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading verify job",
			fmt.Sprintf("Could not read verify job %s after update: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	setVerifyStateFromAPI(updatedJob, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *verifyJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state verifyJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	digest := ""
	if !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		digest = state.Digest.ValueString()
	}

	if err := r.client.Jobs.DeleteVerifyJob(ctx, state.ID.ValueString(), digest); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting verify job",
			fmt.Sprintf("Could not delete verify job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *verifyJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildVerifyJobFromPlan(plan *verifyJobResourceModel) *jobs.VerifyJob {
	job := &jobs.VerifyJob{
		ID:       plan.ID.ValueString(),
		Store:    plan.Store.ValueString(),
		Schedule: plan.Schedule.ValueString(),
	}

	if !plan.Namespace.IsNull() && !plan.Namespace.IsUnknown() {
		job.Namespace = plan.Namespace.ValueString()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		job.Comment = plan.Comment.ValueString()
	}
	if !plan.Digest.IsNull() && !plan.Digest.IsUnknown() {
		job.Digest = plan.Digest.ValueString()
	}

	job.IgnoreVerified = boolPointerFromAttr(plan.IgnoreVerified)
	job.Disable = boolPointerFromAttr(plan.Disable)
	job.OutdatedAfter = intPointerFromAttr(plan.OutdatedAfter)
	job.MaxDepth = intPointerFromAttr(plan.MaxDepth)

	return job
}

func computeVerifyDeletes(plan, state *verifyJobResourceModel) []string {
	if state == nil {
		return nil
	}

	var deletes []string

	if shouldDeleteStringAttr(plan.Namespace, state.Namespace) {
		deletes = append(deletes, "ns")
	}
	if shouldDeleteBoolAttr(plan.IgnoreVerified, state.IgnoreVerified) {
		deletes = append(deletes, "ignore-verified")
	}
	if shouldDeleteIntAttr(plan.OutdatedAfter, state.OutdatedAfter) {
		deletes = append(deletes, "outdated-after")
	}
	if shouldDeleteIntAttr(plan.MaxDepth, state.MaxDepth) {
		deletes = append(deletes, "max-depth")
	}
	if shouldDeleteStringAttr(plan.Comment, state.Comment) {
		deletes = append(deletes, "comment")
	}
	if shouldDeleteBoolAttr(plan.Disable, state.Disable) {
		deletes = append(deletes, "disable")
	}

	return deletes
}

func setVerifyStateFromAPI(job *jobs.VerifyJob, state *verifyJobResourceModel) {
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)
	state.Namespace = stringValueOrNull(job.Namespace)
	state.Comment = stringValueOrNull(job.Comment)
	state.Digest = stringValueOrNull(job.Digest)
	state.OutdatedAfter = int64ValueOrNull(job.OutdatedAfter)
	state.MaxDepth = int64ValueOrNull(job.MaxDepth)

	if job.IgnoreVerified != nil {
		state.IgnoreVerified = types.BoolValue(*job.IgnoreVerified)
	} else {
		state.IgnoreVerified = types.BoolValue(false)
	}
	if job.Disable != nil {
		state.Disable = types.BoolValue(*job.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}
}
