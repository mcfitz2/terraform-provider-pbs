/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package config provides the global provider's configuration for all resources and datasources.
package config

import "github.com/micah/terraform-provider-pbs/pbs"

// DataSource is the global configuration for all datasources.
type DataSource struct {
	Client *pbs.Client
}

// Resource is the global configuration for all resources.
type Resource struct {
	Client *pbs.Client
}
