// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/upstream/grpc/protobufparser"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ConvertMCPToolToProto transforms an *mcp.Tool, which uses a flexible schema.
//
// Summary: transforms an *mcp.Tool, which uses a flexible schema.
//
// Parameters:
//   - tool: *mcp.Tool. The tool.
//
// Returns:
//   - *pb.Tool: The *pb.Tool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func ConvertMCPToolToProto(tool *mcp.Tool) (*pb.Tool, error) {
	if tool == nil {
		return nil, fmt.Errorf("cannot convert nil mcp tool to proto")
	}

	displayName := tool.Name
	if tool.Title != "" {
		displayName = tool.Title
	} else if tool.Annotations != nil && tool.Annotations.Title != "" {
		displayName = tool.Annotations.Title
	}

	annotationsBuilder := pb.ToolAnnotations_builder{}
	if tool.Annotations != nil {
		annotationsBuilder.Title = proto.String(tool.Annotations.Title)
		annotationsBuilder.ReadOnlyHint = proto.Bool(tool.Annotations.ReadOnlyHint)
		annotationsBuilder.IdempotentHint = proto.Bool(tool.Annotations.IdempotentHint)
		if tool.Annotations.DestructiveHint != nil {
			annotationsBuilder.DestructiveHint = proto.Bool(*tool.Annotations.DestructiveHint)
		}
		if tool.Annotations.OpenWorldHint != nil {
			annotationsBuilder.OpenWorldHint = proto.Bool(*tool.Annotations.OpenWorldHint)
		}
	}

	if tool.InputSchema != nil {
		inputSchema, err := SanitizeJSONSchema(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input schema: %w", err)
		}
		annotationsBuilder.InputSchema = inputSchema
	} else {
		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		})
		if err == nil {
			annotationsBuilder.InputSchema = inputSchema
		}
	}

	if tool.OutputSchema != nil {
		outputSchema, err := convertJSONSchemaToStruct(tool.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert output schema: %w", err)
		}
		annotationsBuilder.OutputSchema = outputSchema
	}

	pbTool := pb.Tool_builder{
		Name:         proto.String(tool.Name),
		Description:  proto.String(tool.Description),
		DisplayName:  proto.String(displayName),
		Annotations:  annotationsBuilder.Build(),
		InputSchema:  annotationsBuilder.InputSchema,
		OutputSchema: annotationsBuilder.OutputSchema,
	}.Build()

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

	// Optimization: Use structpb.NewStruct directly instead of round-tripping through JSON.
	// This is significantly faster (~10x) and avoids unnecessary memory allocations.
	return structpb.NewStruct(schemaMap)
}

// ConvertMcpFieldsToInputSchemaProperties converts a slice of McpField, which.
//
// Summary: converts a slice of McpField, which.
//
// Parameters:
//   - fields: []*protobufparser.McpField. The fields.
//
// Returns:
//   - *structpb.Struct: The *structpb.Struct.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func ConvertMcpFieldsToInputSchemaProperties(fields []*protobufparser.McpField) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	for _, field := range fields {
		schema, err := GetJSONSchemaForScalarType(field.Type, field.Description)
		if err != nil {
			return nil, err
		}

		fieldsMap := map[string]interface{}{
			"type":        schema.Type,
			"description": schema.Description,
		}

		value, err := structpb.NewValue(fieldsMap)
		if err != nil {
			return nil, fmt.Errorf("failed to create structpb value: %w", err)
		}

		properties.Fields[field.Name] = value
	}
	return properties, nil
}


// ConvertToolDefinitionToProto transforms a *configv1.ToolDefinition into a.
//
// Summary: transforms a *configv1.ToolDefinition into a.
//
// Parameters:
//   - toolDef: *configv1.ToolDefinition. The toolDef.
//   - inputSchema: *structpb.Struct. The inputSchema.
//   - outputSchema: *structpb.Struct. The outputSchema.
//
// Returns:
//   - *pb.Tool: The *pb.Tool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func ConvertToolDefinitionToProto(toolDef *configv1.ToolDefinition, inputSchema, outputSchema *structpb.Struct) (*pb.Tool, error) {
	if toolDef == nil {
		return nil, fmt.Errorf("cannot convert nil tool definition to proto")
	}

	annotationsBuilder := pb.ToolAnnotations_builder{
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}

	var profiles []string
	for _, p := range toolDef.GetProfiles() {
		if p.GetId() != "" {
			profiles = append(profiles, p.GetId())
		} else if p.GetName() != "" {
			profiles = append(profiles, p.GetName())
		}
	}

	builder := pb.Tool_builder{
		Name:        proto.String(toolDef.GetName()),
		Description: proto.String(toolDef.GetDescription()),
		DisplayName: proto.String(toolDef.GetTitle()),
		ServiceId:   proto.String(toolDef.GetServiceId()),
		Annotations: annotationsBuilder.Build(),
		Tags:        toolDef.GetTags(),
		Profiles:    profiles,
	}

	if toolDef.GetIntegrity() != nil {
		builder.Integrity = pb.ToolIntegrity_builder{
			Hash:      proto.String(toolDef.GetIntegrity().GetHash()),
			Algorithm: proto.String(toolDef.GetIntegrity().GetAlgorithm()),
		}.Build()
	}

	return builder.Build(), nil
}

// GetJSONSchemaForScalarType maps a protobuf scalar type (e.g., "TYPE_STRING",.
//
// Summary: maps a protobuf scalar type (e.g., "TYPE_STRING",.
//
// Parameters:
//   - scalarType: string. The scalarType.
//   - description: string. The description.
//
// Returns:
//   - *jsonschema.Schema: The *jsonschema.Schema.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func GetJSONSchemaForScalarType(scalarType, description string) (*jsonschema.Schema, error) {
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


// ConvertProtoToMCPTool transforms a protobuf-defined *pb.Tool into an.
//
// Summary: transforms a protobuf-defined *pb.Tool into an.
//
// Parameters:
//   - pbTool: *pb.Tool. The pbTool.
//
// Returns:
//   - *mcp.Tool: The *mcp.Tool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func ConvertProtoToMCPTool(pbTool *pb.Tool) (*mcp.Tool, error) {
	if pbTool == nil {
		return nil, fmt.Errorf("cannot convert nil pb tool to mcp tool")
	}

	if pbTool.GetName() == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	mcpTool := &mcp.Tool{
		Name:        pbTool.GetServiceId() + "." + pbTool.GetName(),
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
			schema := annotations.GetOutputSchema().AsMap()
			// Only include output schema if it's an object, as the MCP SDK might panic otherwise
			if t, ok := schema["type"].(string); ok && t == "object" {
				mcpTool.OutputSchema = schema
			}
		}
	}

	return mcpTool, nil
}
