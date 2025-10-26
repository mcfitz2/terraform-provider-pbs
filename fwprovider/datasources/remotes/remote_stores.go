/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package remotes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
)

var (
	_ datasource.DataSource              = &remoteStoresDataSource{}
	_ datasource.DataSourceWithConfigure = &remoteStoresDataSource{}
)

// NewRemoteStoresDataSource is a helper function to simplify the provider implementation.
func NewRemoteStoresDataSource() datasource.DataSource {
	return &remoteStoresDataSource{}
}

// remoteStoresDataSource is the data source implementation.
type remoteStoresDataSource struct {
	client *pbs.Client
}

// remoteStoresDataSourceModel maps the data source schema data.
type remoteStoresDataSourceModel struct {
	RemoteName types.String   `tfsdk:"remote_name"`
	Stores     []types.String `tfsdk:"stores"`
}

// Metadata returns the data source type name.
func (d *remoteStoresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_stores"
}

// Schema defines the schema for the data source.
func (d *remoteStoresDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available datastores on a remote PBS server.",
		MarkdownDescription: `Lists available datastores on a remote PBS server.

This data source scans the specified remote server and returns a list of datastore names that are accessible. 
Useful for validating sync job configurations or discovering available backup targets on a remote PBS instance.`,
		Attributes: map[string]schema.Attribute{
			"remote_name": schema.StringAttribute{
				Description:         "The name of the configured remote to scan.",
				MarkdownDescription: "The name of the configured remote to scan.",
				Required:            true,
			},
			"stores": schema.ListAttribute{
				Description:         "List of datastore names available on the remote server.",
				MarkdownDescription: "List of datastore names available on the remote server.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *remoteStoresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.Resource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = cfg.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *remoteStoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state remoteStoresDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stores, err := d.client.Remotes.ListRemoteStores(ctx, state.RemoteName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error scanning remote stores",
			fmt.Sprintf("Could not scan stores on remote %s: %s", state.RemoteName.ValueString(), err.Error()),
		)
		return
	}

	// Extract store names and convert to []types.String
	storesList := make([]types.String, 0, len(stores))
	for _, store := range stores {
		storesList = append(storesList, types.StringValue(store.Name))
	}
	state.Stores = storesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
