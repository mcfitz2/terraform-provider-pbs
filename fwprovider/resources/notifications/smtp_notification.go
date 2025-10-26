/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &smtpNotificationResource{}
	_ resource.ResourceWithConfigure   = &smtpNotificationResource{}
	_ resource.ResourceWithImportState = &smtpNotificationResource{}
)

// NewSMTPNotificationResource is a helper function to simplify the provider implementation.
func NewSMTPNotificationResource() resource.Resource {
	return &smtpNotificationResource{}
}

// smtpNotificationResource is the resource implementation.
type smtpNotificationResource struct {
	client *pbs.Client
}

// smtpNotificationResourceModel maps the resource schema data.
type smtpNotificationResourceModel struct {
	Name       types.String `tfsdk:"name"`
	Server     types.String `tfsdk:"server"`
	Port       types.Int64  `tfsdk:"port"`
	Mode       types.String `tfsdk:"mode"`
	From       types.String `tfsdk:"from_address"`
	Mailto     types.List   `tfsdk:"mailto"`
	MailtoUser types.List   `tfsdk:"mailto_user"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Author     types.String `tfsdk:"author"`
	Comment    types.String `tfsdk:"comment"`
	Disable    types.Bool   `tfsdk:"disable"`
	Origin     types.String `tfsdk:"origin"`
}

// Metadata returns the resource type name.
func (r *smtpNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_notification"
}

// Schema defines the schema for the resource.
func (r *smtpNotificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an SMTP notification target for PBS alerts and notifications.",
		MarkdownDescription: `Manages an SMTP notification target.

Configure an SMTP server to receive notifications from PBS about backup jobs, 
verification tasks, and system events.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique name identifier for the SMTP target.",
				MarkdownDescription: "The unique name identifier for the SMTP notification target.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server": schema.StringAttribute{
				Description:         "SMTP server hostname or IP address.",
				MarkdownDescription: "SMTP server hostname or IP address.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				Description:         "SMTP server port.",
				MarkdownDescription: "SMTP server port. Common values: `25` (unencrypted), `465` (TLS), `587` (STARTTLS). Defaults to `25`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(25),
			},
			"mode": schema.StringAttribute{
				Description:         "Connection mode for SMTP.",
				MarkdownDescription: "Connection mode for SMTP. Valid values: `insecure` (no encryption), `starttls` (upgrade to TLS), `tls` (direct TLS). Defaults to `insecure`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("insecure"),
				Validators: []validator.String{
					stringvalidator.OneOf("insecure", "starttls", "tls"),
				},
			},
			"from_address": schema.StringAttribute{
				Description:         "Sender email address.",
				MarkdownDescription: "Sender email address. This will appear as the 'From' address in notification emails.",
				Required:            true,
			},
			"mailto": schema.ListAttribute{
				Description:         "Recipient email address(es).",
				MarkdownDescription: "Recipient email address(es). Specify as a list of email strings.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"mailto_user": schema.ListAttribute{
				Description:         "User(s) from PBS user database to receive notifications.",
				MarkdownDescription: "User(s) from PBS user database to receive notifications. Specify as PBS user IDs.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description:         "SMTP authentication username.",
				MarkdownDescription: "SMTP authentication username. Required if the SMTP server requires authentication.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Description:         "SMTP authentication password.",
				MarkdownDescription: "SMTP authentication password. Required if the SMTP server requires authentication.",
				Optional:            true,
				Sensitive:           true,
			},
			"author": schema.StringAttribute{
				Description:         "Author name for notification emails.",
				MarkdownDescription: "Author name that will appear in the email headers.",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this notification target.",
				MarkdownDescription: "A comment describing this notification target.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this notification target.",
				MarkdownDescription: "Disable this notification target. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"origin": schema.StringAttribute{
				Description:         "Origin of this configuration as reported by PBS (e.g., config file or built-in).",
				MarkdownDescription: "Origin of this configuration as reported by PBS (e.g., `user`, `builtin`).",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *smtpNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *smtpNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan smtpNotificationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create SMTP target via API
	target := &notifications.SMTPTarget{
		Name:   plan.Name.ValueString(),
		Server: plan.Server.ValueString(),
		From:   plan.From.ValueString(),
	}

	// Set optional fields
	if !plan.Port.IsNull() {
		port := int(plan.Port.ValueInt64())
		target.Port = &port
	}
	if !plan.Mode.IsNull() && !plan.Mode.IsUnknown() {
		target.Mode = plan.Mode.ValueString()
	}
	if plan.Mailto.IsNull() {
		target.To = []string{}
	} else if !plan.Mailto.IsUnknown() {
		var mailto []string
		diags := plan.Mailto.ElementsAs(ctx, &mailto, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.To = mailto
	}
	if plan.MailtoUser.IsNull() {
		target.MailtoUser = []string{}
	} else if !plan.MailtoUser.IsUnknown() {
		var mailtoUsers []string
		diags := plan.MailtoUser.ElementsAs(ctx, &mailtoUsers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.MailtoUser = mailtoUsers
	}
	if !plan.Username.IsNull() {
		target.Username = plan.Username.ValueString()
	}
	if !plan.Password.IsNull() {
		target.Password = plan.Password.ValueString()
	}
	if !plan.Author.IsNull() {
		target.Author = plan.Author.ValueString()
	}
	if !plan.Comment.IsNull() {
		target.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		target.Disable = &disable
	}

	err := r.client.Notifications.CreateSMTPTarget(ctx, target)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SMTP notification target",
			fmt.Sprintf("Could not create SMTP notification target %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back from API to get computed values (like mode default)
	created, err := r.client.Notifications.GetSMTPTarget(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created SMTP notification target",
			fmt.Sprintf("Created SMTP notification target %s but could not read it back: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Update plan with values from API
	plan.Server = types.StringValue(created.Server)
	plan.From = types.StringValue(created.From)
	if created.Port != nil {
		plan.Port = types.Int64Value(int64(*created.Port))
	}
	if created.Mode != "" {
		plan.Mode = types.StringValue(created.Mode)
	} else if plan.Mode.IsUnknown() || plan.Mode.IsNull() {
		// If API doesn't return mode and plan has no value, use default
		plan.Mode = types.StringValue("insecure")
	}
	if created.To != nil {
		mailtoList, diags := types.ListValueFrom(ctx, types.StringType, created.To)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Mailto = mailtoList
	} else {
		plan.Mailto = types.ListNull(types.StringType)
	}
	if created.MailtoUser != nil {
		mailtoUserList, diags := types.ListValueFrom(ctx, types.StringType, created.MailtoUser)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.MailtoUser = mailtoUserList
	} else {
		plan.MailtoUser = types.ListNull(types.StringType)
	}
	if created.Username != "" {
		plan.Username = types.StringValue(created.Username)
	} else {
		plan.Username = types.StringNull()
	}
	if created.Author != "" {
		plan.Author = types.StringValue(created.Author)
	} else {
		plan.Author = types.StringNull()
	}
	if created.Comment != "" {
		plan.Comment = types.StringValue(created.Comment)
	} else {
		plan.Comment = types.StringNull()
	}
	if created.Disable != nil {
		plan.Disable = types.BoolValue(*created.Disable)
	} else {
		plan.Disable = types.BoolNull()
	}
	if created.Origin != "" {
		plan.Origin = types.StringValue(created.Origin)
	} else {
		plan.Origin = types.StringNull()
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *smtpNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state smtpNotificationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get SMTP target from API
	target, err := r.client.Notifications.GetSMTPTarget(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading SMTP notification target",
			fmt.Sprintf("Could not read SMTP notification target %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	// Update state with values from API
	state.Server = types.StringValue(target.Server)
	state.From = types.StringValue(target.From)

	if target.Port != nil {
		state.Port = types.Int64Value(int64(*target.Port))
	}
	if target.Mode != "" {
		state.Mode = types.StringValue(target.Mode)
	} else {
		state.Mode = types.StringNull()
	}
	if target.To != nil {
		mailtoList, diags := types.ListValueFrom(ctx, types.StringType, target.To)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Mailto = mailtoList
	} else {
		state.Mailto = types.ListNull(types.StringType)
	}
	if target.MailtoUser != nil {
		mailtoUserList, diags := types.ListValueFrom(ctx, types.StringType, target.MailtoUser)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.MailtoUser = mailtoUserList
	} else {
		state.MailtoUser = types.ListNull(types.StringType)
	}
	if target.Username != "" {
		state.Username = types.StringValue(target.Username)
	} else {
		state.Username = types.StringNull()
	}
	// Don't update password from API (sensitive field)
	if target.Author != "" {
		state.Author = types.StringValue(target.Author)
	} else {
		state.Author = types.StringNull()
	}
	if target.Comment != "" {
		state.Comment = types.StringValue(target.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	if target.Disable != nil {
		state.Disable = types.BoolValue(*target.Disable)
	} else {
		state.Disable = types.BoolNull()
	}
	if target.Origin != "" {
		state.Origin = types.StringValue(target.Origin)
	} else {
		state.Origin = types.StringNull()
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *smtpNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan smtpNotificationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update SMTP target via API
	target := &notifications.SMTPTarget{
		Name:   plan.Name.ValueString(),
		Server: plan.Server.ValueString(),
		From:   plan.From.ValueString(),
	}

	// Set optional fields
	if !plan.Port.IsNull() {
		port := int(plan.Port.ValueInt64())
		target.Port = &port
	}
	if !plan.Mode.IsNull() {
		target.Mode = plan.Mode.ValueString()
	}
	if !plan.Mailto.IsNull() && !plan.Mailto.IsUnknown() {
		var mailto []string
		diags := plan.Mailto.ElementsAs(ctx, &mailto, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.To = mailto
	}
	if !plan.MailtoUser.IsNull() && !plan.MailtoUser.IsUnknown() {
		var mailtoUsers []string
		diags := plan.MailtoUser.ElementsAs(ctx, &mailtoUsers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.MailtoUser = mailtoUsers
	}
	if !plan.Username.IsNull() {
		target.Username = plan.Username.ValueString()
	}
	if !plan.Password.IsNull() {
		target.Password = plan.Password.ValueString()
	}
	if !plan.Author.IsNull() {
		target.Author = plan.Author.ValueString()
	}
	if !plan.Comment.IsNull() {
		target.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		target.Disable = &disable
	}

	err := r.client.Notifications.UpdateSMTPTarget(ctx, plan.Name.ValueString(), target)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating SMTP notification target",
			fmt.Sprintf("Could not update SMTP notification target %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	updated, err := r.client.Notifications.GetSMTPTarget(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated SMTP notification target",
			fmt.Sprintf("Updated SMTP notification target %s but could not read it back: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	plan.Server = types.StringValue(updated.Server)
	plan.From = types.StringValue(updated.From)

	if updated.Port != nil {
		plan.Port = types.Int64Value(int64(*updated.Port))
	}
	if updated.Mode != "" {
		plan.Mode = types.StringValue(updated.Mode)
	} else {
		plan.Mode = types.StringNull()
	}

	if updated.To != nil {
		mailtoList, diags := types.ListValueFrom(ctx, types.StringType, updated.To)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Mailto = mailtoList
	} else {
		plan.Mailto = types.ListNull(types.StringType)
	}

	if updated.MailtoUser != nil {
		mailtoUserList, diags := types.ListValueFrom(ctx, types.StringType, updated.MailtoUser)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.MailtoUser = mailtoUserList
	} else {
		plan.MailtoUser = types.ListNull(types.StringType)
	}

	if updated.Username != "" {
		plan.Username = types.StringValue(updated.Username)
	} else {
		plan.Username = types.StringNull()
	}
	if updated.Author != "" {
		plan.Author = types.StringValue(updated.Author)
	} else {
		plan.Author = types.StringNull()
	}
	if updated.Comment != "" {
		plan.Comment = types.StringValue(updated.Comment)
	} else {
		plan.Comment = types.StringNull()
	}
	if updated.Disable != nil {
		plan.Disable = types.BoolValue(*updated.Disable)
	} else {
		plan.Disable = types.BoolNull()
	}
	if updated.Origin != "" {
		plan.Origin = types.StringValue(updated.Origin)
	} else {
		plan.Origin = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *smtpNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state smtpNotificationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete SMTP target via API
	err := r.client.Notifications.DeleteSMTPTarget(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting SMTP notification target",
			fmt.Sprintf("Could not delete SMTP notification target %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *smtpNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
