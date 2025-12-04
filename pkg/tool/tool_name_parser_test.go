
/*
 * Copyright 2025 Author(s) of MCP Any
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

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToolName(t *testing.T) {
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
	serviceID := "test-service"
	methodName := "test-method"
	expected := "test-service.test-method"
	actual := GetFullyQualifiedToolName(serviceID, methodName)
	assert.Equal(t, expected, actual)
}
