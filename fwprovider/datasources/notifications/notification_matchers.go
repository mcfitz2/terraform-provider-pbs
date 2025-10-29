/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &notificationMatchersDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationMatchersDataSource{}
)

// NewNotificationMatchersDataSource is a helper function to simplify the provider implementation.
func NewNotificationMatchersDataSource() datasource.DataSource {
	return &notificationMatchersDataSource{}
}

// notificationMatchersDataSource is the data source implementation.
type notificationMatchersDataSource struct {
	client *pbs.Client
}

// notificationMatchersDataSourceModel maps the data source schema data.
type notificationMatchersDataSourceModel struct {
	Matchers []notificationMatcherModel `tfsdk:"matchers"`
}

// notificationMatcherModel represents a single notification matcher in the list
type notificationMatcherModel struct {
	Name          types.String `tfsdk:"name"`
	Targets       types.List   `tfsdk:"targets"`
	MatchSeverity types.List   `tfsdk:"match_severity"`
	MatchField    types.List   `tfsdk:"match_field"`
	MatchCalendar types.List   `tfsdk:"match_calendar"`
	Mode          types.String `tfsdk:"mode"`
	InvertMatch   types.Bool   `tfsdk:"invert_match"`
	Comment       types.String `tfsdk:"comment"`
	Disable       types.Bool   `tfsdk:"disable"`
	Origin        types.String `tfsdk:"origin"`
}

// Metadata returns the data source type name.
func (d *notificationMatchersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_matchers"
}

// Schema defines the schema for the data source.
func (d *notificationMatchersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all notification matcher (routing rule) configurations from Proxmox Backup Server.",
		MarkdownDescription: "Lists all notification matcher (routing rule) configurations from Proxmox Backup Server. Matchers route notifications to specific endpoints based on severity, field values, or calendar events.",

		Attributes: map[string]schema.Attribute{
			"matchers": schema.ListNestedAttribute{
				Description:         "List of notification matchers.",
				MarkdownDescription: "List of notification matchers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique name identifier for the notification matcher.",
							MarkdownDescription: "The unique name identifier for the notification matcher.",
							Computed:            true,
						},
						"targets": schema.ListAttribute{
							Description:         "List of notification endpoint names to send notifications to.",
							MarkdownDescription: "List of notification endpoint names to send notifications to.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"match_severity": schema.ListAttribute{
							Description:         "List of severities to match.",
							MarkdownDescription: "List of severities to match (info, notice, warning, error).",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"match_field": schema.ListAttribute{
							Description:         "List of field=value pairs to match.",
							MarkdownDescription: "List of field=value pairs to match (e.g., `type=gc`, `hostname=pbs1`).",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"match_calendar": schema.ListAttribute{
							Description:         "List of calendar event IDs to match.",
							MarkdownDescription: "List of calendar event IDs to match.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"mode": schema.StringAttribute{
							Description:         "Matching mode for multiple filters.",
							MarkdownDescription: "Matching mode: `all` (AND logic) or `any` (OR logic). Default is `all`.",
							Computed:            true,
						},
						"invert_match": schema.BoolAttribute{
							Description:         "Whether to invert the match logic.",
							MarkdownDescription: "Whether to invert the match logic (send notification when criteria do NOT match).",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							Description:         "A comment describing this matcher.",
							MarkdownDescription: "A comment describing this matcher.",
							Computed:            true,
						},
						"disable": schema.BoolAttribute{
							Description:         "Whether this matcher is disabled.",
							MarkdownDescription: "Whether this matcher is disabled.",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							Description:         "The origin of this matcher configuration.",
							MarkdownDescription: "The origin of this matcher configuration (user-api, user-file, etc.).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *notificationMatchersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *notificationMatchersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state notificationMatchersDataSourceModel

	matchersList, err := d.client.Notifications.ListNotificationMatchers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Notification Matchers",
			fmt.Sprintf("Could not list notification matchers: %s", err.Error()),
		)
		return
	}

	// Map API response to state
	state.Matchers = make([]notificationMatcherModel, 0, len(matchersList))
	for _, matcher := range matchersList {
		m := notificationMatcherModel{
			Name:        types.StringValue(matcher.Name),
			Disable:     boolValueOrNull(matcher.Disable),
			Comment:     stringValueOrNull(matcher.Comment),
			Origin:      stringValueOrNull(matcher.Origin),
			Mode:        stringValueOrNull(matcher.Mode),
			InvertMatch: boolValueOrNull(matcher.InvertMatch),
		}

		// Convert string slices to lists
		if matcher.Targets != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, matcher.Targets)
			resp.Diagnostics.Append(d...)
			m.Targets = list
		} else {
			m.Targets = types.ListNull(types.StringType)
		}

		if matcher.MatchSeverity != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, matcher.MatchSeverity)
			resp.Diagnostics.Append(d...)
			m.MatchSeverity = list
		} else {
			m.MatchSeverity = types.ListNull(types.StringType)
		}

		if matcher.MatchField != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, matcher.MatchField)
			resp.Diagnostics.Append(d...)
			m.MatchField = list
		} else {
			m.MatchField = types.ListNull(types.StringType)
		}

		if matcher.MatchCalendar != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, matcher.MatchCalendar)
			resp.Diagnostics.Append(d...)
			m.MatchCalendar = list
		} else {
			m.MatchCalendar = types.ListNull(types.StringType)
		}

		state.Matchers = append(state.Matchers, m)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
