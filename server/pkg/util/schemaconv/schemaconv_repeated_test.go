// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// mockFieldDescriptorWithList is a mock implementation of protoreflect.FieldDescriptor for testing.
type mockFieldDescriptorWithList struct {
	protoreflect.FieldDescriptor
	kind   protoreflect.Kind
	name   string
	isList bool
}

func (m *mockFieldDescriptorWithList) Kind() protoreflect.Kind {
	return m.kind
}

func (m *mockFieldDescriptorWithList) Name() protoreflect.Name {
	return protoreflect.Name(m.name)
}

func (m *mockFieldDescriptorWithList) IsList() bool {
	return m.isList
}

func (m *mockFieldDescriptorWithList) IsMap() bool {
	return false
}

func TestMethodDescriptorToProtoProperties_RepeatedField(t *testing.T) {
	mockMethod := &mockMethodDescriptor{
		input: &mockMessageDescriptor{
			fields: &mockFieldDescriptors{
				fields: []protoreflect.FieldDescriptor{
					&mockFieldDescriptorWithList{name: "tags", kind: protoreflect.StringKind, isList: true},
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

	// Expectation: type should be array, and items should have type string
	assert.Equal(t, "array", s.Fields["type"].GetStringValue())

	items := s.Fields["items"].GetStructValue()
	require.NotNil(t, items)
	assert.Equal(t, "string", items.Fields["type"].GetStringValue())
}
