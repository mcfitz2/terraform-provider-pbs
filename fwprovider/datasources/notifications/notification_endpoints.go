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
	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &notificationEndpointsDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationEndpointsDataSource{}
)

// NewNotificationEndpointsDataSource is a helper function to simplify the provider implementation.
func NewNotificationEndpointsDataSource() datasource.DataSource {
	return &notificationEndpointsDataSource{}
}

// notificationEndpointsDataSource is the data source implementation.
type notificationEndpointsDataSource struct {
	client *pbs.Client
}

// notificationEndpointsDataSourceModel maps the data source schema data.
type notificationEndpointsDataSourceModel struct {
	Endpoints []notificationEndpointModel `tfsdk:"endpoints"`
}

// notificationEndpointModel represents a single notification endpoint in the list
type notificationEndpointModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Disable types.Bool   `tfsdk:"disable"`
	Comment types.String `tfsdk:"comment"`
	Origin  types.String `tfsdk:"origin"`

	// SMTP fields
	Server     types.String `tfsdk:"server"`
	Port       types.Int64  `tfsdk:"port"`
	From       types.String `tfsdk:"from_address"`
	Mailto     types.List   `tfsdk:"mailto"`
	MailtoUser types.List   `tfsdk:"mailto_user"`
	Mode       types.String `tfsdk:"mode"`
	Username   types.String `tfsdk:"username"`
	Author     types.String `tfsdk:"author"`

	// Gotify/Webhook fields
	URL types.String `tfsdk:"url"`

	// Webhook fields
	Body    types.String `tfsdk:"body"`
	Method  types.String `tfsdk:"method"`
	Headers types.Map    `tfsdk:"headers"`
}

// Metadata returns the data source type name.
func (d *notificationEndpointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_endpoints"
}

