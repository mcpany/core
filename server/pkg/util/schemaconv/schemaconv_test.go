// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	weatherv1 "github.com/mcpany/core/proto/examples/weather/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
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
	isRepeated  bool
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

func (m *mockMcpFieldParameter) GetIsRepeated() bool {
	return m.isRepeated
}

func TestMcpFieldsToProtoProperties_Repeated(t *testing.T) {
	params := []*mockMcpFieldParameter{
		{
			name:        "repeated_string",
			description: "list of strings",
			typ:         "TYPE_STRING",
			isRepeated:  true,
		},
	}

	properties, err := McpFieldsToProtoProperties(params)
	require.NoError(t, err)
	assert.Len(t, properties.Fields, 1)

	field, ok := properties.Fields["repeated_string"]
	require.True(t, ok)
	s := field.GetStructValue()
	require.NotNil(t, s)

	// Expect type "array"
	assert.Equal(t, "array", s.Fields["type"].GetStringValue())

	// Expect items to be "string"
	items := s.Fields["items"].GetStructValue()
	require.NotNil(t, items)
	assert.Equal(t, "string", items.Fields["type"].GetStringValue())
}

func TestConfigSchemaToProtoProperties(t *testing.T) {
	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])
	intType := configv1.ParameterType(configv1.ParameterType_value["INTEGER"])
	params := []*mockConfigParameter{
		{schema: configv1.ParameterSchema_builder{Name: proto.String("param1"), Description: proto.String("a string param"), Type: &stringType}.Build()},
		{schema: configv1.ParameterSchema_builder{Name: proto.String("param2"), Description: proto.String("an int param"), Type: &intType}.Build()},
		{schema: nil},
	}

	properties, _, err := ConfigSchemaToProtoProperties(params)
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
	kind        protoreflect.Kind
	name        string
	cardinality protoreflect.Cardinality
	message     protoreflect.MessageDescriptor
	isMap       bool
	mapKey      protoreflect.FieldDescriptor
	mapValue    protoreflect.FieldDescriptor
	enumDesc    protoreflect.EnumDescriptor
}

func (m *mockFieldDescriptor) Kind() protoreflect.Kind {
	return m.kind
}

func (m *mockFieldDescriptor) Message() protoreflect.MessageDescriptor {
	return m.message
}

func (m *mockFieldDescriptor) Enum() protoreflect.EnumDescriptor {
	return m.enumDesc
}

func (m *mockFieldDescriptor) Name() protoreflect.Name {
	return protoreflect.Name(m.name)
}

func (m *mockFieldDescriptor) Cardinality() protoreflect.Cardinality {
	return m.cardinality
}

func (m *mockFieldDescriptor) IsList() bool {
	// Maps are repeated on wire but IsList() usually returns false for maps in higher level abstractions
	// but protoreflect says IsList() returns true if Cardinality is Repeated.
	// HOWEVER, for Map fields, IsMap() is true, and IsList() might be false depending on implementation?
	// The protoreflect docs say:
	// IsList reports whether this field is a list.
	// If IsList is true, then Cardinality is Repeated.
	// If IsMap is true, then IsList is false.
	if m.isMap {
		return false
	}
	return m.cardinality == protoreflect.Repeated
}

func (m *mockFieldDescriptor) IsMap() bool {
	return m.isMap
}

func (m *mockFieldDescriptor) MapKey() protoreflect.FieldDescriptor {
	return m.mapKey
}

func (m *mockFieldDescriptor) MapValue() protoreflect.FieldDescriptor {
	return m.mapValue
}

// mockEnumDescriptor is a mock implementation of protoreflect.EnumDescriptor for testing.
type mockEnumDescriptor struct {
	protoreflect.EnumDescriptor
	values protoreflect.EnumValueDescriptors
}

func (m *mockEnumDescriptor) Values() protoreflect.EnumValueDescriptors {
	return m.values
}

// mockEnumValueDescriptors is a mock implementation of protoreflect.EnumValueDescriptors for testing.
type mockEnumValueDescriptors struct {
	protoreflect.EnumValueDescriptors
	values []protoreflect.EnumValueDescriptor
}

func (m *mockEnumValueDescriptors) Len() int {
	return len(m.values)
}

func (m *mockEnumValueDescriptors) Get(i int) protoreflect.EnumValueDescriptor {
	return m.values[i]
}

