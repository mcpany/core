/*
 * Copyright 2025 Author(s) of MCPXY
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

package protobufparser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func loadTestFileDescriptorSet(t *testing.T) *descriptorpb.FileDescriptorSet {
	t.Helper()
	// This path is relative to the package directory where the test is run.
	b, err := os.ReadFile("../../../../build/all.protoset")
	require.NoError(t, err, "Failed to read protoset file. Ensure 'make gen-local' has been run.")

	fds := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(b, fds)
	require.NoError(t, err, "Failed to unmarshal protoset file")

	return fds
}

func TestExtractMcpDefinitions(t *testing.T) {
	fds := loadTestFileDescriptorSet(t)

	t.Run("successful extraction", func(t *testing.T) {
		parsedData, err := ExtractMcpDefinitions(fds)
		require.NoError(t, err)
		assert.NotNil(t, parsedData)

		// Basic checks
		assert.NotEmpty(t, parsedData.Tools)

		// Find a specific tool to inspect
		var addTool *McpTool
		for i, tool := range parsedData.Tools {
			if tool.Name == "CalculatorAdd" {
				addTool = &parsedData.Tools[i]
				break
			}
		}

		require.NotNil(t, addTool, "Tool 'CalculatorAdd' should be found")
		assert.Equal(t, "Adds two integers.", addTool.Description)
		assert.Equal(t, "CalculatorService", addTool.ServiceName)
		assert.Equal(t, "Add", addTool.MethodName)
		assert.Equal(t, "/examples.calculator.v1.CalculatorService/Add", addTool.FullMethodName)
		assert.Equal(t, "examples.calculator.v1.AddRequest", addTool.RequestType)
		assert.Equal(t, "examples.calculator.v1.AddResponse", addTool.ResponseType)
		assert.False(t, addTool.IdempotentHint)
		assert.False(t, addTool.DestructiveHint)

		// Check request fields
		require.Len(t, addTool.RequestFields, 2)
		assert.Equal(t, "a", addTool.RequestFields[0].Name)
		assert.Equal(t, "", addTool.RequestFields[0].Description)
		assert.Equal(t, "int32", addTool.RequestFields[0].Type)
		assert.False(t, addTool.RequestFields[0].IsRepeated)
	})

	t.Run("nil fds", func(t *testing.T) {
		_, err := ExtractMcpDefinitions(nil)
		assert.Error(t, err)
	})

	t.Run("corrupted fds", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:       proto.String("invalid.proto"),
					Package:    proto.String("invalid"),
					Dependency: []string{"a", "b"},
				},
				{
					Name: proto.String("a"),
				},
			},
		}
		_, err := ExtractMcpDefinitions(fds)
		assert.Error(t, err, "Should fail with corrupted/incomplete FileDescriptorSet")
	})
}
