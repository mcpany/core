// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	typeObject = "object"
	typeString = "string"
	methodGet  = "GET"
)

// ParsedOpenAPIData holds the high-level information extracted from an OpenAPI
// specification, such as metadata, server details, and the defined paths.
//
// Summary: Extracted OpenAPI data.
type ParsedOpenAPIData struct {
	Info    openapi3.Info
	Servers openapi3.Servers
	Paths   map[string]*PathItem
}

// PathItem represents a single path within an OpenAPI specification and holds a
// reference to its corresponding openapi3.PathItem.
//
// Summary: OpenAPI Path Item wrapper.
type PathItem struct {
	PathRef *openapi3.PathItem
}

// McpOperation provides a simplified, MCP-centric representation of an OpenAPI
// operation.
//
// Summary: Simplified OpenAPI operation for MCP.
//
// It contains the essential details needed to convert an API
// endpoint into an executable tool.
type McpOperation struct {
	OperationID string
	Summary     string
	Description string
	Method      string // e.g., GET, POST
	Path        string
	// Simplified schema representation for request body
	// Key: content type (e.g., "application/json")
	// Value: schema reference
	RequestBodySchema map[string]*openapi3.SchemaRef
	// Simplified schema representation for responses
	// Key: status code (e.g., "200")
	// Value: map of content type to schema reference
	ResponseSchemas map[string]map[string]*openapi3.SchemaRef
	Parameters      openapi3.Parameters // Store operation parameters (query, path, header, cookie)
}

// ParseOpenAPISpec loads and parses an OpenAPI specification from a byte slice.
// It validates the spec and returns both a simplified ParsedOpenAPIData view
// and the original, more detailed openapi3.T document.
func parseOpenAPISpec(ctx context.Context, specData []byte) (*ParsedOpenAPIData, *openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true // Depending on requirements

	// Load the spec from the byte slice
	doc, err := loader.LoadFromData(specData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load OpenAPI spec from data: %w", err)
	}

	// It's important to validate the spec.
	if err = doc.Validate(ctx); err != nil {
		return nil, nil, fmt.Errorf("OpenAPI spec validation failed: %w", err)
	}

	parsedData := &ParsedOpenAPIData{
		Info:    *doc.Info,
		Servers: doc.Servers,
		Paths:   make(map[string]*PathItem),
	}

	// Using .Map() is safer as it handles nil doc.Paths gracefully.
	if doc.Paths != nil {
		for path, pathItem := range doc.Paths.Map() {
			if pathItem == nil {
				continue
			}
			parsedData.Paths[path] = &PathItem{PathRef: pathItem}
		}
	}

	return parsedData, doc, nil
}

// ExtractMcpOperationsFromOpenAPI iterates through the paths and methods of a
// parsed OpenAPI document and transforms each operation into a simplified
// McpOperation struct, which is more convenient for tool registration.
func extractMcpOperationsFromOpenAPI(doc *openapi3.T) []McpOperation {
	var mcpOps []McpOperation

	if doc == nil || doc.Paths == nil {
		return mcpOps
	}

	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			op := McpOperation{
				OperationID:       operation.OperationID,
				Summary:           operation.Summary,
				Description:       operation.Description,
				Method:            method,
				Path:              path,
				RequestBodySchema: make(map[string]*openapi3.SchemaRef),
				ResponseSchemas:   make(map[string]map[string]*openapi3.SchemaRef),
				Parameters:        operation.Parameters, // Populate parameters
			}

			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				for contentType, mediaType := range operation.RequestBody.Value.Content {
					if mediaType != nil && mediaType.Schema != nil {
						op.RequestBodySchema[contentType] = mediaType.Schema
					}
				}
			}

			if operation.Responses != nil && operation.Responses.Map() != nil {
				for statusCode, responseRef := range operation.Responses.Map() {
					if responseRef != nil && responseRef.Value != nil {
						op.ResponseSchemas[statusCode] = make(map[string]*openapi3.SchemaRef)
						for contentType, mediaType := range responseRef.Value.Content {
							if mediaType != nil && mediaType.Schema != nil {
								op.ResponseSchemas[statusCode][contentType] = mediaType.Schema
							}
						}
					}
				}
			}
			mcpOps = append(mcpOps, op)
		}
	}
	return mcpOps
}

