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

package openapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mcpxy/core/pkg/util"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ParsedOpenAPIData holds the high-level information extracted from an OpenAPI
// specification, such as metadata, server details, and the defined paths.
type ParsedOpenAPIData struct {
	Info    openapi3.Info
	Servers openapi3.Servers
	Paths   map[string]*PathItem
}

// PathItem represents a single path within an OpenAPI specification and holds a
// reference to its corresponding openapi3.PathItem.
type PathItem struct {
	PathRef *openapi3.PathItem
}

// McpOperation provides a simplified, MCP-centric representation of an OpenAPI
// operation. It contains the essential details needed to convert an API
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
// It also returns the original openapi3.T document for more detailed
// inspection if needed.
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

// ExtractMcpOperationsFromOpenAPI iterates through the parsed OpenAPI spec
// and extracts information into a list of McpOperation structs. It takes the
// full 'doc' as input now for easier access to components.
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

// ConvertMcpOperationsToTools converts McpOperation structs into MCP Tool
// protobuf messages.
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

		// For RequestFields, pick the most common content type, e.g., application/json
		if _, ok := op.RequestBodySchema["application/json"]; ok {
			// requestPbFields = convertSchemaToPbFields(schemaRef, doc)
		} else if len(op.RequestBodySchema) > 0 { // Pick first available if no json
			for _, sr := range op.RequestBodySchema {
				_ = convertSchemaToPbFields(sr, doc)
				break
			}
		}

		// For ResponseFields, pick the most common success response (e.g., "200")
		// and content type (e.g., "application/json")
		if responseContent, ok := op.ResponseSchemas["200"]; ok {
			if _, ok := responseContent["application/json"]; ok {
				// responsePbFields = convertSchemaToPbFields(schemaRef, doc)
			} else if len(responseContent) > 0 { // Pick first available if no json
				for _, sr := range responseContent {
					_ = convertSchemaToPbFields(sr, doc)
					break
				}
			}
		} else if responseContent, ok := op.ResponseSchemas["201"]; ok { // Check for 201 if 200 not found
			if _, ok := responseContent["application/json"]; ok {
				// responsePbFields = convertSchemaToPbFields(schemaRef, doc)
			} else if len(responseContent) > 0 {
				for _, sr := range responseContent {
					_ = convertSchemaToPbFields(sr, doc)
					break
				}
			}
		} else if len(op.ResponseSchemas) > 0 { // Pick first available status if no 200/201
			for _, responseContent := range op.ResponseSchemas {
				if len(responseContent) > 0 {
					for _, sr := range responseContent {
						_ = convertSchemaToPbFields(sr, doc)
						goto foundResponse // break out of nested loops
					}
				}
			}
		foundResponse:
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

		inputSchemaProps, err := convertOpenAPISchemaToInputSchemaProperties(bodySchemaRef, op.Parameters, doc)
		if err != nil {
			// Use baseOperationID for the error message as toolID is removed.
			fmt.Printf("Warning: Failed to convert OpenAPI schema to InputSchema for tool '%s': %v. Input schema will be empty.\n", baseOperationID, err)
			inputSchemaProps = &structpb.Struct{Fields: make(map[string]*structpb.Value)} // Empty properties
		}

		toolBuilder := pb.Tool_builder{
			Name:                proto.String(baseOperationID),
			DisplayName:         proto.String(displayName),
			Description:         proto.String(op.Description),
			ServiceId:           proto.String(mcpServerServiceKey),
			InputSchema:         pb.InputSchema_builder{Type: proto.String("object"), Properties: inputSchemaProps}.Build(),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("%s %s", op.Method, op.Path)),
			Annotations: pb.ToolAnnotations_builder{
				Title:          proto.String(op.Summary),
				IdempotentHint: proto.Bool(isOperationIdempotent(op.Method)),
				ReadOnlyHint:   proto.Bool(op.Method == "GET"),
				OpenWorldHint:  proto.Bool(true), // Default, can be refined
			}.Build(),
		}
		tools = append(tools, toolBuilder.Build())
	}
	return tools
}

// isOperationIdempotent checks common HTTP methods for idempotency.
func isOperationIdempotent(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS", "TRACE", "PUT", "DELETE":
		return true
	default: // POST, PATCH etc. are generally not idempotent
		return false
	}
}