// mockEnumValueDescriptor is a mock implementation of protoreflect.EnumValueDescriptor for testing.
type mockEnumValueDescriptor struct {
	protoreflect.EnumValueDescriptor
	name string
}

func (m *mockEnumValueDescriptor) Name() protoreflect.Name {
	return protoreflect.Name(m.name)
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

	t.Run("with repeated fields", func(t *testing.T) {
		// Verify that repeated fields are converted to array type with correct item type
		mockMethod := &mockMethodDescriptor{
			output: &mockMessageDescriptor{
				fields: &mockFieldDescriptors{
					fields: []protoreflect.FieldDescriptor{
						&mockFieldDescriptor{
							name:        "tags",
							kind:        protoreflect.StringKind,
							cardinality: protoreflect.Repeated,
						},
					},
				},
			},
		}

		properties, err := MethodOutputDescriptorToProtoProperties(mockMethod)
		require.NoError(t, err)
		require.Len(t, properties.Fields, 1)

		tagsField, ok := properties.Fields["tags"]
		require.True(t, ok)
		s := tagsField.GetStructValue()
		require.NotNil(t, s)

		assert.Equal(t, "array", s.Fields["type"].GetStringValue())
		items := s.Fields["items"].GetStructValue()
		require.NotNil(t, items)
		assert.Equal(t, "string", items.Fields["type"].GetStringValue())
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
}

func TestMethodDescriptorToProtoProperties_MessageKind(t *testing.T) {
	// Nested message descriptor
	nestedMsg := &mockMessageDescriptor{
		fields: &mockFieldDescriptors{
			fields: []protoreflect.FieldDescriptor{
				&mockFieldDescriptor{
					name: "nested_field",
					kind: protoreflect.StringKind,
				},
			},
		},
	}

	// Input message descriptor containing a MessageKind field
	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:    "nested_msg",
						kind:    protoreflect.MessageKind,
						message: nestedMsg,
					},
				},
			},
		},
	}

	properties, err := MethodDescriptorToProtoProperties(mockMethod)
	require.NoError(t, err)
	field := properties.Fields["nested_msg"]
	require.NotNil(t, field)
	s := field.GetStructValue()

	// Expect "object" type for MessageKind
	assert.Equal(t, "object", s.Fields["type"].GetStringValue(), "Expected MessageKind to be converted to object")

	// Expect nested properties
	nestedProps := s.Fields["properties"].GetStructValue()
	require.NotNil(t, nestedProps)
	assert.Contains(t, nestedProps.Fields, "nested_field")
}

func TestMethodDescriptorToProtoProperties_Enum(t *testing.T) {
	// Create an enum descriptor with values
	enumDesc := &mockEnumDescriptor{
		values: &mockEnumValueDescriptors{
			values: []protoreflect.EnumValueDescriptor{
				&mockEnumValueDescriptor{name: "VAL_A"},
				&mockEnumValueDescriptor{name: "VAL_B"},
				&mockEnumValueDescriptor{name: "VAL_C"},
			},
		},
	}

	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:     "enum_field",
						kind:     protoreflect.EnumKind,
						enumDesc: enumDesc,
					},
				},
			},
		},
	}

	properties, err := MethodDescriptorToProtoProperties(mockMethod)
	require.NoError(t, err)

	field, ok := properties.Fields["enum_field"]
	require.True(t, ok)
	s := field.GetStructValue()
	require.NotNil(t, s)

	// Expect "string" type for Enum
	assert.Equal(t, "string", s.Fields["type"].GetStringValue())

	// Expect "enum" property to list values
	enumVal, ok := s.Fields["enum"]
	require.True(t, ok, "enum field missing")

	enumList := enumVal.GetListValue()
	require.NotNil(t, enumList)
	assert.Len(t, enumList.Values, 3)

	vals := []string{}
	for _, v := range enumList.Values {
		vals = append(vals, v.GetStringValue())
	}
	assert.ElementsMatch(t, []string{"VAL_A", "VAL_B", "VAL_C"}, vals)
}

func TestMethodDescriptorToProtoProperties_Repeated(t *testing.T) {
	// This test covers the missing test case due to copy-paste error in the original test file
	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:        "tags",
						kind:        protoreflect.StringKind,
						cardinality: protoreflect.Repeated,
					},
				},
			},
		},
	}

	properties, err := MethodDescriptorToProtoProperties(mockMethod)
	require.NoError(t, err)
	require.Len(t, properties.Fields, 1)

	tagsField, ok := properties.Fields["tags"]
	require.True(t, ok)
	s := tagsField.GetStructValue()
	require.NotNil(t, s)

	assert.Equal(t, "array", s.Fields["type"].GetStringValue())
	items := s.Fields["items"].GetStructValue()
	require.NotNil(t, items)
	assert.Equal(t, "string", items.Fields["type"].GetStringValue())
}

