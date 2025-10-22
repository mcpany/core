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

package schemaconv

import (
	"fmt"
	"strings"

	configv1 "github.com/mcpxy/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// ConfigParameter an interface for config parameter schemas
type ConfigParameter interface {
	GetSchema() *configv1.ParameterSchema
}

// McpFieldParameter an interface for McpField parameter schemas
type McpFieldParameter interface {
	GetName() string
	GetDescription() string
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
			schema["type"] = "number"
		case "int32", "int64", "sint32", "sint64", "uint32", "uint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
			schema["type"] = "integer"
		case "bool":
			schema["type"] = "boolean"
		}

		structValue, err := structpb.NewStruct(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct from schema: %w", err)
		}
		properties.Fields[param.GetName()] = structpb.NewStructValue(structValue)
	}

	return properties, nil
}