// convertOpenAPISchemaToInputSchemaProperties converts an OpenAPI SchemaRef
// and a list of Parameters into a *structpb.Struct suitable for
// InputSchema.Properties.
func convertOpenAPISchemaToInputSchemaProperties(
	bodySchemaRef *openapi3.SchemaRef, // Schema for the request body
	opParameters openapi3.Parameters, // Parameters for the operation (query, path, header)
	doc *openapi3.T,
) (*structpb.Struct, error) {
	props := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	// Helper to convert a single OpenAPI Schema (or SchemaRef) to a structpb.Value representing its schema
	// Now accepts an optional explicitDescription to be used for the schema.
	var convertSingleSchema func(name string, sr *openapi3.SchemaRef, explicitDescription string) (*structpb.Value, error)
	convertSingleSchema = func(name string, sr *openapi3.SchemaRef, explicitDescription string) (*structpb.Value, error) {
		if sr == nil {
			return nil, fmt.Errorf("schema reference is nil for %s", name)
		}

		var sVal *openapi3.Schema
		if sr.Ref != "" {
			refName := strings.TrimPrefix(sr.Ref, "#/components/schemas/")
			if doc != nil && doc.Components != nil && doc.Components.Schemas != nil {
				if componentSchemaRef, ok := doc.Components.Schemas[refName]; ok {
					sVal = componentSchemaRef.Value
				} else {
					return nil, fmt.Errorf("could not resolve schema reference: %s", sr.Ref)
				}
			} else {
				return nil, fmt.Errorf("cannot resolve schema reference '%s' due to nil doc or components", sr.Ref)
			}
		} else {
			sVal = sr.Value
		}

		if sVal == nil {
			return nil, fmt.Errorf("schema value is nil for %s after potential dereferencing", name)
		}

		// Ensure sVal.Type is not nil before accessing it
		schemaType := "object" // Default or if type is not specified for an object-like schema
		if sVal.Type != nil && len(*sVal.Type) > 0 {
			schemaType = (*sVal.Type)[0]
		}

		description := sVal.Description
		if explicitDescription != "" {
			description = explicitDescription // Prioritize explicit description if provided
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
			fieldSchema["type"] = "object"
			nestedProps := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
			if sVal.Properties != nil {
				for propName, propSchemaRef := range sVal.Properties {
					// For nested properties, explicitDescription is not passed down, they use their own schema's description
					nestedVal, err := convertSingleSchema(propName, propSchemaRef, "")
					if err != nil {
						// Log or collect errors for properties?
						// For now, skip problematic properties.
						fmt.Printf("Warning: skipping property '%s' of object '%s': %v\n", propName, name, err)
						continue
					}
					nestedProps.Fields[propName] = nestedVal
				}
			}
			// If it's an object and has no properties, it's an empty object.
			// Otherwise, add its properties.
			fieldSchema["properties"] = structpb.NewStructValue(nestedProps)

		case openapi3.TypeArray:
			fieldSchema["type"] = "array"
			if sVal.Items != nil {
				// For array items, explicitDescription is not passed down.
				itemSchemaVal, err := convertSingleSchema(name+"_items", sVal.Items, "")
				if err != nil {
					// If item schema cannot be converted, this array is underspecified.
					fmt.Printf("Warning: could not determine item type for array '%s': %v\n", name, err)
					// Potentially represent as array of any or skip items field.
					// For now, we'll just not set the "items" field in the schema.
				} else if itemSchemaVal != nil {
					// The 'items' field should be a schema object (map), not a Value.
					// We need to unwrap the struct from the Value and convert it to a map.
					if sv := itemSchemaVal.GetStructValue(); sv != nil {
						fieldSchema["items"] = sv.AsMap()
					}
				}
			} else {
				// Array with no items specified.
				fmt.Printf("Warning: array '%s' has no items schema specified.\n", name)
			}
		case openapi3.TypeString, openapi3.TypeNumber, openapi3.TypeInteger, openapi3.TypeBoolean:
			// sVal.Type should be non-nil and have at least one element here.
			fieldSchema["type"] = (*sVal.Type)[0]
		default:
			// This case should ideally not be reached if schemaType is derived from sVal.Type.
			// If sVal.Type was nil or empty, schemaType defaults to "object", handled above.
			// If it's some other unknown type string from sVal.Type[0].
			fmt.Printf("Warning: unhandled schema type '%s' for field '%s'. Defaulting to 'string'.\n", schemaType, name)
			fieldSchema["type"] = "string" // Fallback for unknown types
		}

		finalSchemaStruct, err := structpb.NewStruct(fieldSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field schema map for '%s' to structpb.Struct: %w", name, err)
		}
		return structpb.NewStructValue(finalSchemaStruct), nil
	}

	// Process request body schema (if any, typically for application/json)
	if bodySchemaRef != nil {
		// The request body itself is not a named parameter, but its properties become top-level.
		// If the body schema is an object, its properties are merged directly into props.Fields.
		// If it's not an object (e.g. a direct array or primitive), it's harder to represent as "properties".
		// For now, assume request body (if present) is an object and its properties are merged.

		var bodyActualSchema *openapi3.Schema
		if bodySchemaRef.Ref != "" {
			refName := strings.TrimPrefix(bodySchemaRef.Ref, "#/components/schemas/")
			if doc != nil && doc.Components != nil && doc.Components.Schemas != nil {
				if componentSchemaRef, ok := doc.Components.Schemas[refName]; ok {
					bodyActualSchema = componentSchemaRef.Value
				} else {
					return nil, fmt.Errorf("could not resolve request body schema reference: %s", bodySchemaRef.Ref)
				}
			} else {
				return nil, fmt.Errorf("cannot resolve request body schema reference '%s' due to nil doc or components", bodySchemaRef.Ref)
			}
		} else {
			bodyActualSchema = bodySchemaRef.Value
		}

		if bodyActualSchema != nil {
			// Ensure bodyActualSchema.Type is not nil before accessing it
			bodySchemaType := "object" // Default if type is not specified
			if bodyActualSchema.Type != nil && len(*bodyActualSchema.Type) > 0 {
				bodySchemaType = (*bodyActualSchema.Type)[0]
			}

			if bodySchemaType == openapi3.TypeObject {
				for propName, propSchemaRef := range bodyActualSchema.Properties {
					// For request body properties, explicitDescription is not passed.
					val, err := convertSingleSchema(propName, propSchemaRef, "")
					if err != nil {
						fmt.Printf("Warning: skipping property '%s' from request body: %v\n", propName, err)
						continue
					}
					props.Fields[propName] = val
				}
			} else {
				// If the request body is not an object (e.g., an array or a primitive),
				// wrap its schema under a special "request_body" property.
				val, err := convertSingleSchema("request_body", bodySchemaRef, bodyActualSchema.Description)
				if err != nil {
					// Log the error but continue; this might result in an empty input schema for this part.
					fmt.Printf("Warning: Failed to convert non-object request body schema: %v\n", err)
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
				fmt.Printf("Warning: parameter '%s' in '%s' has no schema, skipping.\n", param.Name, param.In)
				continue
			}
			// Pass param.Description as the explicit description for the parameter's schema
			val, err := convertSingleSchema(param.Name, param.Schema, param.Description)
			if err != nil {
				fmt.Printf("Warning: skipping parameter '%s': %v\n", param.Name, err)
				continue
			}
			props.Fields[param.Name] = val
		}
	}

	return props, nil
}

// convertSchemaToPbFields converts an OpenAPI schema (and its properties if
// it's an object) into a slice of MCP Field protobuf messages.
func convertSchemaToPbFields(schemaRef *openapi3.SchemaRef, doc *openapi3.T) []*pb.Field {
	var pbFields []*pb.Field
	if schemaRef == nil {
		return pbFields
	}

	var schema *openapi3.Schema
	if schemaRef.Ref != "" {
		// Attempt to resolve the reference using the document's components.
		// This is a simplified $ref resolver. kin-openapi's loader should generally handle this
		// if components are correctly defined and $refs are within the same document or resolvable.
		refName := strings.TrimPrefix(schemaRef.Ref, "#/components/schemas/")
		if doc != nil && doc.Components != nil && doc.Components.Schemas != nil {
			if componentSchemaRef, ok := doc.Components.Schemas[refName]; ok {
				schema = componentSchemaRef.Value // Use the resolved schema's value
			} else {
				// If ref not found in components, or it's an external/complex ref not supported by this simple resolver
				// We might fall back to schemaRef.Value if it's somehow pre-resolved or contains partial info.
				// However, if Ref is set, Value is usually nil until resolved.
				// For robustness, if we can't resolve, we can't proceed with this schema.
				// Or, one could try `schemaRef.Resolve(doc)` if loader context is set up for it.
				// For now, if direct lookup fails, we assume it's unresolvable by this simple code.
				fmt.Printf("Warning: Could not resolve schema reference: %s\n", schemaRef.Ref)
				return pbFields // Cannot proceed with this schema
			}
		} else {
			// Document or components are nil, cannot resolve.
			fmt.Printf("Warning: Cannot resolve schema reference '%s' due to nil doc or components.\n", schemaRef.Ref)
			return pbFields
		}
	} else {
		schema = schemaRef.Value
	}

	if schema == nil {
		return pbFields
	}

	// Handle nil schema.Type or empty type array
	if schema.Type == nil || len(*schema.Type) == 0 {
		// If type is not specified or empty, cannot determine structure.
		// Consider logging this or returning a default field if appropriate.
		return pbFields
	}
	primaryType := (*schema.Type)[0] // Use the first type for switch

	switch primaryType { // Use the first type for switching logic
	case openapi3.TypeObject:
		if schema.Properties != nil {
			for propName, propSchemaRef := range schema.Properties {
				// Recursively call for nested schemas, but pass the propSchemaRef itself
				// This simple version doesn't try to represent full nested structure in pb.Fields,
				// but it extracts properties of the current object.
				// For a property that is itself an object or array, its 'Type' will reflect that.

				var propSchemaValue *openapi3.Schema
				if propSchemaRef.Ref != "" {
					refName := strings.TrimPrefix(propSchemaRef.Ref, "#/components/schemas/")
					if doc != nil && doc.Components != nil && doc.Components.Schemas != nil {
						if componentSchemaRef, ok := doc.Components.Schemas[refName]; ok {
							propSchemaValue = componentSchemaRef.Value
						} else {
							fmt.Printf("Warning: Could not resolve property schema reference: %s\n", propSchemaRef.Ref)
							continue
						}
					} else {
						fmt.Printf("Warning: Cannot resolve property schema reference '%s' due to nil doc or components.\n", propSchemaRef.Ref)
						continue
					}
				} else {
					propSchemaValue = propSchemaRef.Value
				}

				if propSchemaValue != nil {
					var fieldTypeStr string
					if propSchemaValue.Type != nil && len(*propSchemaValue.Type) > 0 {
						fieldTypeStr = (*propSchemaValue.Type)[0] // Use first type
					} else {
						fieldTypeStr = "unknown_property_type"
					}
					fieldBuilder := pb.Field_builder{
						Name:        &propName,
						Description: &propSchemaValue.Description,
						Type:        &fieldTypeStr,
					}
					// Check type of property, if it's an array, format its item type
					if propSchemaValue.Type != nil && len(*propSchemaValue.Type) > 0 && (*propSchemaValue.Type)[0] == openapi3.TypeArray &&
						propSchemaValue.Items != nil && propSchemaValue.Items.Value != nil &&
						propSchemaValue.Items.Value.Type != nil && len(*propSchemaValue.Items.Value.Type) > 0 {
						val := fmt.Sprintf("array[%s]", (*propSchemaValue.Items.Value.Type)[0])
						fieldBuilder.Type = &val
					}
					// Note: For nested objects (fieldTypeStr == "object"), field.Type will be "object".
					// Deeper conversion of nested objects into a flat list of pb.Field is complex
					// and depends on how you want to represent that structure.
					pbFields = append(pbFields, fieldBuilder.Build())
				}
			}
		}
	case openapi3.TypeArray:
		fieldType := "array"
		if schema.Items != nil { // Check schema.Items itself first
			itemSchemaRef := schema.Items // itemSchemaRef is *openapi3.SchemaRef
			itemTypeStr := "unknown_array_item_type"

			if itemSchemaRef.Ref != "" {
				itemTypeStr = strings.TrimPrefix(itemSchemaRef.Ref, "#/components/schemas/")
			} else if itemSchemaRef.Value != nil && itemSchemaRef.Value.Type != nil && len(*itemSchemaRef.Value.Type) > 0 {
				itemTypeStr = (*itemSchemaRef.Value.Type)[0] // Use first type
			}
			fieldType = fmt.Sprintf("array[%s]", itemTypeStr)
		}
		name := "array_items"
		pbFields = append(pbFields, pb.Field_builder{
			Name:        &name, // Placeholder name if schema itself is an array
			Description: &schema.Description,
			Type:        &fieldType,
		}.Build())
	default: // string, number, integer, boolean
		// schema.Type is already confirmed non-nil and non-empty here due to earlier check and primaryType assignment
		fieldTypeStr := primaryType
		name := "value"
		pbFields = append(pbFields, pb.Field_builder{
			Name:        &name, // Placeholder name if schema is a primitive
			Description: &schema.Description,
			Type:        &fieldTypeStr,
		}.Build())
	}

	return pbFields
}
