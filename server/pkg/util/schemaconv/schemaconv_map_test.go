// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMethodDescriptorToProtoProperties_Map(t *testing.T) {
	// Map<string, int32>
	// Modeled as a repeated message field where the message is the Map Entry.

	// Map Entry Message: key (string), value (int32)
	mapEntryMsg := &mockMessageDescriptor{
		fields: &mockFieldDescriptors{
			fields: []protoreflect.FieldDescriptor{
				&mockFieldDescriptor{
					name:   "key",
					number: 1,
					kind:   protoreflect.StringKind,
				},
				&mockFieldDescriptor{
					name:   "value",
					number: 2,
					kind:   protoreflect.Int32Kind,
				},
			},
		},
	}

	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:        "my_map",
						kind:        protoreflect.MessageKind,
						cardinality: protoreflect.Repeated,
						message:     mapEntryMsg,
						isMap:       true,
					},
				},
			},
		},
	}

	properties, err := MethodDescriptorToProtoProperties(mockMethod)
	require.NoError(t, err)
	require.Len(t, properties.Fields, 1)

	mapField, ok := properties.Fields["my_map"]
	require.True(t, ok)
	s := mapField.GetStructValue()
	require.NotNil(t, s)

	// Expected for Map:
	// type: object
	// additionalProperties: { type: integer }
	// properties: SHOULD NOT EXIST (or be empty/nil)

	assert.Equal(t, "object", s.Fields["type"].GetStringValue())

	// Check additionalProperties
	additionalProps := s.Fields["additionalProperties"]
	if assert.NotNil(t, additionalProps, "additionalProperties should be set for Map fields") {
		apStruct := additionalProps.GetStructValue()
		assert.Equal(t, "integer", apStruct.Fields["type"].GetStringValue())
	}

	// Check that properties is NOT set (or does not contain key/value)
	// The current buggy implementation sets 'properties' to {key:..., value:...}
	props := s.Fields["properties"]
	if props != nil {
		propsStruct := props.GetStructValue()
		assert.NotContains(t, propsStruct.Fields, "key", "Map schema should not expose 'key' field in properties")
		assert.NotContains(t, propsStruct.Fields, "value", "Map schema should not expose 'value' field in properties")
	}
}

func TestFieldToSchema_MapErrors(t *testing.T) {
	t.Run("missing value field", func(t *testing.T) {
		// Map entry with missing value field (tag 2)
		mapEntryMsg := &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:   "key",
						number: 1,
						kind:   protoreflect.StringKind,
					},
				},
			},
		}

		mockMethod := &mockMethodDescriptor{
			input: &mockMessageDescriptor{
				fields: &mockFieldDescriptors{
					fields: []protoreflect.FieldDescriptor{
						&mockFieldDescriptor{
							name:        "my_map",
							kind:        protoreflect.MessageKind,
							cardinality: protoreflect.Repeated,
							message:     mapEntryMsg,
							isMap:       true,
						},
					},
				},
			},
		}

		_, err := MethodDescriptorToProtoProperties(mockMethod)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing value field")
	})

	t.Run("recursion limit in map value", func(t *testing.T) {
		// Map<string, RecursiveMessage>
		// RecursiveMessage -> field "next" (MessageKind) -> RecursiveMessage
		// Depth will eventually exceed limit.

		// We need a recursive structure for the value.
		// Since we can't easily make a self-referencing struct in initialization without forward declaration:
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

		// Map entry
		mapEntryMsg := &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptor{
						name:   "key",
						number: 1,
						kind:   protoreflect.StringKind,
					},
					&mockFieldDescriptor{
						name:    "value",
						number:  2,
						kind:    protoreflect.MessageKind,
						message: recursiveMsg,
					},
				},
			},
		}

		// Field
		mockMethod := &mockMethodDescriptor{
			input: &mockMessageDescriptor{
				fields: &mockFieldDescriptors{
					fields: []protoreflect.FieldDescriptor{
						&mockFieldDescriptor{
							name:        "my_map",
							kind:        protoreflect.MessageKind,
							cardinality: protoreflect.Repeated,
							message:     mapEntryMsg,
							isMap:       true,
						},
					},
				},
			},
		}

		// To trigger limit, we set start depth close to limit?
		// No, fieldToSchema calls recursively.
		// fieldToSchema(my_map, 0) -> fieldToSchema(value, 1) -> fieldsToProperties(recursiveMsg, 2) -> ...
		// It should fail.

		_, err := MethodDescriptorToProtoProperties(mockMethod)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recursion depth limit reached")
	})
}

func TestFieldsToProperties_Error(t *testing.T) {
	// Trigger error from fieldToSchema
	// e.g. recursion depth

	// Or trigger error from structpb.NewStruct
	// But that's hard as we control the map content.

	// Error propagation is tested via TestFieldToSchema_MapErrors above (MethodDescriptorToProtoProperties calls fieldsToProperties calls fieldToSchema).
}
