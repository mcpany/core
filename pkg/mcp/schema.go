// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import "fmt"

// Schema is a struct that represents a JSON schema.
type Schema struct {
	Type       string             `json:"type"`
	Properties map[string]*Schema `json:"properties"`
	Required   []string           `json:"required"`
	Items      *Schema            `json:"items"`
}

// Validate validates the schema.
func (s *Schema) Validate() error {
	if s.Type == "" {
		return fmt.Errorf("type is required")
	}

	validTypes := []string{"object", "array", "string", "number", "integer", "boolean"}
	isValidType := false
	for _, validType := range validTypes {
		if s.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid type: %s", s.Type)
	}

	if s.Type == "object" {
		for _, prop := range s.Properties {
			if err := prop.Validate(); err != nil {
				return err
			}
		}
	}

	if s.Type == "array" {
		if s.Items == nil {
			return fmt.Errorf("items is required for array type")
		}
		if err := s.Items.Validate(); err != nil {
			return err
		}
	}

	return nil
}
