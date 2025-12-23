// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	weatherv1 "github.com/mcpany/core/proto/examples/weather/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		{schema: nil},
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
	testCases := []struct {
		name         string
		typ          string
		expectedType string
	}{
		{"string field", "TYPE_STRING", "string"},
		{"int field", "TYPE_INT32", "integer"},
		{"double field", "TYPE_DOUBLE", "number"},
		{"float field", "TYPE_FLOAT", "number"},
		{"int64 field", "TYPE_INT64", "integer"},
		{"uint32 field", "TYPE_UINT32", "integer"},
		{"uint64 field", "TYPE_UINT64", "integer"},
		{"sint32 field", "TYPE_SINT32", "integer"},
		{"sint64 field", "TYPE_SINT64", "integer"},
		{"fixed32 field", "TYPE_FIXED32", "integer"},
		{"fixed64 field", "TYPE_FIXED64", "integer"},
		{"sfixed32 field", "TYPE_SFIXED32", "integer"},
		{"sfixed64 field", "TYPE_SFIXED64", "integer"},
		{"bool field", "TYPE_BOOL", "boolean"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := []*mockMcpFieldParameter{
				{name: "field1", description: tc.name, typ: tc.typ},
			}

			properties, err := McpFieldsToProtoProperties(params)
			require.NoError(t, err)
			assert.Len(t, properties.Fields, 1)

			field, ok := properties.Fields["field1"]
			require.True(t, ok)
			s := field.GetStructValue()
			require.NotNil(t, s)
			assert.Equal(t, tc.expectedType, s.Fields["type"].GetStringValue())
			assert.Equal(t, tc.name, s.Fields["description"].GetStringValue())
		})
	}
}

// mockFieldDescriptor is a mock implementation of protoreflect.FieldDescriptor for testing.
type mockFieldDescriptor struct {
	protoreflect.FieldDescriptor
	kind protoreflect.Kind
	name string
}

func (m *mockFieldDescriptor) Kind() protoreflect.Kind {
	return m.kind
}

func (m *mockFieldDescriptor) Name() protoreflect.Name {
	return protoreflect.Name(m.name)
}

func (m *mockFieldDescriptor) IsList() bool {
	return false
}

// mockFieldDescriptors is a mock implementation of protoreflect.FieldDescriptors for testing.
type mockFieldDescriptors struct {
	protoreflect.FieldDescriptors
	fields []protoreflect.FieldDescriptor
}

func (m *mockFieldDescriptors) Len() int {
	return len(m.fields)
}

func (m *mockFieldDescriptors) Get(i int) protoreflect.FieldDescriptor {
	return m.fields[i]
}

// mockMessageDescriptor is a mock implementation of protoreflect.MessageDescriptor for testing.
type mockMessageDescriptor struct {
	protoreflect.MessageDescriptor
	fields protoreflect.FieldDescriptors
}

func (m *mockMessageDescriptor) Fields() protoreflect.FieldDescriptors {
	return m.fields
}

// mockMethodDescriptor is a mock implementation of protoreflect.MethodDescriptor for testing.
type mockMethodDescriptor struct {
	protoreflect.MethodDescriptor
	input  protoreflect.MessageDescriptor
	output protoreflect.MessageDescriptor
}

func (m *mockMethodDescriptor) Input() protoreflect.MessageDescriptor {
	return m.input
}

func (m *mockMethodDescriptor) Output() protoreflect.MessageDescriptor {
	return m.output
}