// Schema defines the schema for the data source.
func (d *notificationEndpointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all notification endpoint configurations from Proxmox Backup Server.",
		MarkdownDescription: "Lists all notification endpoint configurations from Proxmox Backup Server. Includes Gotify, SMTP, Sendmail, and Webhook endpoints.",

		Attributes: map[string]schema.Attribute{
			"endpoints": schema.ListNestedAttribute{
				Description:         "List of notification endpoints.",
				MarkdownDescription: "List of notification endpoints.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique name identifier for the notification endpoint.",
							MarkdownDescription: "The unique name identifier for the notification endpoint.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							Description:         "The type of notification endpoint.",
							MarkdownDescription: "The type of notification endpoint (gotify, smtp, sendmail, or webhook).",
							Computed:            true,
						},
						"disable": schema.BoolAttribute{
							Description:         "Whether this endpoint is disabled.",
							MarkdownDescription: "Whether this endpoint is disabled.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							Description:         "A comment describing this endpoint.",
							MarkdownDescription: "A comment describing this endpoint.",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							Description:         "The origin of this endpoint configuration.",
							MarkdownDescription: "The origin of this endpoint configuration (user-api, user-file, etc.).",
							Computed:            true,
						},

						// SMTP/Sendmail fields
						"server": schema.StringAttribute{
							Description:         "SMTP server address (SMTP only).",
							MarkdownDescription: "SMTP server address (SMTP only).",
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							Description:         "SMTP server port (SMTP only).",
							MarkdownDescription: "SMTP server port (SMTP only).",
							Computed:            true,
						},
						"from_address": schema.StringAttribute{
							Description:         "From email address (SMTP/Sendmail only).",
							MarkdownDescription: "From email address (SMTP/Sendmail only).",
							Computed:            true,
						},
						"mailto": schema.ListAttribute{
							Description:         "List of recipient email addresses (SMTP/Sendmail only).",
							MarkdownDescription: "List of recipient email addresses (SMTP/Sendmail only).",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"mailto_user": schema.ListAttribute{
							Description:         "List of PBS user IDs to notify (SMTP/Sendmail only).",
							MarkdownDescription: "List of PBS user IDs to notify (SMTP/Sendmail only).",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"mode": schema.StringAttribute{
							Description:         "SMTP connection mode (SMTP only).",
							MarkdownDescription: "SMTP connection mode: insecure, starttls, or tls (SMTP only).",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							Description:         "SMTP authentication username (SMTP only).",
							MarkdownDescription: "SMTP authentication username (SMTP only).",
							Computed:            true,
						},
						"author": schema.StringAttribute{
							Description:         "Email author/sender name (SMTP/Sendmail only).",
							MarkdownDescription: "Email author/sender name (SMTP/Sendmail only).",
							Computed:            true,
						},

						// Gotify/Webhook fields
						"url": schema.StringAttribute{
							Description:         "Target URL (Gotify/Webhook only).",
							MarkdownDescription: "Target URL (Gotify/Webhook only).",
							Computed:            true,
						},

						// Webhook fields
						"body": schema.StringAttribute{
							Description:         "Webhook request body template (Webhook only).",
							MarkdownDescription: "Webhook request body template (Webhook only).",
							Computed:            true,
						},
						"method": schema.StringAttribute{
							Description:         "HTTP method for webhook (Webhook only).",
							MarkdownDescription: "HTTP method for webhook: POST or PUT (Webhook only).",
							Computed:            true,
						},
						"headers": schema.MapAttribute{
							Description:         "Custom HTTP headers for webhook (Webhook only).",
							MarkdownDescription: "Custom HTTP headers for webhook (Webhook only).",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *notificationEndpointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *notificationEndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state notificationEndpointsDataSourceModel

	// PBS doesn't have a single endpoint for all notification endpoints
	// We need to query each type separately
	var allEndpoints []notificationEndpointModel

	// Fetch Gotify endpoints
	gotifyTargets, err := d.client.Notifications.ListGotifyTargets(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Gotify Endpoints",
			fmt.Sprintf("Could not list Gotify endpoints: %s", err.Error()),
		)
		return
	}
	for _, target := range gotifyTargets {
		endpoint := notificationEndpointModel{
			Name:       types.StringValue(target.Name),
			Type:       types.StringValue(string(notifications.NotificationTargetTypeGotify)),
			Disable:    boolValueOrNull(target.Disable),
			Comment:    stringValueOrNull(target.Comment),
			Origin:     stringValueOrNull(target.Origin),
			URL:        stringValueOrNull(target.Server),
			Headers:    types.MapNull(types.StringType),  // Not applicable for Gotify
			Mailto:     types.ListNull(types.StringType), // Not applicable for Gotify
			MailtoUser: types.ListNull(types.StringType), // Not applicable for Gotify
		}
		allEndpoints = append(allEndpoints, endpoint)
	}

	// Fetch SMTP endpoints
	smtpTargets, err := d.client.Notifications.ListSMTPTargets(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SMTP Endpoints",
			fmt.Sprintf("Could not list SMTP endpoints: %s", err.Error()),
		)
		return
	}
	for _, target := range smtpTargets {
		endpoint := notificationEndpointModel{
			Name:     types.StringValue(target.Name),
			Type:     types.StringValue(string(notifications.NotificationTargetTypeSMTP)),
			Disable:  boolValueOrNull(target.Disable),
			Comment:  stringValueOrNull(target.Comment),
			Origin:   stringValueOrNull(target.Origin),
			Server:   stringValueOrNull(target.Server),
			Port:     intValueOrNull(target.Port),
			From:     stringValueOrNull(target.From),
			Mode:     stringValueOrNull(target.Mode),
			Username: stringValueOrNull(target.Username),
			Author:   stringValueOrNull(target.Author),
			Headers:  types.MapNull(types.StringType), // Not applicable for SMTP
		}

		if target.To != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, target.To)
			resp.Diagnostics.Append(d...)
			endpoint.Mailto = list
		} else {
			endpoint.Mailto = types.ListNull(types.StringType)
		}

		if target.MailtoUser != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, target.MailtoUser)
			resp.Diagnostics.Append(d...)
			endpoint.MailtoUser = list
		} else {
			endpoint.MailtoUser = types.ListNull(types.StringType)
		}

		allEndpoints = append(allEndpoints, endpoint)
	}

	// Fetch Sendmail endpoints
	sendmailTargets, err := d.client.Notifications.ListSendmailTargets(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Sendmail Endpoints",
			fmt.Sprintf("Could not list Sendmail endpoints: %s", err.Error()),
		)
		return
	}
	for _, target := range sendmailTargets {
		endpoint := notificationEndpointModel{
			Name:    types.StringValue(target.Name),
			Type:    types.StringValue(string(notifications.NotificationTargetTypeSendmail)),
			Disable: boolValueOrNull(target.Disable),
			Comment: stringValueOrNull(target.Comment),
			Origin:  stringValueOrNull(target.Origin),
			From:    stringValueOrNull(target.From),
			Author:  stringValueOrNull(target.Author),
			Headers: types.MapNull(types.StringType), // Not applicable for Sendmail
		}

		if target.Mailto != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, target.Mailto)
			resp.Diagnostics.Append(d...)
			endpoint.Mailto = list
		} else {
			endpoint.Mailto = types.ListNull(types.StringType)
		}

		if target.MailtoUser != nil {
			list, d := types.ListValueFrom(ctx, types.StringType, target.MailtoUser)
			resp.Diagnostics.Append(d...)
			endpoint.MailtoUser = list
		} else {
			endpoint.MailtoUser = types.ListNull(types.StringType)
		}

		allEndpoints = append(allEndpoints, endpoint)
	}

	// Fetch Webhook endpoints
	webhookTargets, err := d.client.Notifications.ListWebhookTargets(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Webhook Endpoints",
			fmt.Sprintf("Could not list Webhook endpoints: %s", err.Error()),
		)
		return
	}
	for _, target := range webhookTargets {
		endpoint := notificationEndpointModel{
			Name:       types.StringValue(target.Name),
			Type:       types.StringValue(string(notifications.NotificationTargetTypeWebhook)),
			Disable:    boolValueOrNull(target.Disable),
			Comment:    stringValueOrNull(target.Comment),
			Origin:     stringValueOrNull(target.Origin),
			URL:        stringValueOrNull(target.URL),
			Body:       stringValueOrNull(target.Body),
			Method:     stringValueOrNull(target.Method),
			Mailto:     types.ListNull(types.StringType), // Not applicable for Webhook
			MailtoUser: types.ListNull(types.StringType), // Not applicable for Webhook
		}

		if target.Headers != nil && len(target.Headers) > 0 {
			headers, d := types.MapValueFrom(ctx, types.StringType, target.Headers)
			resp.Diagnostics.Append(d...)
			endpoint.Headers = headers
		} else {
			endpoint.Headers = types.MapNull(types.StringType)
		}

		allEndpoints = append(allEndpoints, endpoint)
	}

	state.Endpoints = allEndpoints
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
