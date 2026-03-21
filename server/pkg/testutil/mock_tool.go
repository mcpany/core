// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
)

// MockTool is a mock implementation of tool.Tool
type MockTool struct {
	Name        string
	Description string
	ToolSchema  *configv1.ToolSchema
	InputSchema *configv1.JSONSchema
	ExecuteFunc func(ctx context.Context, arguments map[string]interface{}) (interface{}, error)
}

func (m *MockTool) GetName() string {
	return m.Name
}

func (m *MockTool) GetDescription() string {
	return m.Description
}

func (m *MockTool) GetInputSchema() *configv1.JSONSchema {
	return m.InputSchema
}

func (m *MockTool) Execute(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, arguments)
	}

	// Default fallback: return a complex nested array structure to test the RichResultViewer's flattening logic
	return []interface{}{
		map[string]interface{}{
			"user": map[string]interface{}{
				"profile": map[string]interface{}{
					"name": "Alice Liddell",
					"age":  28,
				},
				"role": "admin",
			},
			"metadata": map[string]interface{}{
				"preferences": map[string]interface{}{
					"theme": "dark",
				},
			},
			"contacts": []interface{}{
				map[string]interface{}{"type": "email", "value": "alice@example.com"},
			},
		},
		map[string]interface{}{
			"user": map[string]interface{}{
				"profile": map[string]interface{}{
					"name": "Bob Builder",
					"age":  35,
				},
				"role": "user",
			},
			"metadata": map[string]interface{}{
				"preferences": map[string]interface{}{
					"theme": "light",
				},
			},
			"contacts": []interface{}{
				map[string]interface{}{"type": "phone", "value": "555-0102"},
			},
		},
	}, nil
}

func (m *MockTool) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":        m.Name,
		"description": m.Description,
	})
}
