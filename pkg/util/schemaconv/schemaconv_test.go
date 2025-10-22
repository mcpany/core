/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package schemaconv

import (
	"testing"

	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockConfigParameter struct {
	schema *configv1.ParameterSchema
}

func (m *mockConfigParameter) GetSchema() *configv1.ParameterSchema {
	return m.schema
}

type mockMcpFieldParameter struct {
	name        string
	description string
	typ         string
}

func (m *mockMcpFieldParameter) GetName() string {
	return m.name
}

func (m *mockMcpFieldParameter) GetDescription() string {
	return m.description
}

func (m *mockMcpFieldParameter) GetType() string {
	return m.typ
}

func TestConfigSchemaToProtoProperties(t *testing.T) {
	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])
	intType := configv1.ParameterType(configv1.ParameterType_value["INTEGER"])
	params := []*mockConfigParameter{
		{schema: configv1.ParameterSchema_builder{Name: proto.String("param1"), Description: proto.String("a string param"), Type: &stringType}.Build()},
		{schema: configv1.ParameterSchema_builder{Name: proto.String("param2"), Description: proto.String("an int param"), Type: &intType}.Build()},
	}

	properties, err := ConfigSchemaToProtoProperties(params)
	require.NoError(t, err)
	assert.Len(t, properties.Fields, 2)

	param1, ok := properties.Fields["param1"]
	require.True(t, ok)
	s1 := param1.GetStructValue()
	require.NotNil(t, s1)
	assert.Equal(t, "STRING", s1.Fields["type"].GetStringValue())
	assert.Equal(t, "a string param", s1.Fields["description"].GetStringValue())

	param2, ok := properties.Fields["param2"]
	require.True(t, ok)
	s2 := param2.GetStructValue()
	require.NotNil(t, s2)
	assert.Equal(t, "INTEGER", s2.Fields["type"].GetStringValue())
	assert.Equal(t, "an int param", s2.Fields["description"].GetStringValue())
}

func TestMcpFieldsToProtoProperties(t *testing.T) {
	params := []*mockMcpFieldParameter{
		{name: "field1", description: "a string field", typ: "TYPE_STRING"},
		{name: "field2", description: "an int field", typ: "TYPE_INT32"},
	}

	properties, err := McpFieldsToProtoProperties(params)
	require.NoError(t, err)
	assert.Len(t, properties.Fields, 2)

	field1, ok := properties.Fields["field1"]
	require.True(t, ok)
	s1 := field1.GetStructValue()
	require.NotNil(t, s1)
	assert.Equal(t, "string", s1.Fields["type"].GetStringValue())
	assert.Equal(t, "a string field", s1.Fields["description"].GetStringValue())

	field2, ok := properties.Fields["field2"]
	require.True(t, ok)
	s2 := field2.GetStructValue()
	require.NotNil(t, s2)
	assert.Equal(t, "integer", s2.Fields["type"].GetStringValue())
	assert.Equal(t, "an int field", s2.Fields["description"].GetStringValue())
}
