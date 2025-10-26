/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package remotes

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/remotes"
)

var (
	_ resource.Resource                = &remoteResource{}
	_ resource.ResourceWithConfigure   = &remoteResource{}
	_ resource.ResourceWithImportState = &remoteResource{}
)

var (
	authIDRegex      = regexp.MustCompile(`^(?:(?:[^\s:/[:cntrl:]]+)@(?:[A-Za-z0-9_][A-Za-z0-9._\-]*)|(?:[^\s:/[:cntrl:]]+)@(?:[A-Za-z0-9_][A-Za-z0-9._\-]*)!(?:[A-Za-z0-9_][A-Za-z0-9._\-]*))$`)
	fingerprintRegex = regexp.MustCompile(`^(?:[0-9a-fA-F][0-9a-fA-F])(?::[0-9a-fA-F][0-9a-fA-F]){31}$`)
)

// NewRemoteResource is a helper function to simplify the provider implementation.
func NewRemoteResource() resource.Resource {
	return &remoteResource{}
}

// remoteResource is the resource implementation.
type remoteResource struct {
	client *pbs.Client
}

// remoteResourceModel maps the resource schema data.
type remoteResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Host        types.String `tfsdk:"host"`
	Port        types.Int64  `tfsdk:"port"`
	AuthID      types.String `tfsdk:"auth_id"`
	Password    types.String `tfsdk:"password"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Comment     types.String `tfsdk:"comment"`
	Digest      types.String `tfsdk:"digest"`
}

// Metadata returns the resource type name.
func (r *remoteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote"
}

// Schema defines the schema for the resource.
func (r *remoteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PBS remote server configuration for sync jobs.",
		MarkdownDescription: `Manages a PBS remote server configuration.

Remote configurations allow PBS to connect to other PBS servers for pulling or pushing backups. 
Remotes are referenced by sync jobs to replicate data between PBS instances.

