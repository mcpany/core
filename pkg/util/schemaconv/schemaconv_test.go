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

"google.golang.org/protobuf/reflect/protoreflect"
"google.golang.org/protobuf/reflect/protodesc"
"google.golang.org/protobuf/types/descriptorpb"

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
assert.Equal(t, "string", s1.Fields["type"].GetStringValue())
assert.Equal(t, "a string param", s1.Fields["description"].GetStringValue())

param2, ok := properties.Fields["param2"]
require.True(t, ok)
s2 := param2.GetStructValue()
require.NotNil(t, s2)
assert.Equal(t, "integer", s2.Fields["type"].GetStringValue())
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

func createTestMethodDescriptor(t *testing.T) protoreflect.MethodDescriptor {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("schemaconv_test"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("string_field"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Name: proto.String("float_field"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_FLOAT.Enum()},
					{Name: proto.String("double_field"), Number: proto.Int32(3), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()},
					{Name: proto.String("bool_field"), Number: proto.Int32(4), Type: descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum()},
					{Name: proto.String("int32_field"), Number: proto.Int32(5), Type: descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()},
					{Name: proto.String("int64_field"), Number: proto.Int32(6), Type: descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()},
					{Name: proto.String("uint32_field"), Number: proto.Int32(7), Type: descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum()},
					{Name: proto.String("uint64_field"), Number: proto.Int32(8), Type: descriptorpb.FieldDescriptorProto_TYPE_UINT64.Enum()},
					{Name: proto.String("sint32_field"), Number: proto.Int32(9), Type: descriptorpb.FieldDescriptorProto_TYPE_SINT32.Enum()},
					{Name: proto.String("sint64_field"), Number: proto.Int32(10), Type: descriptorpb.FieldDescriptorProto_TYPE_SINT64.Enum()},
					{Name: proto.String("fixed32_field"), Number: proto.Int32(11), Type: descriptorpb.FieldDescriptorProto_TYPE_FIXED32.Enum()},
					{Name: proto.String("fixed64_field"), Number: proto.Int32(12), Type: descriptorpb.FieldDescriptorProto_TYPE_FIXED64.Enum()},
					{Name: proto.String("sfixed32_field"), Number: proto.Int32(13), Type: descriptorpb.FieldDescriptorProto_TYPE_SFIXED32.Enum()},
					{Name: proto.String("sfixed64_field"), Number: proto.Int32(14), Type: descriptorpb.FieldDescriptorProto_TYPE_SFIXED64.Enum()},
				},
			},
			{
				Name: proto.String("TestResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("result"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("TestService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("TestMethod"),
						InputType:  proto.String(".schemaconv_test.TestRequest"),
						OutputType: proto.String(".schemaconv_test.TestResponse"),
					},
				},
			},
		},
	}

	file, err := protodesc.NewFile(fd, nil)
	require.NoError(t, err)

	return file.Services().Get(0).Methods().Get(0)
}

func TestMethodDescriptorToProtoProperties(t *testing.T) {
	testMethod := createTestMethodDescriptor(t)

	properties, err := MethodDescriptorToProtoProperties(testMethod)
	require.NoError(t, err)

	assert.Len(t, properties.Fields, 14)

	assert.Equal(t, "string", properties.Fields["string_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "number", properties.Fields["float_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "number", properties.Fields["double_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "boolean", properties.Fields["bool_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["int32_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["int64_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["uint32_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["uint64_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["sint32_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["sint64_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["fixed32_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["fixed64_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["sfixed32_field"].GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "integer", properties.Fields["sfixed64_field"].GetStructValue().Fields["type"].GetStringValue())
}

func TestMethodOutputDescriptorToProtoProperties(t *testing.T) {
	testMethod := createTestMethodDescriptor(t)

	properties, err := MethodOutputDescriptorToProtoProperties(testMethod)
	require.NoError(t, err)

	assert.Len(t, properties.Fields, 1)

	assert.Equal(t, "string", properties.Fields["result"].GetStructValue().Fields["type"].GetStringValue())
}