// ConvertMcpOperationsToTools transforms a slice of McpOperation structs into a
// slice of MCP Tool protobuf messages, which can then be registered with the
// tool manager.
func convertMcpOperationsToTools(ops []McpOperation, doc *openapi3.T, mcpServerServiceKey string) []*pb.Tool {
	tools := make([]*pb.Tool, 0, len(ops))

	if doc == nil {
		return tools
	}

	for _, op := range ops {
		var baseOperationID string
		if op.OperationID != "" {
			baseOperationID = util.SanitizeOperationID(op.OperationID)
		} else {
			baseOperationID = util.SanitizeOperationID(op.Method + "_" + op.Path)
		}
		if baseOperationID == "" {
			baseOperationID = "unnamed_operation"
		}

		displayName := op.Summary
		if displayName == "" {
			displayName = op.OperationID // Fallback to OperationID
		}
		if displayName == "" { // Further fallback
			displayName = op.Method + " " + op.Path
		}

		// Determine request body schema (e.g. application/json)
		var bodySchemaRef *openapi3.SchemaRef
		if ref, ok := op.RequestBodySchema["application/json"]; ok {
			bodySchemaRef = ref
		} else if len(op.RequestBodySchema) > 0 { // Pick first available if no json
			for _, sr := range op.RequestBodySchema {
				bodySchemaRef = sr
				break
			}
		}

		properties, err := convertOpenAPISchemaToInputSchemaProperties(bodySchemaRef, op.Parameters, doc)
		if err != nil {
			// Use baseOperationID for the error message as toolID is removed.
			logging.GetLogger().Warn("Failed to convert OpenAPI schema to InputSchema for tool. Input schema will be empty.", "tool", baseOperationID, "error", err)
			properties = &structpb.Struct{Fields: make(map[string]*structpb.Value)} // Empty properties
		}

		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue(typeObject),
				"properties": structpb.NewStructValue(properties),
			},
		}

		// Determine response body schema for output
		var outputSchemaRef *openapi3.SchemaRef
		if responseContent, ok := op.ResponseSchemas["200"]; ok {
			if ref, ok := responseContent["application/json"]; ok {
				outputSchemaRef = ref
			}
		} else if responseContent, ok := op.ResponseSchemas["201"]; ok {
			if ref, ok := responseContent["application/json"]; ok {
				outputSchemaRef = ref
			}
		}

		outputProperties, err := convertOpenAPISchemaToOutputSchemaProperties(outputSchemaRef, doc)
		if err != nil {
			logging.GetLogger().Warn("Failed to convert OpenAPI schema to OutputSchema for tool. Output schema will be empty.", "tool", baseOperationID, "error", err)
			outputProperties = &structpb.Struct{Fields: make(map[string]*structpb.Value)} // Empty properties
		}

		outputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue(typeObject),
				"properties": structpb.NewStructValue(outputProperties),
			},
		}

		toolBuilder := pb.Tool_builder{
			Name:                proto.String(baseOperationID),
			DisplayName:         proto.String(displayName),
			Description:         proto.String(op.Description),
			ServiceId:           proto.String(mcpServerServiceKey),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("%s %s", op.Method, op.Path)),
			InputSchema:         inputSchema,
			OutputSchema:        outputSchema,
			Annotations: pb.ToolAnnotations_builder{
				Title:          proto.String(op.Summary),
				IdempotentHint: proto.Bool(isOperationIdempotent(op.Method)),
				ReadOnlyHint:   proto.Bool(op.Method == methodGet),
				OpenWorldHint:  proto.Bool(true), // Default, can be refined
				InputSchema:    inputSchema,
				OutputSchema:   outputSchema,
			}.Build(),
		}
		tools = append(tools, toolBuilder.Build())
	}
	return tools
}

// isOperationIdempotent checks common HTTP methods for idempotency, which is a
// useful hint for AI models using the tools.
func isOperationIdempotent(method string) bool {
	switch strings.ToUpper(method) {
	case methodGet, "HEAD", "OPTIONS", "TRACE", "PUT", "DELETE":
		return true
	default: // POST, PATCH etc. are generally not idempotent
		return false
	}
}

