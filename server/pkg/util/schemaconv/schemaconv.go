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
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeBoolean = "boolean"
	TypeObject  = "object"
	TypeArray   = "array"
)

// MaxRecursionDepth limits the depth of nested messages to prevent infinite recursion.
const MaxRecursionDepth = 10

// MethodDescriptorToProtoProperties converts the fields of a method's input
// message into a `structpb.Struct` for use as the `properties` field in a tool
// input schema.
func MethodDescriptorToProtoProperties(methodDesc protoreflect.MethodDescriptor) (*structpb.Struct, error) {
	return fieldsToProperties(methodDesc.Input().Fields(), 0)
}

// MethodOutputDescriptorToProtoProperties converts the fields of a method's
// output message into a `structpb.Struct` for use as the `properties` field in
// a tool output schema.
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
		"type": "string", // Default
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
	GetSchema() *configv1.ParameterSchema
}

// McpFieldParameter an interface for McpField parameter schemas.
type McpFieldParameter interface {
	// GetName returns the name of the parameter.
	GetName() string
	// GetDescription returns the description of the parameter.
	GetDescription() string
	// GetType returns the type of the parameter.
	GetType() string
}

// ConfigSchemaToProtoProperties converts a slice of parameter schema definitions
// from a service configuration into a `structpb.Struct` that can be used as the
// `properties` field in a protobuf-based tool input schema.
func ConfigSchemaToProtoProperties[T ConfigParameter](params []T) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	for _, param := range params {
		paramSchema := param.GetSchema()
		if paramSchema == nil {
			continue
		}

		paramStruct := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":        structpb.NewStringValue(strings.ToLower(configv1.ParameterType_name[int32(paramSchema.GetType())])),
				"description": structpb.NewStringValue(paramSchema.GetDescription()),
			},
		}

		if paramSchema.GetDefaultValue() != nil {
			paramStruct.Fields["default"] = paramSchema.GetDefaultValue()
		}

		properties.Fields[paramSchema.GetName()] = structpb.NewStructValue(paramStruct)
	}

	return properties, nil
}

// McpFieldsToProtoProperties converts a slice of McpField definitions into a
// `structpb.Struct` that can be used as the `properties` field in a
// protobuf-based tool input schema.
func McpFieldsToProtoProperties[T McpFieldParameter](params []T) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	for _, param := range params {
		scalarType := strings.ToLower(strings.TrimPrefix(param.GetType(), "TYPE_"))
		typeVal := "string" // Default

		switch scalarType {
		case "double", "float":
			typeVal = TypeNumber
		case "int32", "int64", "sint32", "sint64", "uint32", "uint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
			typeVal = TypeInteger
		case "bool":
			typeVal = TypeBoolean
		}

		structValue := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":        structpb.NewStringValue(typeVal),
				"description": structpb.NewStringValue(param.GetDescription()),
			},
		}
		properties.Fields[param.GetName()] = structpb.NewStructValue(structValue)
	}

	return properties, nil
}
