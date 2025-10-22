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

package tool

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/mcpxy/core/pkg/upstream/grpc/protobufparser"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvertMCPToolToProto(t *testing.T) {
	t.Parallel()

	destructiveHint := true
	openWorldHint := true

	mcpTool := &mcp.Tool{
		Name:        "test-tool",
		Description: "A tool for testing",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Test Tool",
			ReadOnlyHint:    true,
			DestructiveHint: &destructiveHint,
			IdempotentHint:  true,
			OpenWorldHint:   &openWorldHint,
		},
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"arg1": map[string]any{
					"type":        "string",
					"description": "Argument 1",
				},
			},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"result": map[string]any{
					"type":        "string",
					"description": "The result",
				},
			},
		},
	}

	protoTool, err := convertMCPToolToProto(mcpTool)
	if err != nil {
		t.Fatalf("convertMCPToolToProto() failed: %v", err)
	}

	// Helper to create a structpb.Struct from a map
	mustNewStruct := func(m map[string]any) *structpb.Struct {
		s, err := structpb.NewStruct(m)
		if err != nil {
			t.Fatalf("Failed to create struct: %v", err)
		}
		return s
	}

	expectedTool := pb.Tool_builder{
		Name:        proto.String("test-tool"),
		Description: proto.String("A tool for testing"),
		DisplayName: proto.String("Test Tool"),
		Annotations: pb.ToolAnnotations_builder{
			Title:           proto.String("Test Tool"),
			ReadOnlyHint:    proto.Bool(true),
			DestructiveHint: proto.Bool(true),
			IdempotentHint:  proto.Bool(true),
			OpenWorldHint:   proto.Bool(true),
		}.Build(),
		InputSchema: pb.InputSchema_builder{
			Type: proto.String("object"),
			Properties: mustNewStruct(map[string]any{
				"arg1": map[string]any{
					"type":        "string",
					"description": "Argument 1",
				},
			}),
		}.Build(),
		OutputSchema: pb.OutputSchema_builder{
			Type: proto.String("object"),
			Properties: mustNewStruct(map[string]any{
				"result": map[string]any{
					"type":        "string",
					"description": "The result",
				},
			}),
		}.Build(),
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

func TestConvertProtoSchemaToJSONSchema(t *testing.T) {
	t.Parallel()

	// Helper to create a structpb.Struct from a map
	mustNewStruct := func(m map[string]any) *structpb.Struct {
		s, err := structpb.NewStruct(m)
		if err != nil {
			t.Fatalf("Failed to create struct: %v", err)
		}
		return s
	}

	// mustMarshal is a helper function that marshals a map to a json.RawMessage, failing the test on error.
	mustMarshal := func(m map[string]any) json.RawMessage {
		bytes, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("Failed to marshal map to JSON: %v", err)
		}
		return bytes
	}

	testCases := []struct {
		name          string
		protoSchema   protoSchema
		expected      json.RawMessage
		expectedError bool
	}{
		{
			name:        "Nil schema",
			protoSchema: nil,
			expected:    mustMarshal(map[string]any{"type": "object"}),
		},
		{
			name: "Empty schema",
			protoSchema: pb.InputSchema_builder{
				Type:       proto.String("object"),
				Properties: mustNewStruct(map[string]any{}),
			}.Build(),
			expected: mustMarshal(map[string]any{"type": "object", "properties": map[string]any{}}),
		},
		{
			name: "Schema with properties",
			protoSchema: pb.InputSchema_builder{
				Type: proto.String("object"),
				Properties: mustNewStruct(map[string]any{
					"arg1": map[string]any{
						"type":        "string",
						"description": "Argument 1",
					},
					"arg2": map[string]any{
						"type": "integer",
					},
				}),
			}.Build(),
			expected: mustMarshal(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"arg1": map[string]any{
						"type":        "string",
						"description": "Argument 1",
					},
					"arg2": map[string]any{
						"type": "integer",
					},
				},
			}),
		},
		{
			name: "Schema with no type",
			protoSchema: pb.InputSchema_builder{
				Properties: mustNewStruct(map[string]any{}),
			}.Build(),
			expected: mustMarshal(map[string]any{"type": "object", "properties": map[string]any{}}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := convertProtoSchemaToJSONSchema(tc.protoSchema)
			if (err != nil) != tc.expectedError {
				t.Fatalf("convertProtoSchemaToJSONSchema() error = %v, wantErr %v", err, tc.expectedError)
			}

			var expectedMap, resultMap map[string]any
			if err := json.Unmarshal(tc.expected, &expectedMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}
			if err := json.Unmarshal(result, &resultMap); err != nil {
				t.Fatalf("Failed to unmarshal result JSON: %v", err)
			}

			if diff := cmp.Diff(expectedMap, resultMap); diff != "" {
				t.Errorf("convertProtoSchemaToJSONSchema() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProtoToJSONSchemaConversion(t *testing.T) {
	t.Parallel()

	httpDef := configv1.HttpCallDefinition_builder{
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("getLocation"),
		}.Build(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name:        proto.String("ip"),
					Description: proto.String("The IP address to get the location for."),
					Type:        configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name:        proto.String("format"),
					Description: proto.String("The format of the response."),
					Type:        configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	properties := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	for _, param := range httpDef.GetParameters() {
		paramSchema := param.GetSchema()
		paramStruct := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":        structpb.NewStringValue("string"),
				"description": structpb.NewStringValue(paramSchema.GetDescription()),
			},
		}
		properties.Fields[paramSchema.GetName()] = structpb.NewStructValue(paramStruct)
	}

	inputSchema := pb.InputSchema_builder{
		Type:       proto.String("object"),
		Properties: properties,
	}.Build()

	// Now, convert this back to JSON
	jsonSchema, err := convertProtoSchemaToJSONSchema(inputSchema)
	require.NoError(t, err)

	// And compare with the expected JSON
	expectedJSON := `{"properties":{"format":{"description":"The format of the response.","type":"string"},"ip":{"description":"The IP address to get the location for.","type":"string"}},"type":"object"}`

	var expectedMap, actualMap map[string]any
	err = json.Unmarshal([]byte(expectedJSON), &expectedMap)
	require.NoError(t, err)
	err = json.Unmarshal(jsonSchema, &actualMap)
	require.NoError(t, err)

	if diff := cmp.Diff(expectedMap, actualMap); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
