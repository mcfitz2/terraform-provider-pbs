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
	_ resource.Resource                = &notificationMatcherResource{}
	_ resource.ResourceWithConfigure   = &notificationMatcherResource{}
	_ resource.ResourceWithImportState = &notificationMatcherResource{}
)

// NewNotificationMatcherResource is a helper function to simplify the provider implementation.
func NewNotificationMatcherResource() resource.Resource {
	return &notificationMatcherResource{}
}

// notificationMatcherResource is the resource implementation.
type notificationMatcherResource struct {
	client *pbs.Client
}

// notificationMatcherResourceModel maps the resource schema data.
type notificationMatcherResourceModel struct {
	Name          types.String `tfsdk:"name"`
	Targets       types.List   `tfsdk:"targets"`
	MatchSeverity types.List   `tfsdk:"match_severity"`
	MatchField    types.List   `tfsdk:"match_field"`
	MatchCalendar types.List   `tfsdk:"match_calendar"`
	Mode          types.String `tfsdk:"mode"`
	InvertMatch   types.Bool   `tfsdk:"invert_match"`
	Comment       types.String `tfsdk:"comment"`
	Disable       types.Bool   `tfsdk:"disable"`
}

// Metadata returns the resource type name.
func (r *notificationMatcherResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_matcher"
}

