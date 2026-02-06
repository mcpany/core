// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package schemaconv provides utilities for converting between schema formats.
package schemaconv

import (
	"fmt"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

// JSON schema types.
const (
	// TypeNumber represents a JSON number type.
	//
	// Summary: JSON number type.
	TypeNumber = "number"
	// TypeInteger represents a JSON integer type.
	//
	// Summary: JSON integer type.
	TypeInteger = "integer"
	// TypeBoolean represents a JSON boolean type.
	//
	// Summary: JSON boolean type.
	TypeBoolean = "boolean"
	// TypeObject represents a JSON object type.
	//
	// Summary: JSON object type.
	TypeObject = "object"
	// TypeArray represents a JSON array type.
	//
	// Summary: JSON array type.
	TypeArray = "array"
	// TypeString represents a JSON string type.
	//
	// Summary: JSON string type.
	TypeString = "string"
)

// MaxRecursionDepth limits the depth of nested messages to prevent infinite recursion.
//
// Summary: Limit for recursion depth.
const MaxRecursionDepth = 10

// MethodDescriptorToProtoProperties converts the fields of a method's input
// message into a `structpb.Struct` for use as the `properties` field in a tool
// input schema.
//
// Parameters:
//   - methodDesc: The method descriptor to convert.
//
// Returns:
//   - *structpb.Struct: The properties structure for the input schema.
//   - error: An error if the conversion fails.
func MethodDescriptorToProtoProperties(methodDesc protoreflect.MethodDescriptor) (*structpb.Struct, error) {
	return fieldsToProperties(methodDesc.Input().Fields(), 0)
}

// MethodOutputDescriptorToProtoProperties converts the fields of a method's
// output message into a `structpb.Struct` for use as the `properties` field in
// a tool output schema.
//
// Parameters:
//   - methodDesc: The method descriptor to convert.
//
// Returns:
//   - *structpb.Struct: The properties structure for the output schema.
//   - error: An error if the conversion fails.
func MethodOutputDescriptorToProtoProperties(methodDesc protoreflect.MethodDescriptor) (*structpb.Struct, error) {
	return fieldsToProperties(methodDesc.Output().Fields(), 0)
}

func fieldsToProperties(fields protoreflect.FieldDescriptors, depth int) (*structpb.Struct, error) {
	if depth > MaxRecursionDepth {
		return nil, fmt.Errorf("recursion depth limit reached (%d)", MaxRecursionDepth)
	}

	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)

		if field.IsMap() {
			// Handle map fields: convert to object with additionalProperties
			mapValueField := field.MapValue()
			valueSchema, err := fieldToSchema(mapValueField, depth)
			if err != nil {
				return nil, fmt.Errorf("failed to process map value for field %s: %w", field.Name(), err)
			}

			schema := map[string]interface{}{
				"type":                 TypeObject,
				"additionalProperties": valueSchema,
			}

			structValue, err := structpb.NewStruct(schema)
			if err != nil {
				return nil, fmt.Errorf("failed to create struct for field %s: %w", field.Name(), err)
			}
			properties.Fields[string(field.Name())] = structpb.NewStructValue(structValue)
			continue
		}

		schema, err := fieldToSchema(field, depth)
		if err != nil {
			return nil, err
		}

		if field.IsList() {
			itemSchema := make(map[string]interface{})
			for k, v := range schema {
				itemSchema[k] = v
			}
			schema = map[string]interface{}{
				"type":  TypeArray,
				"items": itemSchema,
			}
		}

		structValue, err := structpb.NewStruct(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct for field %s: %w", field.Name(), err)
		}
		properties.Fields[string(field.Name())] = structpb.NewStructValue(structValue)
	}

	return properties, nil
}