func convertOpenAPISchemaToOutputSchemaProperties(
	bodySchemaRef *openapi3.SchemaRef, // Schema for the response body
	doc *openapi3.T,
) (*structpb.Struct, error) {
	props := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	if bodySchemaRef != nil {
		bodyActualSchema, err := resolveSchemaRef(bodySchemaRef, doc)
		if err != nil {
			return nil, fmt.Errorf("could not resolve request body schema reference: %w", err)
		}

		if bodyActualSchema != nil {
			isObject := len(bodyActualSchema.Properties) > 0
			if bodyActualSchema.Type != nil && len(*bodyActualSchema.Type) > 0 {
				isObject = (*bodyActualSchema.Type)[0] == openapi3.TypeObject
			}
			if !isObject && len(bodyActualSchema.AllOf) > 0 {
				// If AllOf is present, check if any component implies object or if we should treat it as one.
				// Usually AllOf is used for object composition.
				isObject = true
			}

			if isObject {
				mergedProps, err := mergeSchemaProperties(bodyActualSchema, doc)
				if err != nil {
					logging.GetLogger().Warn("Failed to merge properties", "error", err)
				}
				for propName, propSchemaRef := range mergedProps {
					val, err := convertSchemaToStructPB(propName, propSchemaRef, "", doc)
					if err != nil {
						logging.GetLogger().Warn("Skipping property from request body", "property", propName, "error", err)
						continue
					}
					props.Fields[propName] = val
				}
			} else {
				val, err := convertSchemaToStructPB("response_body", bodySchemaRef, bodyActualSchema.Description, doc)
				if err != nil {
					logging.GetLogger().Warn("Failed to convert non-object request body schema", "error", err)
				} else {
					props.Fields["response_body"] = val
				}
			}
		}
	}

	return props, nil
}

// convertOpenAPISchemaToInputSchemaProperties converts an OpenAPI SchemaRef and a
// list of Parameters into a *structpb.Struct that is suitable for use as the
// Properties field of an InputSchema.
func convertOpenAPISchemaToInputSchemaProperties(
	bodySchemaRef *openapi3.SchemaRef, // Schema for the request body
	opParameters openapi3.Parameters, // Parameters for the operation (query, path, header)
	doc *openapi3.T,
) (*structpb.Struct, error) {
	props := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	// Process request body schema (if any, typically for application/json)
	if bodySchemaRef != nil {
		bodyActualSchema, err := resolveSchemaRef(bodySchemaRef, doc)
		if err != nil {
			return nil, fmt.Errorf("could not resolve request body schema reference: %w", err)
		}

		if bodyActualSchema != nil {
			// Check if the schema is a non-object type (e.g., array, string).
			isObject := len(bodyActualSchema.Properties) > 0
			if bodyActualSchema.Type != nil && len(*bodyActualSchema.Type) > 0 {
				isObject = (*bodyActualSchema.Type)[0] == openapi3.TypeObject
			}
			if !isObject && len(bodyActualSchema.AllOf) > 0 {
				isObject = true
			}

			if isObject {
				mergedProps, err := mergeSchemaProperties(bodyActualSchema, doc)
				if err != nil {
					logging.GetLogger().Warn("Failed to merge properties", "error", err)
				}
				for propName, propSchemaRef := range mergedProps {
					val, err := convertSchemaToStructPB(propName, propSchemaRef, "", doc)
					if err != nil {
						logging.GetLogger().Warn("Skipping property from request body", "property", propName, "error", err)
						continue
					}
					props.Fields[propName] = val
				}
			} else {
				// If the request body is not an object, wrap its schema under "request_body".
				val, err := convertSchemaToStructPB("request_body", bodySchemaRef, bodyActualSchema.Description, doc)
				if err != nil {
					logging.GetLogger().Warn("Failed to convert non-object request body schema", "error", err)
				} else {
					props.Fields["request_body"] = val
				}
			}
		}
	}

	// Process operation parameters (query, path, header)
	for _, paramRef := range opParameters {
		if paramRef == nil {
			continue
		}
		param := paramRef.Value
		if param == nil { // Should not happen if paramRef is valid
			continue
		}
		// We are interested in parameters that are part of the input: query, path, header. Cookie params are ignored for now.
		if param.In == openapi3.ParameterInQuery || param.In == openapi3.ParameterInPath || param.In == openapi3.ParameterInHeader {
			if param.Schema == nil { // Parameter must have a schema
				logging.GetLogger().Warn("Warning: parameter '%s' in '%s' has no schema, skipping.\n", param.Name, param.In)
				continue
			}
			// Pass param.Description as the explicit description for the parameter's schema
			val, err := convertSchemaToStructPB(param.Name, param.Schema, param.Description, doc)
			if err != nil {
				logging.GetLogger().Warn("Warning: skipping parameter '%s': %v\n", param.Name, err)
				continue
			}
			props.Fields[param.Name] = val
		}
	}

	return props, nil
}

