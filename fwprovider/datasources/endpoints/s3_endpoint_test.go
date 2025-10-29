/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package endpoints

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/stretchr/testify/assert"
)

// TestS3EndpointDataSourceSchema verifies the s3_endpoint data source schema
func TestS3EndpointDataSourceSchema(t *testing.T) {
	ds := NewS3EndpointDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	ds.Schema(nil, schemaReq, schemaResp)

	assert.NotNil(t, schemaResp.Schema)
	assert.NotEmpty(t, schemaResp.Schema.Description)

	// Verify required attributes
	idAttr, exists := schemaResp.Schema.Attributes["id"]
	assert.True(t, exists, "id attribute should exist")
	assert.True(t, idAttr.(schema.StringAttribute).Required, "id should be required")

	// Verify computed attributes
	accessKeyAttr, exists := schemaResp.Schema.Attributes["access_key"]
	assert.True(t, exists, "access_key attribute should exist")
	assert.True(t, accessKeyAttr.(schema.StringAttribute).Computed, "access_key should be computed")
	assert.True(t, accessKeyAttr.(schema.StringAttribute).Sensitive, "access_key should be sensitive")

	endpointAttr, exists := schemaResp.Schema.Attributes["endpoint"]
	assert.True(t, exists, "endpoint attribute should exist")
	assert.True(t, endpointAttr.(schema.StringAttribute).Computed, "endpoint should be computed")

	regionAttr, exists := schemaResp.Schema.Attributes["region"]
	assert.True(t, exists, "region attribute should exist")
	assert.True(t, regionAttr.(schema.StringAttribute).Computed, "region should be computed")

	fingerprintAttr, exists := schemaResp.Schema.Attributes["fingerprint"]
	assert.True(t, exists, "fingerprint attribute should exist")
	assert.True(t, fingerprintAttr.(schema.StringAttribute).Computed, "fingerprint should be computed")

	portAttr, exists := schemaResp.Schema.Attributes["port"]
	assert.True(t, exists, "port attribute should exist")
	assert.True(t, portAttr.(schema.Int64Attribute).Computed, "port should be computed")

	pathStyleAttr, exists := schemaResp.Schema.Attributes["path_style"]
	assert.True(t, exists, "path_style attribute should exist")
	assert.True(t, pathStyleAttr.(schema.BoolAttribute).Computed, "path_style should be computed")

	providerQuirksAttr, exists := schemaResp.Schema.Attributes["provider_quirks"]
	assert.True(t, exists, "provider_quirks attribute should exist")
	assert.True(t, providerQuirksAttr.(schema.SetAttribute).Computed, "provider_quirks should be computed")
}

// TestS3EndpointsDataSourceSchema verifies the s3_endpoints data source schema
func TestS3EndpointsDataSourceSchema(t *testing.T) {
	ds := NewS3EndpointsDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	ds.Schema(nil, schemaReq, schemaResp)

	assert.NotNil(t, schemaResp.Schema)
	assert.NotEmpty(t, schemaResp.Schema.Description)

	// Verify endpoints list attribute
	endpointsAttr, exists := schemaResp.Schema.Attributes["endpoints"]
	assert.True(t, exists, "endpoints attribute should exist")
	assert.True(t, endpointsAttr.(schema.ListNestedAttribute).Computed, "endpoints should be computed")

	// Verify nested attributes in the endpoints list
	nestedObj := endpointsAttr.(schema.ListNestedAttribute).NestedObject
	assert.NotNil(t, nestedObj.Attributes)

	// Check all nested attributes exist and are computed
	nestedAttrs := []string{"id", "access_key", "endpoint", "region", "fingerprint", "port", "path_style", "provider_quirks"}
	for _, attrName := range nestedAttrs {
		_, exists := nestedObj.Attributes[attrName]
		assert.True(t, exists, "%s attribute should exist in nested object", attrName)
	}
}
