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

// JSON schema types
const (
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeBoolean = "boolean"
	TypeObject  = "object"
)

// MethodDescriptorToProtoProperties converts the fields of a method's input
// message into a `structpb.Struct` for use as the `properties` field in a tool
// input schema.
func MethodDescriptorToProtoProperties(methodDesc protoreflect.MethodDescriptor) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	inputFields := methodDesc.Input().Fields()

	for i := 0; i < inputFields.Len(); i++ {
		field := inputFields.Get(i)
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
		case protoreflect.MessageKind:
			schema["type"] = TypeObject
		}

		structValue, err := structpb.NewStruct(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct for field %s: %w", field.Name(), err)
		}
		properties.Fields[string(field.Name())] = structpb.NewStructValue(structValue)
	}

	return properties, nil
}

// MethodOutputDescriptorToProtoProperties converts the fields of a method's
// output message into a `structpb.Struct` for use as the `properties` field in
// a tool output schema.
func MethodOutputDescriptorToProtoProperties(methodDesc protoreflect.MethodDescriptor) (*structpb.Struct, error) {
	properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	outputFields := methodDesc.Output().Fields()

	for i := 0; i < outputFields.Len(); i++ {
		field := outputFields.Get(i)
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
		case protoreflect.MessageKind:
			schema["type"] = TypeObject
		}

		structValue, err := structpb.NewStruct(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct for field %s: %w", field.Name(), err)
		}
		properties.Fields[string(field.Name())] = structpb.NewStructValue(structValue)
	}

	return properties, nil
}

// ConfigParameter an interface for config parameter schemas
type ConfigParameter interface {
	// GetSchema returns the parameter schema.
	GetSchema() *configv1.ParameterSchema
}

// McpFieldParameter an interface for McpField parameter schemas
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
	properties, err := structpb.NewStruct(make(map[string]interface{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create properties struct: %w", err)
	}

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
		properties.Fields[paramSchema.GetName()] = structpb.NewStructValue(paramStruct)
	}

	return properties, nil
}

// McpFieldsToProtoProperties converts a slice of McpField definitions into a
// `structpb.Struct` that can be used as the `properties` field in a
// protobuf-based tool input schema.
func McpFieldsToProtoProperties[T McpFieldParameter](params []T) (*structpb.Struct, error) {
	properties, err := structpb.NewStruct(make(map[string]interface{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create properties struct: %w", err)
	}

	for _, param := range params {
		schema := map[string]interface{}{
			"type":        "string", // Default to string for simplicity
			"description": param.GetDescription(),
		}

		scalarType := strings.ToLower(strings.TrimPrefix(param.GetType(), "TYPE_"))
		switch scalarType {
		case "double", "float":
			schema["type"] = TypeNumber
		case "int32", "int64", "sint32", "sint64", "uint32", "uint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
			schema["type"] = TypeInteger
		case "bool":
			schema["type"] = TypeBoolean
		}

		structValue, err := structpb.NewStruct(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct from schema: %w", err)
		}
		properties.Fields[param.GetName()] = structpb.NewStructValue(structValue)
	}

	return properties, nil
}
