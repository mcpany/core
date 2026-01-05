// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToolName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		toolName      string
		namespace     string
		tool          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid tool name",
			toolName:    "my-service.my-tool",
			namespace:   "my-service",
			tool:        "my-tool",
			expectError: false,
		},
		{
			name:        "valid tool name with dashes",
			toolName:    "my-service.--my-tool",
			namespace:   "my-service",
			tool:        "my-tool",
			expectError: false,
		},
		{
			name:        "tool name without namespace",
			toolName:    "my-tool",
			namespace:   "",
			tool:        "my-tool",
			expectError: false,
		},
		{
			name:          "empty tool name",
			toolName:      "",
			expectError:   true,
			errorContains: "invalid tool name",
		},
		{
			name:          "tool name with only separator",
			toolName:      ".",
			expectError:   true,
			errorContains: "invalid tool name",
		},
		{
			name:          "tool name with only separator and dashes",
			toolName:      ".--",
			expectError:   true,
			errorContains: "invalid tool name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namespace, tool, err := ParseToolName(tc.toolName)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.namespace, namespace)
				assert.Equal(t, tc.tool, tool)
			}
		})
	}
}

func TestGetFullyQualifiedToolName(t *testing.T) {
	t.Parallel()
	serviceID := "test-service"
	methodName := "test-method"
	expected := "test-service.test-method"
	actual := GetFullyQualifiedToolName(serviceID, methodName)
	assert.Equal(t, expected, actual)
}