// Schema defines the schema for the resource.
func (r *notificationMatcherResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a notification matcher (routing rule).",
		MarkdownDescription: `Manages a notification matcher.

Notification matchers define rules for routing notification events to specific targets or endpoints. 
They can filter notifications based on severity, custom fields, and calendar schedules. This allows 
for sophisticated notification routing, such as sending critical errors to on-call staff or filtering 
informational messages to specific channels.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique name identifier for the notification matcher.",
				MarkdownDescription: "The unique name identifier for the notification matcher.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"targets": schema.ListAttribute{
				Description:         "List of notification target or endpoint names to route matching notifications to.",
				MarkdownDescription: "List of notification target or endpoint names to route matching notifications to.",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"match_severity": schema.ListAttribute{
				Description:         "List of severity levels to match (info, notice, warning, error).",
				MarkdownDescription: "List of severity levels to match. Valid values: `info`, `notice`, `warning`, `error`.",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"match_field": schema.ListAttribute{
				Description:         "List of field=value pairs to match against notification metadata.",
				MarkdownDescription: "List of `field=value` pairs to match against notification metadata (e.g., `type=prune`, `hostname=server01`).",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"match_calendar": schema.ListAttribute{
				Description:         "List of calendar IDs to match for time-based routing.",
				MarkdownDescription: "List of calendar IDs to match for time-based routing (requires calendar configuration in PBS).",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"mode": schema.StringAttribute{
				Description:         "Match mode: all (all conditions must match) or any (at least one condition must match).",
				MarkdownDescription: "Match mode: `all` (all conditions must match) or `any` (at least one condition must match). Defaults to `all`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "any"),
				},
			},
			"invert_match": schema.BoolAttribute{
				Description:         "Invert the matching logic (route notifications that DON'T match the criteria).",
				MarkdownDescription: "Invert the matching logic (route notifications that DON'T match the criteria). Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this notification matcher.",
				MarkdownDescription: "A comment describing this notification matcher.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this notification matcher.",
				MarkdownDescription: "Disable this notification matcher. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *notificationMatcherResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *notificationMatcherResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationMatcherResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	matcher := &notifications.NotificationMatcher{
		Name: plan.Name.ValueString(),
	}

	// Convert lists
	if !plan.Targets.IsNull() {
		var targets []string
		diags := plan.Targets.ElementsAs(ctx, &targets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.Targets = targets
	}

	if !plan.MatchSeverity.IsNull() {
		var matchSeverity []string
		diags := plan.MatchSeverity.ElementsAs(ctx, &matchSeverity, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.MatchSeverity = matchSeverity
	}

	if !plan.MatchField.IsNull() {
		var matchField []string
		diags := plan.MatchField.ElementsAs(ctx, &matchField, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.MatchField = matchField
	}

	if !plan.MatchCalendar.IsNull() {
		var matchCalendar []string
		diags := plan.MatchCalendar.ElementsAs(ctx, &matchCalendar, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.MatchCalendar = matchCalendar
	}

	if !plan.Mode.IsNull() {
		matcher.Mode = plan.Mode.ValueString()
	}

	if !plan.InvertMatch.IsNull() {
		invertMatch := plan.InvertMatch.ValueBool()
		matcher.InvertMatch = &invertMatch
	}

	if !plan.Comment.IsNull() {
		matcher.Comment = plan.Comment.ValueString()
	}

	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		matcher.Disable = &disable
	}

	err := r.client.Notifications.CreateNotificationMatcher(ctx, matcher)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating notification matcher",
			fmt.Sprintf("Could not create notification matcher %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *notificationMatcherResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationMatcherResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	matcher, err := r.client.Notifications.GetNotificationMatcher(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading notification matcher",
			fmt.Sprintf("Could not read notification matcher %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	// Update state with values from API
	state.Name = types.StringValue(matcher.Name)

	if len(matcher.Targets) > 0 {
		targets, diags := types.ListValueFrom(ctx, types.StringType, matcher.Targets)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Targets = targets
	} else {
		state.Targets = types.ListNull(types.StringType)
	}

	if len(matcher.MatchSeverity) > 0 {
		matchSeverity, diags := types.ListValueFrom(ctx, types.StringType, matcher.MatchSeverity)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.MatchSeverity = matchSeverity
	} else {
		state.MatchSeverity = types.ListNull(types.StringType)
	}

	if len(matcher.MatchField) > 0 {
		matchField, diags := types.ListValueFrom(ctx, types.StringType, matcher.MatchField)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.MatchField = matchField
	} else {
		state.MatchField = types.ListNull(types.StringType)
	}

	if len(matcher.MatchCalendar) > 0 {
		matchCalendar, diags := types.ListValueFrom(ctx, types.StringType, matcher.MatchCalendar)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.MatchCalendar = matchCalendar
	} else {
		state.MatchCalendar = types.ListNull(types.StringType)
	}

	if matcher.Mode != "" {
		state.Mode = types.StringValue(matcher.Mode)
	} else {
		state.Mode = types.StringValue("all")
	}

	if matcher.InvertMatch != nil {
		state.InvertMatch = types.BoolValue(*matcher.InvertMatch)
	} else {
		state.InvertMatch = types.BoolValue(false)
	}

	if matcher.Comment != "" {
		state.Comment = types.StringValue(matcher.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	if matcher.Disable != nil {
		state.Disable = types.BoolValue(*matcher.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *notificationMatcherResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationMatcherResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	matcher := &notifications.NotificationMatcher{
		Name: plan.Name.ValueString(),
	}

	// Convert lists
	if !plan.Targets.IsNull() {
		var targets []string
		diags := plan.Targets.ElementsAs(ctx, &targets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.Targets = targets
	}

	if !plan.MatchSeverity.IsNull() {
		var matchSeverity []string
		diags := plan.MatchSeverity.ElementsAs(ctx, &matchSeverity, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.MatchSeverity = matchSeverity
	}

	if !plan.MatchField.IsNull() {
		var matchField []string
		diags := plan.MatchField.ElementsAs(ctx, &matchField, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.MatchField = matchField
	}

	if !plan.MatchCalendar.IsNull() {
		var matchCalendar []string
		diags := plan.MatchCalendar.ElementsAs(ctx, &matchCalendar, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		matcher.MatchCalendar = matchCalendar
	}

	if !plan.Mode.IsNull() {
		matcher.Mode = plan.Mode.ValueString()
	}

	if !plan.InvertMatch.IsNull() {
		invertMatch := plan.InvertMatch.ValueBool()
		matcher.InvertMatch = &invertMatch
	}

	if !plan.Comment.IsNull() {
		matcher.Comment = plan.Comment.ValueString()
	}

	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		matcher.Disable = &disable
	}

	err := r.client.Notifications.UpdateNotificationMatcher(ctx, plan.Name.ValueString(), matcher)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating notification matcher",
			fmt.Sprintf("Could not update notification matcher %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *notificationMatcherResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationMatcherResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Notifications.DeleteNotificationMatcher(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting notification matcher",
			fmt.Sprintf("Could not delete notification matcher %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *notificationMatcherResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
