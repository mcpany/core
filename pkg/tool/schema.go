// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"encoding/json"

	"github.com/mcpany/core/pkg/mcp"
	"google.golang.org/protobuf/types/known/structpb"
)

// NewSchemaFromProto creates a new Schema from a structpb.Struct.
func NewSchemaFromProto(s *structpb.Struct) (*mcp.Schema, error) {
	if s == nil {
		return nil, nil
	}

	// Convert the structpb.Struct to a JSON string.
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON string into a Schema.
	var schema mcp.Schema
	if err := json.Unmarshal(jsonBytes, &schema); err != nil {
		return nil, err
	}

	if err := schema.Validate(); err != nil {
		return nil, err
	}

	return &schema, nil
}