// convertSchemaToStructPB converts a single OpenAPI Schema (or SchemaRef) to a *structpb.Value representing its schema.
func convertSchemaToStructPB(name string, sr *openapi3.SchemaRef, explicitDescription string, doc *openapi3.T) (*structpb.Value, error) {
	sVal, err := resolveSchemaRef(sr, doc)
	if err != nil {
		return nil, fmt.Errorf("schema reference resolution failed for %s: %w", name, err)
	}
	if sVal == nil {
		return nil, fmt.Errorf("schema value is nil for %s", name)
	}

	schemaType := typeObject // Default
	if sVal.Type != nil && len(*sVal.Type) > 0 {
		schemaType = (*sVal.Type)[0]
	}
	// If type is missing but AllOf is present, treat as object
	if (sVal.Type == nil || len(*sVal.Type) == 0) && len(sVal.AllOf) > 0 {
		schemaType = typeObject
	}

	description := sVal.Description
	if explicitDescription != "" {
		description = explicitDescription
	}

	fieldSchema := map[string]interface{}{
		"description": description,
	}
	if sVal.Format != "" {
		fieldSchema["format"] = sVal.Format
	}
	if len(sVal.Enum) > 0 {
		var enums []interface{}
		enums = append(enums, sVal.Enum...)
		fieldSchema["enum"] = enums
	}
	if sVal.Default != nil {
		fieldSchema["default"] = sVal.Default
	}

	switch schemaType {
	case openapi3.TypeObject:
		fieldSchema["type"] = typeObject
		nestedProps := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

		mergedProps, err := mergeSchemaProperties(sVal, doc)
		if err != nil {
			logging.GetLogger().Warn("Warning: failed to merge properties for object '%s': %v\n", name, err)
		}

		for propName, propSchemaRef := range mergedProps {
			nestedVal, err := convertSchemaToStructPB(propName, propSchemaRef, "", doc)
			if err != nil {
				logging.GetLogger().Warn("Skipping property of object", "property", propName, "object", name, "error", err)
				continue
			}
			nestedProps.Fields[propName] = nestedVal
		}
		// Fix for nested objects: use AsMap() to allow structpb.NewStruct to recurse properly
		fieldSchema["properties"] = nestedProps.AsMap()

	case openapi3.TypeArray:
		fieldSchema["type"] = "array"
		if sVal.Items != nil {
			itemSchemaVal, err := convertSchemaToStructPB(name+"_items", sVal.Items, "", doc)
			if err != nil {
				logging.GetLogger().Warn("Could not determine item type for array", "array", name, "error", err)
			} else if itemSchemaVal != nil {
				if sv := itemSchemaVal.GetStructValue(); sv != nil {
					fieldSchema["items"] = sv.AsMap()
				}
			}
		} else {
			logging.GetLogger().Warn("Array has no items schema specified", "array", name)
		}
	case openapi3.TypeString, openapi3.TypeNumber, openapi3.TypeInteger, openapi3.TypeBoolean:
		fieldSchema["type"] = (*sVal.Type)[0]
	default:
		logging.GetLogger().Warn("Unhandled schema type for field. Defaulting to 'string'.", "type", schemaType, "field", name)
		fieldSchema["type"] = typeString
	}

	finalSchemaStruct, err := structpb.NewStruct(fieldSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to convert field schema map for '%s' to structpb.Struct: %w", name, err)
	}
	return structpb.NewStructValue(finalSchemaStruct), nil
}

// mergeSchemaProperties returns a merged map of properties from the schema and its AllOf components.
func mergeSchemaProperties(s *openapi3.Schema, doc *openapi3.T) (map[string]*openapi3.SchemaRef, error) {
	props := make(map[string]*openapi3.SchemaRef)

	// Helper to recursively merge properties
	var merge func(*openapi3.Schema) error
	merge = func(curr *openapi3.Schema) error {
		if curr == nil {
			return nil
		}

		for _, ref := range curr.AllOf {
			sub, err := resolveSchemaRef(ref, doc)
			if err != nil {
				return fmt.Errorf("failed to resolve AllOf schema ref: %w", err)
			}
			if sub != nil {
				if err := merge(sub); err != nil {
					return err
				}
			}
		}

		for k, v := range curr.Properties {
			props[k] = v
		}
		return nil
	}

	if err := merge(s); err != nil {
		return nil, err
	}
	return props, nil
}

func resolveSchemaRef(sr *openapi3.SchemaRef, doc *openapi3.T) (*openapi3.Schema, error) {
	if sr == nil {
		return nil, nil
	}
	if sr.Ref == "" {
		return sr.Value, nil
	}

	refName := strings.TrimPrefix(sr.Ref, "#/components/schemas/")
	if doc != nil && doc.Components != nil && doc.Components.Schemas != nil {
		if componentSchemaRef, ok := doc.Components.Schemas[refName]; ok {
			return componentSchemaRef.Value, nil
		}
		return nil, fmt.Errorf("could not resolve schema reference: %s", sr.Ref)
	}
	return nil, fmt.Errorf("cannot resolve schema reference '%s' due to nil doc or components", sr.Ref)
}
