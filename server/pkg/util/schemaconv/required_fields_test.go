// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestConfigSchemaToProtoProperties_RequiredFields(t *testing.T) {
	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])
	params := []*mockConfigParameter{
		{
			schema: configv1.ParameterSchema_builder{
				Name:       proto.String("required_param"),
				Type:       &stringType,
				IsRequired: proto.Bool(true),
			}.Build(),
		},
		{
			schema: configv1.ParameterSchema_builder{
				Name:       proto.String("optional_param"),
				Type:       &stringType,
				IsRequired: proto.Bool(false),
			}.Build(),
		},
	}

	// Now the function returns properties and required fields.
	properties, required := ConfigSchemaToProtoProperties(params)

	assert.NotNil(t, properties)
	assert.Len(t, required, 1)
	assert.Contains(t, required, "required_param")
	assert.NotContains(t, required, "optional_param")
}
