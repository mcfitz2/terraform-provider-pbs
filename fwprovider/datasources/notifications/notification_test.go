/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// TestNotificationEndpointDataSource verifies the data source implements required interfaces
func TestNotificationEndpointDataSource(t *testing.T) {
	t.Parallel()

	// Verify type assertion
	var _ datasource.DataSource = &notificationEndpointDataSource{}
	var _ datasource.DataSourceWithConfigure = &notificationEndpointDataSource{}
}

// TestNotificationEndpointsDataSource verifies the data source implements required interfaces
func TestNotificationEndpointsDataSource(t *testing.T) {
	t.Parallel()

	// Verify type assertion
	var _ datasource.DataSource = &notificationEndpointsDataSource{}
	var _ datasource.DataSourceWithConfigure = &notificationEndpointsDataSource{}
}

// TestNotificationMatcherDataSource verifies the data source implements required interfaces
func TestNotificationMatcherDataSource(t *testing.T) {
	t.Parallel()

	// Verify type assertion
	var _ datasource.DataSource = &notificationMatcherDataSource{}
	var _ datasource.DataSourceWithConfigure = &notificationMatcherDataSource{}
}

// TestNotificationMatchersDataSource verifies the data source implements required interfaces
func TestNotificationMatchersDataSource(t *testing.T) {
	t.Parallel()

	// Verify type assertion
	var _ datasource.DataSource = &notificationMatchersDataSource{}
	var _ datasource.DataSourceWithConfigure = &notificationMatchersDataSource{}
}
