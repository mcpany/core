// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// mockFieldDescriptorWithCardinality extends mockFieldDescriptor to support Cardinality
type mockFieldDescriptorWithCardinality struct {
	mockFieldDescriptor
	cardinality protoreflect.Cardinality
}

func (m *mockFieldDescriptorWithCardinality) Cardinality() protoreflect.Cardinality {
	return m.cardinality
}

func (m *mockFieldDescriptorWithCardinality) IsList() bool {
	return m.cardinality == protoreflect.Repeated
}

func TestMethodDescriptorToProtoProperties_RepeatedField(t *testing.T) {
	// Create a mock method descriptor with a repeated string field
	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptorWithCardinality{
						mockFieldDescriptor: mockFieldDescriptor{
							name: "tags",
							kind: protoreflect.StringKind,
						},
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

	// This assertion is expected to fail if the bug exists
	// The current implementation probably returns "string" instead of "array"
	assert.Equal(t, "array", s.Fields["type"].GetStringValue(), "Repeated field should have type 'array'")

	items := s.Fields["items"].GetStructValue()
	require.NotNil(t, items, "Array field should have 'items' property")
	assert.Equal(t, "string", items.Fields["type"].GetStringValue(), "Items type should be 'string'")
}

func TestMethodOutputDescriptorToProtoProperties_RepeatedField(t *testing.T) {
	// Create a mock method descriptor with a repeated string field in output
	mockMethod := &mockMethodDescriptor{
		output: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptorWithCardinality{
						mockFieldDescriptor: mockFieldDescriptor{
							name: "results",
							kind: protoreflect.StringKind,
						},
						cardinality: protoreflect.Repeated,
					},
				},
			},
		},
	}

	properties, err := MethodOutputDescriptorToProtoProperties(mockMethod)
	require.NoError(t, err)
	require.Len(t, properties.Fields, 1)

	resultsField, ok := properties.Fields["results"]
	require.True(t, ok)
	s := resultsField.GetStructValue()
	require.NotNil(t, s)

	// This assertion is expected to fail if the bug exists
	assert.Equal(t, "array", s.Fields["type"].GetStringValue(), "Repeated field should have type 'array'")

	items := s.Fields["items"].GetStructValue()
	require.NotNil(t, items, "Array field should have 'items' property")
	assert.Equal(t, "string", items.Fields["type"].GetStringValue(), "Items type should be 'string'")
}
