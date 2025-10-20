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
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/mcpxy/core/pkg/upstream/grpc/protobufparser"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// convertMCPToolToProto transforms an *mcp.Tool, which uses a flexible schema
// representation, into a protobuf-defined *pb.Tool with a structured input
// schema. This is used to standardize tool definitions within the system.
func convertMCPToolToProto(tool *mcp.Tool) (*pb.Tool, error) {
	if tool == nil {
		return nil, fmt.Errorf("cannot convert nil mcp tool to proto")
	}

	pbTool := &pb.Tool{}
	pbTool.Reset()
	pbTool.SetName(tool.Name)
	pbTool.SetDescription(tool.Description)

	// Set display name based on precedence: Title, Annotations.Title, then Name
	if tool.Title != "" {
		pbTool.SetDisplayName(tool.Title)
	} else if tool.Annotations != nil && tool.Annotations.Title != "" {
		pbTool.SetDisplayName(tool.Annotations.Title)
	} else {
		pbTool.SetDisplayName(tool.Name)
	}

	if tool.InputSchema != nil {
		inputSchema, err := convertJSONSchemaToProto[pb.InputSchema](tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input schema: %w", err)
		}
		pbTool.SetInputSchema(inputSchema)
	}

	if tool.OutputSchema != nil {
		outputSchema, err := convertJSONSchemaToProto[pb.OutputSchema](tool.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert output schema: %w", err)
		}
		pbTool.SetOutputSchema(outputSchema)
	}

	if tool.Annotations != nil {
		pbAnnotations := &pb.ToolAnnotations{}
		pbAnnotations.Reset()
		pbAnnotations.SetTitle(tool.Annotations.Title)
		pbAnnotations.SetReadOnlyHint(tool.Annotations.ReadOnlyHint)
		pbAnnotations.SetIdempotentHint(tool.Annotations.IdempotentHint)
		if tool.Annotations.DestructiveHint != nil {
			pbAnnotations.SetDestructiveHint(*tool.Annotations.DestructiveHint)
		}
		if tool.Annotations.OpenWorldHint != nil {
			pbAnnotations.SetOpenWorldHint(*tool.Annotations.OpenWorldHint)
		}
		pbTool.SetAnnotations(pbAnnotations)
	}

	return pbTool, nil
}

// convertJSONSchemaToProto is a generic function that converts a JSON schema,
// represented as an `any` type, into a protobuf message that has `Type` and
// `Properties` fields. This is used for both input and output schemas.
func convertJSONSchemaToProto[T any, PT interface {
	*T
	Reset()
	ProtoMessage()
	SetType(string)
	SetProperties(*structpb.Struct)
}](schema any) (PT, error) {
	if schema == nil {
		return nil, nil
	}

	// Marshal the provided schema to JSON.
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema to json: %w", err)
	}

	// Unmarshal into a temporary struct to safely extract top-level schema fields.
	var tempSchema struct {
		Type       string          `json:"type,omitempty"`
		Properties json.RawMessage `json:"properties,omitempty"`
	}
	if err := json.Unmarshal(jsonBytes, &tempSchema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema from json to temporary struct: %w", err)
	}

	// Unmarshal the properties field into a structpb.Struct.
	properties := &structpb.Struct{}
	if len(tempSchema.Properties) > 0 {
		if err := protojson.Unmarshal(tempSchema.Properties, properties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal properties from json to structpb: %w", err)
		}
	}

	pbSchema := PT(new(T))
	pbSchema.Reset()
	pbSchema.SetType(tempSchema.Type)
	pbSchema.SetProperties(properties)

	return pbSchema, nil
}

// convertMcpFieldsToInputSchemaProperties converts a slice of McpField, which
// represent fields from a protobuf message, into a structpb.Struct that can be
// used as the `properties` field in a JSON schema.
func convertMcpFieldsToInputSchemaProperties(fields []*protobufparser.McpField) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	for _, field := range fields {
		schema, err := getJSONSchemaForScalarType(field.Type, field.Description)
		if err != nil {
			return nil, err
		}

		jsonBytes, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema to json: %w", err)
		}

		value := &structpb.Value{}
		if err := protojson.Unmarshal(jsonBytes, value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema from json: %w", err)
		}

		properties.Fields[field.Name] = value
	}
	return properties, nil
}

// getJSONSchemaForScalarType maps a protobuf scalar type (e.g., "TYPE_STRING",
// "TYPE_INT32") to its corresponding JSON schema type ("string", "integer"). It
// is a helper function for building JSON schemas from protobuf definitions.
func getJSONSchemaForScalarType(scalarType, description string) (*jsonschema.Schema, error) {
	s := &jsonschema.Schema{
		Description: description,
	}

	scalarType = strings.TrimPrefix(scalarType, "TYPE_")
	scalarType = strings.ToLower(scalarType)

	switch scalarType {
	case "double", "float":
		s.Type = "number"
	case "int32", "int64", "sint32", "sint64", "uint32", "uint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
		s.Type = "integer"
	case "bool":
		s.Type = "boolean"
	case "string", "bytes":
		s.Type = "string"
	default:
		return nil, fmt.Errorf("unsupported scalar type: %s", scalarType)
	}

	return s, nil
}
