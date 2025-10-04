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

func convertMCPToolToProto(tool *mcp.Tool) (*pb.Tool, error) {
	if tool == nil {
		return nil, fmt.Errorf("cannot convert nil mcp tool to proto")
	}
	if tool.InputSchema == nil {
		// If there's no input schema, we can just return the tool with the name and description.
		name := tool.Name
		description := tool.Description
		return pb.Tool_builder{
			Name:        &name,
			Description: &description,
		}.Build(), nil
	}

	// Marshal the provided input schema to JSON. This is a safe way to handle different
	// underlying types (e.g., map[string]any, *jsonschema.Schema).
	jsonBytes, err := json.Marshal(tool.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input schema to json: %w", err)
	}

	// Unmarshal into a generic structpb.Struct for the properties.
	properties := &structpb.Struct{}
	if err := protojson.Unmarshal(jsonBytes, properties); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input schema from json to structpb: %w", err)
	}

	// Unmarshal into a temporary struct to safely extract top-level schema fields.
	var tempSchema struct {
		Type     string   `json:"type,omitempty"`
		Title    string   `json:"title,omitempty"`
		Required []string `json:"required,omitempty"`
	}
	if err := json.Unmarshal(jsonBytes, &tempSchema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input schema from json to temporary struct: %w", err)
	}

	name := tool.Name
	displayName := tempSchema.Title
	description := tool.Description

	pbInputSchema := pb.InputSchema_builder{
		Type:       &tempSchema.Type,
		Properties: properties,
		Required:   tempSchema.Required,
	}.Build()

	return pb.Tool_builder{
		Name:        &name,
		DisplayName: &displayName,
		Description: &description,
		InputSchema: pbInputSchema,
	}.Build(), nil
}

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
