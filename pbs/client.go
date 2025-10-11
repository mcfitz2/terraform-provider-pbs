/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package pbs provides the main client interface for Proxmox Backup Server
package pbs

import (
	"github.com/micah/terraform-provider-pbs/pbs/api"
	"github.com/micah/terraform-provider-pbs/pbs/datastores"
	"github.com/micah/terraform-provider-pbs/pbs/endpoints"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
	"github.com/micah/terraform-provider-pbs/pbs/metrics"
	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// Client represents the main PBS client interface
type Client struct {
	api           *api.Client
	Endpoints     *endpoints.Client
	Datastores    *datastores.Client
	Metrics       *metrics.Client
	Notifications *notifications.Client
	Jobs          *jobs.Client
}

// NewClient creates a new PBS client
func NewClient(creds api.Credentials, opts api.ClientOptions) (*Client, error) {
	apiClient, err := api.NewClient(creds, opts)
	if err != nil {
		return nil, err
	}

	return &Client{
		api:           apiClient,
		Endpoints:     endpoints.NewClient(apiClient),
		Datastores:    datastores.NewClient(apiClient),
		Metrics:       metrics.NewClient(apiClient),
		Notifications: notifications.NewClient(apiClient),
		Jobs:          jobs.NewClient(apiClient),
	}, nil
}
