/*
 * Copyright 2025 Author(s) of MCPX
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

package tool

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/mcpxy/mcpx/pkg/upstream/grpc/protobufparser"
	pb "github.com/mcpxy/mcpx/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvertMCPToolToProto(t *testing.T) {
	// t.Parallel() // Removed to debug potential race conditions
	toolName := "test-tool"
	displayName := "Test Tool"
	description := "A tool for testing"
	toolType := "object"

	inputSchema := &jsonschema.Schema{
		Title: displayName,
		Type:  toolType,
		Properties: map[string]*jsonschema.Schema{
			"arg1": {
				Type:        "string",
				Description: "Argument 1",
			},
		},
		Required: []string{"arg1"},
	}
	jsonBytes, err := json.Marshal(inputSchema)
	if err != nil {
		t.Fatalf("failed to marshal input schema to json: %v", err)
	}
	propertiesStruct := &structpb.Struct{}
	if err := protojson.Unmarshal(jsonBytes, propertiesStruct); err != nil {
		t.Fatalf("failed to unmarshal input schema from json: %v", err)
	}

	mcpTool := &mcp.Tool{
		Name:        toolName,
		Description: description,
		InputSchema: inputSchema,
	}
	protoTool, err := convertMCPToolToProto(mcpTool)
	if err != nil {
		t.Fatalf("convertMCPToolToProto() failed: %v", err)
	}

	expectedInputSchema := pb.InputSchema_builder{
		Type:       &toolType,
		Properties: propertiesStruct,
		Required:   []string{"arg1"},
	}.Build()

	expectedTool := pb.Tool_builder{
		Name:        &toolName,
		DisplayName: &displayName,
		Description: &description,
		InputSchema: expectedInputSchema,
	}.Build()

	if diff := cmp.Diff(expectedTool, protoTool, protocmp.Transform()); diff != "" {
		t.Errorf("convertMCPToolToProto() returned diff (-want +got):\n%s", diff)
	}
}

func TestConvertMcpFieldsToInputSchemaProperties(t *testing.T) {
	// t.Parallel() // Removed to debug potential race conditions
	fields := []*protobufparser.McpField{
		{
			Name:        "field1",
			Type:        "TYPE_STRING",
			Description: "string field",
		},
		{
			Name:        "field2",
			Type:        "TYPE_INT32",
			Description: "int32 field",
		},
	}

	properties, err := convertMcpFieldsToInputSchemaProperties(fields)
	if err != nil {
		t.Fatalf("convertMcpFieldsToInputSchemaProperties() failed: %v", err)
	}
	expectedProperties := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"field1": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type":        structpb.NewStringValue("string"),
					"description": structpb.NewStringValue("string field"),
				},
			}),
			"field2": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type":        structpb.NewStringValue("integer"),
					"description": structpb.NewStringValue("int32 field"),
				},
			}),
		},
	}
	if diff := cmp.Diff(expectedProperties, properties, protocmp.Transform()); diff != "" {
		t.Errorf("convertMcpFieldsToInputSchemaProperties() returned diff (-want +got):\n%s", diff)
	}
}

func TestGetJSONSchemaForScalarType(t *testing.T) {
	// t.Parallel() // Removed to debug potential race conditions
	testCases := []struct {
		name           string
		scalarType     string
		description    string
		expectedSchema *jsonschema.Schema
		expectedError  error
	}{
		{
			name:        "string type",
			scalarType:  "TYPE_STRING",
			description: "a string",
			expectedSchema: &jsonschema.Schema{
				Type:        "string",
				Description: "a string",
			},
		},
		{
			name:        "integer type",
			scalarType:  "TYPE_INT32",
			description: "an integer",
			expectedSchema: &jsonschema.Schema{
				Type:        "integer",
				Description: "an integer",
			},
		},
		{
			name:        "number type",
			scalarType:  "TYPE_FLOAT",
			description: "a float",
			expectedSchema: &jsonschema.Schema{
				Type:        "number",
				Description: "a float",
			},
		},
		{
			name:        "boolean type",
			scalarType:  "TYPE_BOOL",
			description: "a boolean",
			expectedSchema: &jsonschema.Schema{
				Type:        "boolean",
				Description: "a boolean",
			},
		},
		{
			name:          "unsupported type",
			scalarType:    "TYPE_MESSAGE",
			description:   "a message",
			expectedError: fmt.Errorf("unsupported scalar type: message"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() // Removed to debug potential race conditions
			schema, err := getJSONSchemaForScalarType(tc.scalarType, tc.description)
			if (err != nil) != (tc.expectedError != nil) || (err != nil && err.Error() != tc.expectedError.Error()) {
				t.Fatalf("getJSONSchemaForScalarType() error = %v, wantErr %v", err, tc.expectedError)
			}

			if diff := cmp.Diff(tc.expectedSchema, schema); diff != "" {
				t.Errorf("getJSONSchemaForScalarType() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
