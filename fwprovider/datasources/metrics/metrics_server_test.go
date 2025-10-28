/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package metrics

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestMetricsServerDataSource(t *testing.T) {
	t.Parallel()

	ds := NewMetricsServerDataSource()

	if ds == nil {
		t.Fatal("NewMetricsServerDataSource() returned nil")
	}

	_, ok := ds.(datasource.DataSource)
	if !ok {
		t.Fatal("NewMetricsServerDataSource() did not return a datasource.DataSource")
	}

	_, ok = ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Fatal("NewMetricsServerDataSource() did not return a datasource.DataSourceWithConfigure")
	}
}

func TestMetricsServersDataSource(t *testing.T) {
	t.Parallel()

	ds := NewMetricsServersDataSource()

	if ds == nil {
		t.Fatal("NewMetricsServersDataSource() returned nil")
	}

	_, ok := ds.(datasource.DataSource)
	if !ok {
		t.Fatal("NewMetricsServersDataSource() did not return a datasource.DataSource")
	}

	_, ok = ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Fatal("NewMetricsServersDataSource() did not return a datasource.DataSourceWithConfigure")
	}
}