func TestFieldsToProperties_RecursionLimit(t *testing.T) {
	// Create a recursive message structure
	// recursiveMsg -> field "next" (MessageKind) -> recursiveMsg
	recursiveMsg := &mockMessageDescriptor{}
	fields := &mockFieldDescriptors{
		fields: []protoreflect.FieldDescriptor{
			&mockFieldDescriptor{
				name:    "next",
				kind:    protoreflect.MessageKind,
				message: recursiveMsg,
			},
		},
	}
	recursiveMsg.fields = fields

	mockMethod := &mockMethodDescriptor{
		input: recursiveMsg,
	}

	// This should fail due to recursion limit
	_, err := MethodDescriptorToProtoProperties(mockMethod)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "recursion depth limit reached"))
}

func TestFieldsToProperties_Map(t *testing.T) {
	// Map field: map<string, int32> labels = 1;
	// This should be converted to an object with additionalProperties of type integer

	// MapEntry message (simulated)
	mapEntryMsg := &mockMessageDescriptor{
		fields: &mockFieldDescriptors{
			fields: []protoreflect.FieldDescriptor{
				&mockFieldDescriptor{
					name: "key",
					kind: protoreflect.StringKind,
				},
				&mockFieldDescriptor{
					name: "value",
					kind: protoreflect.Int32Kind,
				},
			},
		},
	}

	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:        "labels",
						kind:        protoreflect.MessageKind,
						cardinality: protoreflect.Repeated,
						isMap:       true,
						message:     mapEntryMsg,
						mapValue:    &mockFieldDescriptor{kind: protoreflect.Int32Kind},
					},
				},
			},
		},
	}

	properties, err := MethodDescriptorToProtoProperties(mockMethod)
	require.NoError(t, err)

	labelsField, ok := properties.Fields["labels"]
	require.True(t, ok)
	s := labelsField.GetStructValue()
	require.NotNil(t, s)

	// Now we expect correct behavior
	assert.Equal(t, "object", s.Fields["type"].GetStringValue())

	additionalProperties := s.Fields["additionalProperties"].GetStructValue()
	require.NotNil(t, additionalProperties)
	assert.Equal(t, "integer", additionalProperties.Fields["type"].GetStringValue())

	// Should NOT have properties "key" and "value"
	assert.NotContains(t, s.Fields, "properties")
}

func TestConfigSchemaToProtoProperties_DefaultValue(t *testing.T) {
	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])
	defaultValue, err := structpb.NewValue("default-value")
	require.NoError(t, err)

	params := []*mockConfigParameter{
		{
			schema: configv1.ParameterSchema_builder{
				Name:         proto.String("param_with_default"),
				Description:  proto.String("param with default value"),
				Type:         &stringType,
				DefaultValue: defaultValue,
			}.Build(),
		},
	}

	properties, _, err := ConfigSchemaToProtoProperties(params)
	require.NoError(t, err)

	param, ok := properties.Fields["param_with_default"]
	require.True(t, ok)
	s := param.GetStructValue()
	require.NotNil(t, s)

	defVal, ok := s.Fields["default"]
	require.True(t, ok, "default field missing from schema properties")
	assert.Equal(t, "default-value", defVal.GetStringValue())
}

func TestConfigSchemaToProtoProperties_InvalidType(t *testing.T) {
	// 999 is likely an invalid type
	invalidType := configv1.ParameterType(999)
	params := []*mockConfigParameter{
		{
			schema: configv1.ParameterSchema_builder{
				Name:        proto.String("param_invalid"),
				Description: proto.String("param with invalid type"),
				Type:        &invalidType,
			}.Build(),
		},
	}

	properties, _, err := ConfigSchemaToProtoProperties(params)
	require.NoError(t, err)

	param, ok := properties.Fields["param_invalid"]
	require.True(t, ok)
	s := param.GetStructValue()
	require.NotNil(t, s)

	// It should default to "string" instead of "" (empty string)
	assert.Equal(t, "string", s.Fields["type"].GetStringValue())
}