**Note:** The password is write-only and will not be stored in Terraform state after the initial creation. 
Updates to the password will always be sent to the API but cannot be verified by reading back the configuration.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique identifier for the remote (3-32 characters).",
				MarkdownDescription: "The unique identifier for the remote (3-32 characters).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._\-]*$`),
						"must start with alphanumeric or underscore, and contain only alphanumeric, underscore, dot, or hyphen",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				Description: "The hostname or IP address of the remote PBS server.",
				MarkdownDescription: "The hostname or IP address of the remote PBS server.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				Description:         "The port number for the remote PBS server (default: 8007).",
				MarkdownDescription: "The port number for the remote PBS server (default: 8007).",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"auth_id": schema.StringAttribute{
				Description: "Authentication ID for the remote server (e.g., 'user@pam' or 'user@pbs!token').",
				MarkdownDescription: "Authentication ID for the remote server (e.g., `user@pam` or `user@pbs!token`).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
					stringvalidator.RegexMatches(
						authIDRegex,
						"must be in format 'user@realm' or 'user@realm!token'",
					),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password or authentication token for the remote server.",
				MarkdownDescription: "Password or authentication token for the remote server. " +
					"This value is write-only and will not be stored in state after creation.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			"fingerprint": schema.StringAttribute{
				Description: "X509 certificate fingerprint (SHA256) for TLS verification.",
				MarkdownDescription: "X509 certificate fingerprint (SHA256) for TLS verification. " +
					"Format: `AA:BB:CC:...` (32 pairs of hex digits separated by colons).",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						fingerprintRegex,
						"must be 32 pairs of hexadecimal digits separated by colons",
					),
				},
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this remote.",
				MarkdownDescription: "A comment describing this remote.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(128),
				},
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
func (r *remoteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *remoteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan remoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remote := &remotes.Remote{
		Name:     plan.Name.ValueString(),
		Host:     plan.Host.ValueString(),
		AuthID:   plan.AuthID.ValueString(),
		Password: plan.Password.ValueString(),
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := int(plan.Port.ValueInt64())
		remote.Port = &port
	}
	if !plan.Fingerprint.IsNull() && !plan.Fingerprint.IsUnknown() {
		remote.Fingerprint = plan.Fingerprint.ValueString()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		remote.Comment = plan.Comment.ValueString()
	}

	if err := r.client.Remotes.CreateRemote(ctx, remote); err != nil {
		resp.Diagnostics.AddError(
			"Error creating remote",
			fmt.Sprintf("Could not create remote %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back to get digest and verify creation
	createdRemote, err := r.client.Remotes.GetRemote(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading remote",
			fmt.Sprintf("Could not read remote %s after creation: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	var state remoteResourceModel
	setRemoteState(createdRemote, &state, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *remoteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state remoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remote, err := r.client.Remotes.GetRemote(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading remote",
			fmt.Sprintf("Could not read remote %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	setRemoteState(remote, &state, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *remoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan remoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state remoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Copy digest from state if not in plan
	if (plan.Digest.IsNull() || plan.Digest.IsUnknown()) && !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		plan.Digest = state.Digest
	}

	remote := &remotes.Remote{
		Name:     plan.Name.ValueString(),
		Host:     plan.Host.ValueString(),
		AuthID:   plan.AuthID.ValueString(),
		Password: plan.Password.ValueString(),
		Digest:   plan.Digest.ValueString(),
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := int(plan.Port.ValueInt64())
		remote.Port = &port
	}
	if !plan.Fingerprint.IsNull() && !plan.Fingerprint.IsUnknown() {
		remote.Fingerprint = plan.Fingerprint.ValueString()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		remote.Comment = plan.Comment.ValueString()
	}

	// Compute delete array for fields that were cleared
	remote.Delete = computeRemoteDeletes(&plan, &state)

	if err := r.client.Remotes.UpdateRemote(ctx, plan.Name.ValueString(), remote); err != nil {
		resp.Diagnostics.AddError(
			"Error updating remote",
			fmt.Sprintf("Could not update remote %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back to get updated digest
	updatedRemote, err := r.client.Remotes.GetRemote(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading remote",
			fmt.Sprintf("Could not read remote %s after update: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	setRemoteState(updatedRemote, &state, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *remoteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state remoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	digest := ""
	if !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		digest = state.Digest.ValueString()
	}

	if err := r.client.Remotes.DeleteRemote(ctx, state.Name.ValueString(), digest); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting remote",
			fmt.Sprintf("Could not delete remote %s: %s", state.Name.ValueString(), err.Error()),
		)
	}
}

// ImportState imports the resource into Terraform state.
func (r *remoteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by name
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	
	// Add a warning about password
	resp.Diagnostics.AddWarning(
		"Password not imported",
		"The remote password/token is not included in the import. You must set the password in your configuration and run 'terraform apply' to update it.",
	)
}

// setRemoteState maps API response to Terraform state
func setRemoteState(remote *remotes.Remote, state *remoteResourceModel, plan *remoteResourceModel) {
	state.Name = types.StringValue(remote.Name)
	state.Host = types.StringValue(remote.Host)
	state.AuthID = types.StringValue(remote.AuthID)
	state.Digest = types.StringValue(remote.Digest)

	// Port handling
	if remote.Port != nil {
		state.Port = types.Int64Value(int64(*remote.Port))
	} else {
		state.Port = types.Int64Null()
	}

	// Fingerprint handling
	if remote.Fingerprint != "" {
		state.Fingerprint = types.StringValue(remote.Fingerprint)
	} else {
		state.Fingerprint = types.StringNull()
	}

	// Comment handling
	if remote.Comment != "" {
		state.Comment = types.StringValue(remote.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	// Password is write-only - preserve from plan, don't read from API
	if plan != nil && !plan.Password.IsNull() {
		state.Password = plan.Password
	} else {
		// On import or if not in plan, set to unknown to trigger update
		state.Password = types.StringValue("")
	}
}

// computeRemoteDeletes determines which optional fields should be deleted
func computeRemoteDeletes(plan, state *remoteResourceModel) []string {
	var deletes []string

	// Port was set, now is null
	if (!state.Port.IsNull() && !state.Port.IsUnknown()) &&
		(plan.Port.IsNull() || plan.Port.IsUnknown()) {
		deletes = append(deletes, "port")
	}

	// Fingerprint was set, now is null
	if (!state.Fingerprint.IsNull() && !state.Fingerprint.IsUnknown()) &&
		(plan.Fingerprint.IsNull() || plan.Fingerprint.IsUnknown()) {
		deletes = append(deletes, "fingerprint")
	}

	// Comment was set, now is null
	if (!state.Comment.IsNull() && !state.Comment.IsUnknown()) &&
		(plan.Comment.IsNull() || plan.Comment.IsUnknown()) {
		deletes = append(deletes, "comment")
	}

	return deletes
}
