/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	Mailto     types.String `tfsdk:"mailto"`
	MailtoUser types.String `tfsdk:"mailto_user"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Author     types.String `tfsdk:"author"`
	Comment    types.String `tfsdk:"comment"`
	Disable    types.Bool   `tfsdk:"disable"`
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
				Validators: []validator.String{
					stringvalidator.OneOf("insecure", "starttls", "tls"),
				},
			},
			"from_address": schema.StringAttribute{
				Description:         "Sender email address.",
				MarkdownDescription: "Sender email address. This will appear as the 'From' address in notification emails.",
				Required:            true,
			},
			"mailto": schema.StringAttribute{
				Description:         "Recipient email address(es).",
				MarkdownDescription: "Recipient email address(es). Multiple addresses can be specified separated by commas or semicolons.",
				Optional:            true,
			},
			"mailto_user": schema.StringAttribute{
				Description:         "User(s) from PBS user database to receive notifications.",
				MarkdownDescription: "User(s) from PBS user database to receive notifications. Email addresses will be looked up from user configuration.",
				Optional:            true,
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
	if !plan.Mode.IsNull() {
		target.Mode = plan.Mode.ValueString()
	}
	if !plan.Mailto.IsNull() {
		// PBS 4.0: mailto is an array, split comma-separated string
		mailtoStr := plan.Mailto.ValueString()
		if mailtoStr != "" {
			target.To = strings.Split(mailtoStr, ",")
			// Trim spaces from each email address
			for i := range target.To {
				target.To[i] = strings.TrimSpace(target.To[i])
			}
		}
	}
	if !plan.MailtoUser.IsNull() {
		target.MailtoUser = plan.MailtoUser.ValueString()
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
	}
	if len(target.To) > 0 {
		// PBS 4.0: mailto is an array, join into comma-separated string
		state.Mailto = types.StringValue(strings.Join(target.To, ","))
	}
	if target.MailtoUser != "" {
		state.MailtoUser = types.StringValue(target.MailtoUser)
	}
	if target.Username != "" {
		state.Username = types.StringValue(target.Username)
	}
	// Don't update password from API (sensitive field)
	if target.Author != "" {
		state.Author = types.StringValue(target.Author)
	}
	if target.Comment != "" {
		state.Comment = types.StringValue(target.Comment)
	}
	if target.Disable != nil {
		state.Disable = types.BoolValue(*target.Disable)
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
	if !plan.Mailto.IsNull() {
		// PBS 4.0: mailto is an array, split comma-separated string
		mailtoStr := plan.Mailto.ValueString()
		if mailtoStr != "" {
			target.To = strings.Split(mailtoStr, ",")
			// Trim spaces from each email address
			for i := range target.To {
				target.To[i] = strings.TrimSpace(target.To[i])
			}
		}
	}
	if !plan.MailtoUser.IsNull() {
		target.MailtoUser = plan.MailtoUser.ValueString()
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

	// Set state
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