func TestFieldsToProperties_MapRecursionLimit(t *testing.T) {
	// Trigger error at line 56: fieldToSchema fails because of recursion depth in map value

	// Create a recursive message structure
	recursiveMsg := &mockMessageDescriptor{}
	fields := &mockFieldDescriptors{
		fields: []protoreflect.FieldDescriptor{
			&mockFieldDescriptor{
				name:    "next",
				kind:    protoreflect.MessageKind,
				message: recursiveMsg,
			},
		},
	}
	recursiveMsg.fields = fields

	// Map field where value is recursiveMsg
	mapEntryMsg := &mockMessageDescriptor{
		fields: &mockFieldDescriptors{
			fields: []protoreflect.FieldDescriptor{
				&mockFieldDescriptor{
					name: "key",
					kind: protoreflect.StringKind,
				},
				&mockFieldDescriptor{
					name: "value",
					kind: protoreflect.MessageKind,
					message: recursiveMsg,
				},
			},
		},
	}

	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:        "labels",
						kind:        protoreflect.MessageKind,
						cardinality: protoreflect.Repeated,
						isMap:       true,
						message:     mapEntryMsg,
						mapValue:    &mockFieldDescriptor{kind: protoreflect.MessageKind, message: recursiveMsg},
					},
				},
			},
		},
	}

	// This should fail due to recursion limit inside map value processing
	// We need to set initial depth close to limit so that fieldToSchema(mapValue) -> fieldsToProperties(recursiveMsg) fails
	// But fieldToSchema takes depth. fieldsToProperties passes depth.
	// If we start at MaxRecursionDepth, fieldsToProperties returns error immediately.
	// We want line 56 to fail.
	// fieldToSchema calls fieldsToProperties with depth+1 if it's a message.

	// If we pass depth = MaxRecursionDepth - 1
	// 1. fieldsToProperties(depth)
	// 2. map field processing
	// 3. fieldToSchema(mapValue, depth)
	// 4. mapValue is MessageKind -> fieldsToProperties(recursiveMsg fields, depth + 1)
	// 5. depth+1 = MaxRecursionDepth + 1 > MaxRecursionDepth -> Error!

	// So we need to call fieldsToProperties with MaxRecursionDepth.
	// But first check in fieldsToProperties is recursion limit.

	// Wait, max recursion depth check is: if depth > MaxRecursionDepth
	// So if depth == MaxRecursionDepth, it proceeds.

	// Let's see:
	// fieldsToProperties(MaxRecursionDepth)
	// -> fieldToSchema(mapValue, MaxRecursionDepth)
	// -> fieldsToProperties(recursiveMsg, MaxRecursionDepth + 1)
	// -> Error!

	// So calling MethodDescriptorToProtoProperties (depth 0) won't trigger it directly unless the schema is deeply nested.
	// We can call fieldsToProperties directly if we export it or use a public wrapper.
	// MethodDescriptorToProtoProperties calls it with 0.

	// We can make the recursive message actually deep enough.
	// Or we can just trust that we covered the logic by analysis, but we want code coverage.
	// The existing TestFieldsToProperties_RecursionLimit covers the "if depth > Max" check at the top.
	// It doesn't cover the error return at line 56.

	// To cover line 56, we need fieldToSchema to return error.
	// fieldToSchema returns error if fieldsToProperties returns error (for nested messages).

	// So we construct a schema where map value is a message, and that message causes recursion error.
	// Since we can't control 'depth' passed to MethodDescriptorToProtoProperties (it's always 0),
	// we must construct a schema that is DEEP enough.
	// But MaxRecursionDepth is 10.
	// We can't easily construct a 10-level deep mock without a loop or lots of code.
	// BUT, we can use the self-referential recursiveMsg!

	// recursiveMsg refers to itself.
	// So traversing it will go infinitely deep until it hits limit.
	// If we put this recursiveMsg as the value of a Map,
	// fieldsToProperties(mapField) -> fieldToSchema(mapValue=recursiveMsg) -> fieldsToProperties(recursiveMsg) -> ... -> Error

	// The error will propagate up:
	// ... -> fieldsToProperties (returns "recursion limit reached")
	// -> fieldToSchema (returns error)
	// -> fieldsToProperties (line 56 catches error and wraps it)

	_, err := MethodDescriptorToProtoProperties(mockMethod)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process map value")
	assert.Contains(t, err.Error(), "recursion depth limit reached")
}
