/*
 * Copyright 2025 Author(s) of MCP Any
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
	"github.com/mcpany/core/pkg/upstream/grpc/protobufparser"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ConvertMCPToolToProto transforms an *mcp.Tool, which uses a flexible schema
// representation, into a protobuf-defined *pb.Tool with a structured input
// schema. This is used to standardize tool definitions within the system.
func ConvertMCPToolToProto(tool *mcp.Tool) (*pb.Tool, error) {
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

	pbAnnotations := &pb.ToolAnnotations{}
	pbAnnotations.Reset()
	if tool.Annotations != nil {
		pbAnnotations.SetTitle(tool.Annotations.Title)
		pbAnnotations.SetReadOnlyHint(tool.Annotations.ReadOnlyHint)
		pbAnnotations.SetIdempotentHint(tool.Annotations.IdempotentHint)
		if tool.Annotations.DestructiveHint != nil {
			pbAnnotations.SetDestructiveHint(*tool.Annotations.DestructiveHint)
		}
		if tool.Annotations.OpenWorldHint != nil {
			pbAnnotations.SetOpenWorldHint(*tool.Annotations.OpenWorldHint)
		}
	}

	if tool.InputSchema != nil {
		inputSchema, err := convertJSONSchemaToStruct(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input schema: %w", err)
		}
		pbAnnotations.SetInputSchema(inputSchema)
	} else {
		// If InputSchema is nil, create a default empty object schema
		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		})
		if err == nil {
			pbAnnotations.SetInputSchema(inputSchema)
		}
	}

	if tool.OutputSchema != nil {
		outputSchema, err := convertJSONSchemaToStruct(tool.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert output schema: %w", err)
		}
		pbAnnotations.SetOutputSchema(outputSchema)
	}

	pbTool.SetAnnotations(pbAnnotations)

	return pbTool, nil
}

// convertJSONSchemaToStruct converts a JSON schema, represented as an `any` type,
// into a `*structpb.Struct`.
func convertJSONSchemaToStruct(schema any) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
	}

	// First, check if the schema is a map, which is expected for a JSON object.
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema is not a valid JSON object")
	}

	// Marshal the provided schema to JSON.
	jsonBytes, err := json.Marshal(schemaMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema to json: %w", err)
	}

	// Unmarshal the JSON into a structpb.Struct.
	s := &structpb.Struct{}
	if err := protojson.Unmarshal(jsonBytes, s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema from json to structpb: %w", err)
	}

	return s, nil
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

// ConvertToolDefinitionToProto transforms a *configv1.ToolDefinition into a
// *pb.Tool.
func ConvertToolDefinitionToProto(toolDef *configv1.ToolDefinition) (*pb.Tool, error) {
	if toolDef == nil {
		return nil, fmt.Errorf("cannot convert nil tool definition to proto")
	}

	pbTool := &pb.Tool{}
	pbTool.Reset()
	pbTool.SetName(toolDef.GetName())
	pbTool.SetDescription(toolDef.GetDescription())
	pbTool.SetDisplayName(toolDef.GetTitle())
	pbTool.SetServiceId(toolDef.GetServiceId())

	pbAnnotations := &pb.ToolAnnotations{}
	pbAnnotations.Reset()
	pbAnnotations.SetInputSchema(toolDef.GetInputSchema())
	pbAnnotations.SetOutputSchema(toolDef.GetOutputSchema())

	pbTool.SetAnnotations(pbAnnotations)

	return pbTool, nil
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

// ConvertProtoToMCPTool transforms a protobuf-defined *pb.Tool into an
// *mcp.Tool. This is the reverse of convertMCPToolToProto and is used when
// exposing internally defined tools to the outside world.
func ConvertProtoToMCPTool(pbTool *pb.Tool) (*mcp.Tool, error) {
	if pbTool == nil {
		return nil, fmt.Errorf("cannot convert nil pb tool to mcp tool")
	}

	sanitizedToolName, err := util.SanitizeToolName(pbTool.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize tool name: %w", err)
	}
	toolID := pbTool.GetServiceId() + "." + sanitizedToolName

	mcpTool := &mcp.Tool{
		Name:        toolID,
		Description: pbTool.GetDescription(),
		Title:       pbTool.GetDisplayName(),
	}

	if pbTool.GetAnnotations() != nil {
		annotations := pbTool.GetAnnotations()
		mcpTool.Annotations = &mcp.ToolAnnotations{
			Title:           annotations.GetTitle(),
			ReadOnlyHint:    annotations.GetReadOnlyHint(),
			IdempotentHint:  annotations.GetIdempotentHint(),
			DestructiveHint: proto.Bool(annotations.GetDestructiveHint()),
			OpenWorldHint:   proto.Bool(annotations.GetOpenWorldHint()),
		}

		if annotations.GetInputSchema() != nil {
			mcpTool.InputSchema = annotations.GetInputSchema().AsMap()
		}
		if annotations.GetOutputSchema() != nil {
			mcpTool.OutputSchema = annotations.GetOutputSchema().AsMap()
		}
	}

	return mcpTool, nil
}