func TestMethodDescriptorToProtoProperties(t *testing.T) {
	t.Run("with real proto", func(t *testing.T) {
		fileDesc := weatherv1.File_proto_examples_weather_v1_weather_proto
		require.NotNil(t, fileDesc)

		serviceDesc := fileDesc.Services().ByName("WeatherService")
		require.NotNil(t, serviceDesc)

		methodDesc := serviceDesc.Methods().ByName("GetWeather")
		require.NotNil(t, methodDesc)

		properties, err := MethodDescriptorToProtoProperties(methodDesc)
		require.NoError(t, err)
		require.Len(t, properties.Fields, 1)

		locationField, ok := properties.Fields["location"]
		require.True(t, ok)
		s1 := locationField.GetStructValue()
		require.NotNil(t, s1)
		assert.Equal(t, "string", s1.Fields["type"].GetStringValue())
	})

	t.Run("with mocks", func(t *testing.T) {
		testCases := []struct {
			name         string
			fieldKind    protoreflect.Kind
			expectedType string
		}{
			{"string", protoreflect.StringKind, "string"},
			{"double", protoreflect.DoubleKind, "number"},
			{"float", protoreflect.FloatKind, "number"},
			{"int32", protoreflect.Int32Kind, "integer"},
			{"bool", protoreflect.BoolKind, "boolean"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockMethod := &mockMethodDescriptor{
					input: &mockMessageDescriptor{
						fields: &mockFieldDescriptors{
							fields: []protoreflect.FieldDescriptor{
								&mockFieldDescriptor{name: "test_field", kind: tc.fieldKind},
							},
						},
					},
				}

				properties, err := MethodDescriptorToProtoProperties(mockMethod)
				require.NoError(t, err)
				require.Len(t, properties.Fields, 1)

				field, ok := properties.Fields["test_field"]
				require.True(t, ok)
				s := field.GetStructValue()
				require.NotNil(t, s)
				assert.Equal(t, tc.expectedType, s.Fields["type"].GetStringValue())
			})
		}
	})

	t.Run("missing types", func(t *testing.T) {
		t.Run("bytes field", func(t *testing.T) {
			mockMethod := &mockMethodDescriptor{
				input: &mockMessageDescriptor{
					fields: &mockFieldDescriptors{
						fields: []protoreflect.FieldDescriptor{
							&mockFieldDescriptor{name: "bytes_field", kind: protoreflect.BytesKind},
						},
					},
				},
			}

			properties, err := MethodDescriptorToProtoProperties(mockMethod)
			require.NoError(t, err)
			require.Len(t, properties.Fields, 1)

			field, ok := properties.Fields["bytes_field"]
			require.True(t, ok)
			s := field.GetStructValue()
			require.NotNil(t, s)
			// Explicitly check for string, as bytes are usually represented as base64 strings in JSON
			assert.Equal(t, "string", s.Fields["type"].GetStringValue())
		})

		t.Run("enum field", func(t *testing.T) {
			mockMethod := &mockMethodDescriptor{
				input: &mockMessageDescriptor{
					fields: &mockFieldDescriptors{
						fields: []protoreflect.FieldDescriptor{
							&mockFieldDescriptor{name: "enum_field", kind: protoreflect.EnumKind},
						},
					},
				},
			}

			properties, err := MethodDescriptorToProtoProperties(mockMethod)
			require.NoError(t, err)
			require.Len(t, properties.Fields, 1)

			field, ok := properties.Fields["enum_field"]
			require.True(t, ok)
			s := field.GetStructValue()
			require.NotNil(t, s)
			// Enums are typically represented as strings (names) or integers. Defaulting to string (names) is good practice for JSON-RPC.
			assert.Equal(t, "string", s.Fields["type"].GetStringValue())
		})

		t.Run("message field (nested object)", func(t *testing.T) {
			mockMethod := &mockMethodDescriptor{
				input: &mockMessageDescriptor{
					fields: &mockFieldDescriptors{
						fields: []protoreflect.FieldDescriptor{
							&mockFieldDescriptor{name: "message_field", kind: protoreflect.MessageKind},
						},
					},
				},
			}

			properties, err := MethodDescriptorToProtoProperties(mockMethod)
			require.NoError(t, err)
			require.Len(t, properties.Fields, 1)

			field, ok := properties.Fields["message_field"]
			require.True(t, ok)
			s := field.GetStructValue()
			require.NotNil(t, s)

			assert.Equal(t, "object", s.Fields["type"].GetStringValue())
		})
	})
}

func TestMethodOutputDescriptorToProtoProperties(t *testing.T) {
	t.Run("with real proto", func(t *testing.T) {
		fileDesc := weatherv1.File_proto_examples_weather_v1_weather_proto
		require.NotNil(t, fileDesc)

		serviceDesc := fileDesc.Services().ByName("WeatherService")
		require.NotNil(t, serviceDesc)

		methodDesc := serviceDesc.Methods().ByName("GetWeather")
		require.NotNil(t, methodDesc)

		properties, err := MethodOutputDescriptorToProtoProperties(methodDesc)
		require.NoError(t, err)
		require.Len(t, properties.Fields, 1)

		weatherField, ok := properties.Fields["weather"]
		require.True(t, ok)
		s1 := weatherField.GetStructValue()
		require.NotNil(t, s1)
		assert.Equal(t, "string", s1.Fields["type"].GetStringValue())
	})

	t.Run("with mocks", func(t *testing.T) {
		testCases := []struct {
			name         string
			fieldKind    protoreflect.Kind
			expectedType string
		}{
			{"string", protoreflect.StringKind, "string"},
			{"double", protoreflect.DoubleKind, "number"},
			{"float", protoreflect.FloatKind, "number"},
			{"int32", protoreflect.Int32Kind, "integer"},
			{"bool", protoreflect.BoolKind, "boolean"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockMethod := &mockMethodDescriptor{
					output: &mockMessageDescriptor{
						fields: &mockFieldDescriptors{
							fields: []protoreflect.FieldDescriptor{
								&mockFieldDescriptor{name: "test_field", kind: tc.fieldKind},
							},
						},
					},
				}

				properties, err := MethodOutputDescriptorToProtoProperties(mockMethod)
				require.NoError(t, err)
				require.Len(t, properties.Fields, 1)

				field, ok := properties.Fields["test_field"]
				require.True(t, ok)
				s := field.GetStructValue()
				require.NotNil(t, s)
				assert.Equal(t, tc.expectedType, s.Fields["type"].GetStringValue())
			})
		}
	})

	t.Run("missing types", func(t *testing.T) {
		t.Run("message field (nested object)", func(t *testing.T) {
			mockMethod := &mockMethodDescriptor{
				output: &mockMessageDescriptor{
					fields: &mockFieldDescriptors{
						fields: []protoreflect.FieldDescriptor{
							&mockFieldDescriptor{name: "message_field", kind: protoreflect.MessageKind},
						},
					},
				},
			}

			properties, err := MethodOutputDescriptorToProtoProperties(mockMethod)
			require.NoError(t, err)
			require.Len(t, properties.Fields, 1)

			field, ok := properties.Fields["message_field"]
			require.True(t, ok)
			s := field.GetStructValue()
			require.NotNil(t, s)

			assert.Equal(t, "object", s.Fields["type"].GetStringValue())
		})
	})
}