func fieldToSchema(field protoreflect.FieldDescriptor, depth int) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"type": TypeString, // Default
	}

	switch field.Kind() {
	case protoreflect.DoubleKind, protoreflect.FloatKind:
		schema["type"] = TypeNumber
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		schema["type"] = TypeInteger
	case protoreflect.BoolKind:
		schema["type"] = TypeBoolean
	case protoreflect.EnumKind:
		schema["type"] = TypeString
		enumVals := field.Enum().Values()
		var values []interface{}
		for i := 0; i < enumVals.Len(); i++ {
			values = append(values, string(enumVals.Get(i).Name()))
		}
		schema["enum"] = values
	case protoreflect.MessageKind, protoreflect.GroupKind:
		schema["type"] = TypeObject
		nestedProps, err := fieldsToProperties(field.Message().Fields(), depth+1)
		if err != nil {
			return nil, fmt.Errorf("failed to process nested message %s: %w", field.Name(), err)
		}
		schema["properties"] = nestedProps.AsMap()
	}
	return schema, nil
}

// ConfigParameter an interface for config parameter schemas.
type ConfigParameter interface {
	// GetSchema returns the parameter schema.
	//
	// Returns:
	//   - *configv1.ParameterSchema: The parameter schema.
	GetSchema() *configv1.ParameterSchema
}

// McpFieldParameter an interface for McpField parameter schemas.
type McpFieldParameter interface {
	// GetName returns the name of the parameter.
	//
	// Returns:
	//   - string: The name of the parameter.
	GetName() string
	// GetDescription returns the description of the parameter.
	//
	// Returns:
	//   - string: The description of the parameter.
	GetDescription() string
	// GetType returns the type of the parameter.
	//
	// Returns:
	//   - string: The type of the parameter.
	GetType() string
	// GetIsRepeated returns true if the parameter is a repeated field (array).
	//
	// Returns:
	//   - bool: True if the parameter is repeated.
	GetIsRepeated() bool
}

// ConfigSchemaToProtoProperties converts a slice of parameter schema definitions
// from a service configuration into a `structpb.Struct` that can be used as the
// `properties` field in a protobuf-based tool input schema.
//
// Parameters:
//   - params: A slice of parameters implementing ConfigParameter.
//
// Returns:
//   - *structpb.Struct: The properties structure for the input schema.
//   - []string: A list of required parameter names.
//   - error: An error if the conversion fails.
func ConfigSchemaToProtoProperties[T ConfigParameter](params []T) (*structpb.Struct, []string, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	var required []string

	for _, param := range params {
		paramSchema := param.GetSchema()
		if paramSchema == nil {
			continue
		}

		if paramSchema.GetIsRequired() {
			required = append(required, paramSchema.GetName())
		}

		typeStr := strings.ToLower(configv1.ParameterType_name[int32(paramSchema.GetType())])
		if typeStr == "" {
			typeStr = TypeString
		}
		paramStruct := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":        structpb.NewStringValue(typeStr),
				"description": structpb.NewStringValue(paramSchema.GetDescription()),
			},
		}

		if paramSchema.GetDefaultValue() != nil {
			paramStruct.Fields["default"] = paramSchema.GetDefaultValue()
		}

		properties.Fields[paramSchema.GetName()] = structpb.NewStructValue(paramStruct)
	}

	return properties, required, nil
}

// McpFieldsToProtoProperties converts a slice of McpField definitions into a
// `structpb.Struct` that can be used as the `properties` field in a
// protobuf-based tool input schema.
//
// Parameters:
//   - params: A slice of parameters implementing McpFieldParameter.
//
// Returns:
//   - *structpb.Struct: The properties structure for the input schema.
//   - error: An error if the conversion fails.
func McpFieldsToProtoProperties[T McpFieldParameter](params []T) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	for _, param := range params {
		scalarType := strings.ToLower(strings.TrimPrefix(param.GetType(), "TYPE_"))
		typeVal := TypeString // Default

		switch scalarType {
		case "double", "float":
			typeVal = TypeNumber
		case "int32", "int64", "sint32", "sint64", "uint32", "uint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
			typeVal = TypeInteger
		case "bool":
			typeVal = TypeBoolean
		}

		scalarSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":        structpb.NewStringValue(typeVal),
				"description": structpb.NewStringValue(param.GetDescription()),
			},
		}

		var finalSchema *structpb.Struct
		if param.GetIsRepeated() {
			finalSchema = &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type":  structpb.NewStringValue(TypeArray),
					"items": structpb.NewStructValue(scalarSchema),
				},
			}
		} else {
			finalSchema = scalarSchema
		}

		properties.Fields[param.GetName()] = structpb.NewStructValue(finalSchema)
	}

	return properties, nil
}
